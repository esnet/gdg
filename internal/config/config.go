package config

import (
	"github.com/spf13/viper"
)

type Configuration struct {
	gdgViperConfig   *viper.Viper
	gdgConfig        *GDGAppConfiguration
	templatingConfig *TemplatingConfig
}

type Provider func() *Configuration
