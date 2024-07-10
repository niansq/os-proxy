package main

import (
	"fmt"
	"os-proxy/bootstrap"
	"os-proxy/bootstrap/plugins"
)

func main() {
	lgConfig := bootstrap.NewConfig("conf/config.yaml")
	fmt.Println(lgConfig.Redis)
	fmt.Println(lgConfig.App)
	fmt.Println(lgConfig.Database[0])
	lgLogger := bootstrap.NewLogger()
	lgLogger.Logger.Error("server_down")
	plugins.NewPlugins()
	defer plugins.ClosePlugins()
}
