package plugins

import (
	"fmt"
	"os-proxy/bootstrap"
)

/*
	项目中会用到各种插件（redis,mysql等），我在plugins中对这些插件进行注册和关闭，提供给项目后续使用
*/

//定义一个plugin接口，使得插件满足一定规范

type Plugin interface {
	Flag() bool       //是否启动,即配置文件中是否使用该插件
	Name() string     //插件名称
	New() interface{} //插件初始化，返回一个该插件示例
	Health()          //插件健康检查
	Close()           //关闭插件
}

// 记录已经注册的插件
var Plugins = make(map[string]Plugin)

// 对插件进行注册,各个插件会调用这个函数
func RegisterdPlugin(plugin Plugin) {
	Plugins[plugin.Name()] = plugin
}

// 根据注册情况，打印各个插件的状态信息，插件初始化生成一个可以使用的实例
func NewPlugins() {
	for _, p := range Plugins {
		if !p.Flag() { //检查配置文件中是否开启了这个插件
			continue
		}
		bootstrap.NewLogger().Logger.Info(fmt.Sprintf("%s Init ...", p.Name()))
		p.New() //创建实例，方便后续使用
		bootstrap.NewLogger().Logger.Info(fmt.Sprintf("%s HealthCheck ...", p.Name()))
		p.Health()
		bootstrap.NewLogger().Logger.Info(fmt.Sprintf("%s Success Init", p.Name()))

	}
}

func ClosePlugins() {
	for _, p := range Plugins {
		if !p.Flag() {
			continue
		}
		p.Close()
		bootstrap.NewLogger().Logger.Info(fmt.Sprintf("%s Success Close", p.Name()))
	}
}
