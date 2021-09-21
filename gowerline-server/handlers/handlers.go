package handlers

import (
	"context"

	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func SetupHandlers(router *gin.Engine, ctx context.Context, log *zap.Logger, plugins map[string]*plugins.Plugin) error {
	router.GET("/ping", PingHandler)
	router.POST("/plugin", BuildPluginHandler(ctx, log, plugins))
	router.GET("/version", versionHandler)
	return nil
}
