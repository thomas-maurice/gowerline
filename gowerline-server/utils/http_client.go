package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/thomas-maurice/gowerline/gowerline-server/config"
	"go.uber.org/zap"
)

func NewHTTPClientFromConfig(cfg *config.Config) *http.Client {
	currentUser, err := user.Current()
	if err != nil {
		log.Panic("could not get current user", zap.Error(err))
	}
	homeDir := currentUser.HomeDir
	listenPath := cfg.Listen.Unix

	if strings.HasPrefix(cfg.Listen.Unix, "~/") {
		listenPath = filepath.Join(homeDir, cfg.Listen.Unix[2:])
	}

	if cfg.Listen.Unix != "" {
		return &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", listenPath)
				},
			},
		}
	}

	return &http.Client{}
}

func BaseURLFromConfig(cfg *config.Config) string {
	return fmt.Sprintf("http://localhost:%d", cfg.Listen.Port)
}
