package config

// 用于接收yaml中app中对应的数据
type App struct {
	Env     string `mapstructure:"env" json:"env" yaml:"env"`
	Port    string `mapstructure:"port" json:"port" yaml:"port"`
	AppName string `mapstructure:"app_name" json:"app_name" yaml:"app_name"`
	APPUrl  string `mapstructure:"app_url" json:"app_url" yaml:"app_url"`
}
