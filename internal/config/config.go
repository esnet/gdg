package config

import (
	"github.com/esnet/gdg/internal/config/domain"
	"github.com/spf13/viper"
)

type Configuration struct {
	gdgViperConfig   *viper.Viper
	gdgConfig        *domain.GDGAppConfiguration
	templatingConfig *domain.TemplatingConfig
}

type Provider func() *Configuration
