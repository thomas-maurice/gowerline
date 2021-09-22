package plugins

import (
	"context"
	"plugin"

	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Called when a plugin starts, returns data such as the plugin name
type PluginStartFunc func(context.Context, *zap.Logger) (*types.PluginStartData, error)

// Called when a plugin stops, must wait before effectively stopping
type PluginStopFunc func(context.Context, *zap.Logger) error

// Called when we need to render a segment effectively
type PluginCallFunc func(context.Context, *zap.Logger, *types.Payload) ([]*types.PowerlineReturn, error)

// Plugin type
type Plugin struct {
	Start PluginStartFunc
	Stop  PluginStopFunc
	Call  PluginCallFunc

	Name     string
	Metadata types.PluginMetadata
}

// PluginConfig will be passed down to plugins
type PluginConfig struct {
	UserHome     string
	GowerlineDir string
	PluginName   string
	// Config is a yaml node containing the configuration that
	// is specific to the plugin
	Config yaml.Node
}

func NewPlugin(ctx context.Context, log *zap.Logger, filePath string, pluginConfig *PluginConfig) (*Plugin, error) {
	p, err := plugin.Open(filePath)
	if err != nil {
		return nil, err
	}

	sym, err := p.Lookup("Init")
	if err != nil {
		return nil, err
	}

	init := sym.(func(context.Context, *zap.Logger, *PluginConfig) (*Plugin, error))

	return init(context.Background(), log.With(zap.String("plugin_path", filePath)), pluginConfig)
}

func (p *Plugin) RunStart(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	return p.Start(ctx, log.With(zap.String("plugin_name", p.Name)))
}

func (p *Plugin) RunStop(ctx context.Context, log *zap.Logger) error {
	if p.Stop != nil {
		return p.Stop(ctx, log.With(zap.String("plugin_name", p.Name)))
	}
	return nil
}

func (p *Plugin) RunCall(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	return p.Call(ctx, log.With(zap.String("plugin_name", p.Name)), payload)
}
