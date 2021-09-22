package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"gopkg.in/fsnotify.v1"

	"go.uber.org/zap"
)

const defaultTemplate = `{{ .RenderedExpiry }}`

type VaultState struct {
	Accessor       string
	CreationTime   int64
	DisplayName    string
	EntityID       string
	RenderedExpiry string
	CreationTTL    int64
	ExpiryTime     int64
}

var (
	defaultHighlightSegments = []string{
		"gwl:vault",
		"information:regular",
	}
	expiredHighlightSegments = []string{
		"gwl:vault_expired",
		"information:regular",
	}
	pluginConfig   *plugins.PluginConfig
	pluginName     = "vault"
	stopChannel    chan bool
	stoppedChannel chan bool
	vaultState     *VaultState
)

type pluginArgs struct {
	Template     string `json:"template"`
	ExpiredTheme bool   `json:"expired_theme"` // changes the colour is the token is expired
}

func (vs *VaultState) Expired() bool {
	return time.Now().Unix() >= (vs.ExpiryTime)
}

func (vs *VaultState) Expire() {
	vs.Accessor = ""
	vs.CreationTime = 0
	vs.CreationTTL = 0
	vs.DisplayName = "not logged"
	vs.EntityID = ""
	vs.RenderedExpiry = ""
	vs.ExpiryTime = 0
}

func (vs *VaultState) ExpiresIn() int64 {
	return (vs.ExpiryTime) - time.Now().Unix()
}

func (vs *VaultState) ExpiresString() string {
	if vs.Expired() {
		return "expired"
	}

	return time.Until(time.Unix(vs.ExpiryTime, 0)).Truncate(time.Second).String()
}

func (vs *VaultState) Render() {
	vs.RenderedExpiry = vs.ExpiresString()
}

// updateVaultInfos gets the data for caching
func updateVaultInfos(log *zap.Logger) error {
	log.Info("updating vault data")
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	if client.Token() == "" {
		tknBytes, err := ioutil.ReadFile(path.Join(pluginConfig.UserHome, ".vault-token"))
		if err != nil {
			return err
		}

		client.SetToken(string(tknBytes))
	}

	result, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}

	var vs VaultState
	var ok bool
	vs.Accessor, ok = result.Data["accessor"].(string)
	if !ok {
		return fmt.Errorf("could not extract accessor")
	}

	vs.DisplayName, ok = result.Data["display_name"].(string)
	if !ok {
		return fmt.Errorf("could not extract display_name")
	}

	vs.EntityID, ok = result.Data["entity_id"].(string)
	if !ok {
		return fmt.Errorf("could not extract entity_id")
	}

	num, ok := result.Data["creation_time"].(json.Number)
	if !ok {
		return fmt.Errorf("could not extract creation_time")
	}
	vs.CreationTime, err = num.Int64()
	if err != nil {
		return fmt.Errorf("could not parse creation_time")
	}

	num, ok = result.Data["creation_ttl"].(json.Number)
	if !ok {
		return fmt.Errorf("could not extract creation_ttl")
	}
	vs.CreationTTL, err = num.Int64()
	if err != nil {
		return fmt.Errorf("could not parse creation_ttl")
	}
	vs.ExpiryTime = vs.CreationTime + vs.CreationTTL

	vaultState = &vs
	return nil
}

func run(log *zap.Logger) {
	err := updateVaultInfos(log)
	if err != nil {
		vaultState.Expire()
		log.Error("failed to update vault informations", zap.Error(err))
	}

	tck := time.NewTicker(time.Minute)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("could not setup fsnotify watcher", zap.Error(err))
	}
	defer watcher.Close()

	err = watcher.Add(path.Join(pluginConfig.UserHome, ".vault-token"))
	if err != nil {
		log.Error("could not watch ~/.vault-token", zap.Error(err))
	}

	var lastUpdate time.Time

	for {
		select {
		case <-tck.C:
			err := updateVaultInfos(log)
			if err != nil {
				vaultState.Expire()
				log.Error("failed to update vault informations", zap.Error(err))
			}
		case _, ok := <-watcher.Events:
			if !ok {
				break
			}

			if time.Since(lastUpdate) > time.Second*5 {
				log.Info("reload triggered by a change of the ~/.vault-token file")
				err := updateVaultInfos(log)
				if err != nil {
					vaultState.Expire()
					log.Error("failed to update vault informations", zap.Error(err))
				}
				lastUpdate = time.Now()
			}
		case <-stopChannel:
			stoppedChannel <- true
			return
		}
	}
}

// Starts the plugin, here you might want to do all the initialisation you need
// load up config/tokens and what not, as well to start long running goroutines
// if your plugin requires it
func Start(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	stopChannel = make(chan bool)
	stoppedChannel = make(chan bool)
	vaultState = &VaultState{}

	log.Info(
		"started plugin",
		zap.String("plugin", pluginName),
	)

	go run(log)

	return &types.PluginStartData{
		Metadata: types.PluginMetadata{
			Description: "Gathers information about the current Vault token and formats the result",
			Author:      "Thomas Maurice <thomas@maurice.fr>",
			Version:     "0.0.1",
			Functions: []types.FunctionDescriptor{
				{
					Name:        "vault",
					Description: "Displays informations about Vault using a formatting string",
					Parameters: map[string]string{
						"template": "Template string to render",
					},
				},
			},
		},
	}, nil
}

// Stops anything you have started that is long runinng, like goroutines and what not
func Stop(ctx context.Context, log *zap.Logger) error {
	log.Info(
		"stopped plugin",
		zap.String("plugin", pluginName),
	)
	return nil
}

// Returns the actual segment iself. If your plugin handles different functions you should
// check what is called using the payload.Function attribute
func Call(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	var args pluginArgs
	if payload.Args != nil {
		err := json.Unmarshal(*payload.Args, &args)
		if err != nil {
			log.Error("could not unmarshal arguments", zap.Error(err))
			return nil, err
		}
	}

	if args.Template == "" {
		args.Template = defaultTemplate
	}

	vaultState.Render()
	t, err := template.New("segment").Parse(args.Template)
	if err != nil {
		return nil, err
	}
	wr := bytes.NewBufferString("")
	err = t.Execute(wr, vaultState)
	if err != nil {
		return nil, err
	}

	hlg := defaultHighlightSegments
	if args.ExpiredTheme && vaultState.Expired() {
		hlg = expiredHighlightSegments
	}

	return []*types.PowerlineReturn{
		{
			Content:        wr.String(),
			HighlightGroup: hlg,
		},
	}, nil
}

// Init builds and returns the plugin itself
func Init(ctx context.Context, log *zap.Logger, pCfg *plugins.PluginConfig) (*plugins.Plugin, error) {
	log.Info(
		"loaded plugin",
		zap.String("plugin", pluginName),
	)

	pluginConfig = pCfg

	return &plugins.Plugin{
		Start: Start,
		Stop:  Stop,
		Call:  Call,
		Name:  pluginName,
	}, nil
}
