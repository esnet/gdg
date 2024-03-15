package config

import "github.com/spf13/viper"

type Configuration struct {
	viperConfiguration map[string]*viper.Viper
	gdgConfig          *GDGAppConfiguration
	templatingConfig   *TemplatingConfig
}
