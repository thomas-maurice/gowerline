package plugins

import (
	"context"
	"plugin"

	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
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

	Name string
}

func NewPlugin(ctx context.Context, log *zap.Logger, filePath string) (*Plugin, error) {
	p, err := plugin.Open(filePath)
	if err != nil {
		return nil, err
	}

	sym, err := p.Lookup("Init")
	if err != nil {
		return nil, err
	}

	init := sym.(func(context.Context, *zap.Logger) (*Plugin, error))

	return init(context.Background(), log.With(zap.String("plugin", filePath)))
}

func (p *Plugin) RunStart(ctx context.Context, log *zap.Logger) (*types.PluginStartData, error) {
	return p.Start(ctx, log.With(zap.String("plugin", p.Name)))
}

func (p *Plugin) RunStop(ctx context.Context, log *zap.Logger) error {
	if p.Stop != nil {
		return p.Stop(ctx, log.With(zap.String("plugin", p.Name)))
	}
	return nil
}

func (p *Plugin) RunCall(ctx context.Context, log *zap.Logger, payload *types.Payload) ([]*types.PowerlineReturn, error) {
	return p.Call(ctx, log.With(zap.String("plugin", p.Name)), payload)
}
