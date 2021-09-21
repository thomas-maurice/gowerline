package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thomas-maurice/gowerline/gowerline-server/types"
	"github.com/thomas-maurice/gowerline/gowerline-server/version"
)

func versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, types.ServerVersionInfo{
		Version:         version.Version,
		BuildDate:       version.BuildTime,
		BuildHost:       version.BuildHost,
		GitHash:         version.BuildHash,
		Architecture:    version.Arch,
		OperatingSystem: version.OS,
	})
}
