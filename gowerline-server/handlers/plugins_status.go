package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
)

func BuildPluginStatusHandler(ctx context.Context, log *zap.Logger, pluginMap map[string]*plugins.Plugin) func(c *gin.Context) {
	return func(c *gin.Context) {
		result := make([]types.PluginStatus, 0)
		for plgName, plg := range pluginMap {
			status := types.PluginStatus{
				Name:        plgName,
				Description: "",
				Functions:   plg.Functions,
			}
			result = append(result, status)
		}
		c.JSON(http.StatusOK, result)
	}
}
