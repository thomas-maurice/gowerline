package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-maurice/gowerline/gowerline-server/plugins"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"go.uber.org/zap"
)

func PluginHandler(c *gin.Context) {}

func BuildPluginHandler(ctx context.Context, log *zap.Logger, pluginMap map[string]*plugins.Plugin) func(c *gin.Context) {
	return func(c *gin.Context) {
		var payload types.Payload

		requestBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Error(
				"could not read request",
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, types.PowerlineReturn{
				Content: fmt.Sprintf("err: %s", err),
			})
			return
		}

		err = json.Unmarshal(requestBytes, &payload)
		if err != nil {
			log.Error(
				"could not unmarshal request",
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, types.PowerlineReturn{
				Content: fmt.Sprintf("err: %s", err),
			})
			return
		}

		plg, ok := pluginMap[payload.Function]
		if !ok {
			c.JSON(http.StatusNotFound, types.PowerlineReturn{
				Content: fmt.Sprintf("no such plugin %s", payload.Function),
			})
			return
		}

		result, err := plg.RunCall(
			context.Background(),
			log.With(zap.String("function", payload.Function)),
			&payload)
		if err != nil {
			log.Error(
				"could not unmarshal request",
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, types.PowerlineReturn{
				Content: fmt.Sprintf("err:%s %s", payload.Function, err),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
