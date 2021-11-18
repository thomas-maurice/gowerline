//nolint:unused
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

const (
	defaultPublicIpService = "https://checkip.amazonaws.com/"
)

var (
	cfg                      Config
	stopChannel              chan bool
	stoppedChannel           chan bool
	pluginConfig             *plugins.PluginConfig
	publicIpAddress          string
	interfacesAddresses      map[string]string
	interfacesAddressesMutex *sync.Mutex
)

type Config struct {
	IpService string `json:"ipService" yaml:"ipService"`
}

type pluginArgs struct {
	Interface string `json:"interface"` // name of the interface to which get the address
}

func getDefaultIPAddress(log *zap.Logger) (string, error) {
	handle, err := netlink.NewHandle(netlink.FAMILY_V4)
	if err != nil {
		return "", err
	}
	routes, err := handle.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return "", err
	}

	defer handle.Close()

	for _, route := range routes {
		if route.Dst == nil {
			ifLink, err := netlink.LinkByIndex(route.LinkIndex)
			if err != nil {
				return "", err
			}
			addresses, err := netlink.AddrList(ifLink, netlink.FAMILY_V4)
			if err != nil {
				return "", err
			}

			if len(addresses) == 0 {
				return "", fmt.Errorf("could not determine any address on %s", ifLink.Attrs().Name)
			}

			return addresses[0].IP.String(), nil
		}
	}

	return "", nil
}

func updateIPAddresses(log *zap.Logger) error {
	ifaces, err := net.Interfaces()

	newInterfacesAddress := make(map[string]string)

	if err != nil {
		log.Error("could not list interfaces", zap.Error(err))
		return err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Error("could not get addresses for interface", zap.String("interface", iface.Name))
			return err
		}

		if len(addrs) > 0 {
			// we only consider the first one
			newInterfacesAddress[iface.Name] = addrs[0].String()
		}
	}

	defaultAddress, err := getDefaultIPAddress(log)
	if err != nil {
		log.Error("could not determine default IP address", zap.Error(err))
	}
	newInterfacesAddress["default"] = defaultAddress

	interfacesAddressesMutex.Lock()
	defer interfacesAddressesMutex.Unlock()

	interfacesAddresses = newInterfacesAddress

	return nil
}

func update(log *zap.Logger) error {
	log.Info("running the update loop")

	err := updateIPAddresses(log)
	if err != nil {
		log.Error("could not update the status of ip addresses", zap.Error(err))
	}

	if cfg.IpService == "" {
		cfg.IpService = defaultPublicIpService
	}

	resp, err := http.Get(cfg.IpService)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	publicIpAddress = strings.ReplaceAll(string(b), "\n", "")

	return nil
}

func run(log *zap.Logger) {
	err := update(log)
	if err != nil {
		log.Error("failed to run plugin data refresh", zap.Error(err))
	}

	tck := time.NewTicker(time.Minute)

	for {
		select {
		case <-tck.C:
			err = update(log)
			if err != nil {
				log.Error("could not update data", zap.Error(err))
			}
		case <-stopChannel:
			stoppedChannel <- true
			return
		}
	}
}

func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	stopChannel = make(chan bool)
	stoppedChannel = make(chan bool)
	interfacesAddresses = make(map[string]string)
	interfacesAddressesMutex = &sync.Mutex{}

	err := pluginConfig.Config.Decode(&cfg)
	if err != nil {
		log.Panic("could not load configuration", zap.Error(err))
	}

	go run(log)

	return &types.PluginStartData{
		Metadata: types.PluginMetadata{
			Description: "Gather information about your network connectivity",
			Author:      "Thomas Maurice <thomas@maurice.fr>",
			Version:     "devel",
			Functions: []types.FunctionDescriptor{
				{
					Name:        "public_ip",
					Description: "Returns your public IP address",
					Parameters:  map[string]string{},
				},
				{
					Name:        "interface_ip",
					Description: "Returns the IP of an interface",
					Parameters: map[string]string{
						"interface": "The interface in question",
					},
				},
				{
					Name:        "hostname",
					Description: "Returns the hostname of the host",
					Parameters:  map[string]string{},
				},
			},
		},
	}, nil
}

func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopped plugin",
	)

	stopChannel <- true
	<-stoppedChannel

	return nil
}

func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	var args pluginArgs
	err := json.Unmarshal(*payload.Args, &args)
	if err != nil {
		log.Error("could not unmarshal plugin arguments", zap.Error(err))
		return nil, err
	}

	switch payload.Function {
	case "public_ip":
		return []*types.PowerlineReturn{
			{
				Content: publicIpAddress,
				HighlightGroup: []string{
					"gwl:public_ip",
				},
			},
		}, nil
	case "interface_ip":
		interfacesAddressesMutex.Lock()
		defer interfacesAddressesMutex.Unlock()
		return []*types.PowerlineReturn{
			{
				Content: interfacesAddresses[args.Interface],
				HighlightGroup: []string{
					"gwl:interface_ip",
				},
			},
		}, nil
	case "hostname":
		hostname, err := os.Hostname()
		if err != nil {
			log.Error("could not get hostname", zap.Error(err))
		}
		return []*types.PowerlineReturn{
			{
				Content: hostname,
				HighlightGroup: []string{
					"gwl:hostname",
				},
			},
		}, err
	default:
		return []*types.PowerlineReturn{
			{
				Content: "no such function",
				HighlightGroup: []string{
					"information:regular",
				},
			},
		}, nil
	}
}

func Init(ctx context.Context, log *zap.Logger, pCfg *plugins.PluginConfig) (*plugins.Plugin, error) { //nolint:deadcode
	log.Info(
		"loaded plugin",
	)

	pluginConfig = pCfg

	return &plugins.Plugin{
		Start: Start,
		Stop:  Stop,
		Call:  Call,
		Name:  pCfg.PluginName,
	}, nil
}

// noop main function
func main() {}
