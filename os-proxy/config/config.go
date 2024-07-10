package config

import "os-proxy/config/plugins"

//集成所有yaml文件中对应的结构体，形成一个大结构体，对应yaml文件中所有参数

type Configuration struct {
	App      App                 `mapstructure:"app" json:"app" yaml:"app"`
	Log      Log                 `mapstructure:"log" json:"log" yaml:"log"`
	Database []*plugins.Database `mapstructure:"database" json:"database" yaml:"database"`
	Redis    *plugins.Redis      `mapstructure:"redis" json:"redis" yaml:"redis"`
	Local    *plugins.Local      `mapstructure:"local" json:"local" yaml:"local"`
}
