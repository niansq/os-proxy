package bootstrap

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os-proxy/config"
	"path/filepath"
	"sync"
)

var (
	configPath   string //
	rootPath     = ""
	lgConfig     = new(LangGoConfig)
	confFilePath = "conf/config.yaml" //默认的yaml文件地址
)

type LangGoConfig struct {
	Conf *config.Configuration //配置文件对应的结构体
	Once *sync.Once
	/*
		sync.Once 是 Go 语言中的一种同步原语，用于确保某个操作或函数在并发环境下只被执行一次。
		它只有一个导出的方法，即 Do，该方法接收一个函数参数。
		在 Do 方法被调用后，该函数将被执行，而且只会执行一次，即使在多个协程同时调用的情况下也是如此。
	*/
}

func newLangGoConfig() *LangGoConfig {
	return &LangGoConfig{
		Conf: &config.Configuration{},
		Once: &sync.Once{},
	}
}

//初始化配置对象lgConfig

func NewConfig(confFile string) *config.Configuration {
	if lgConfig.Conf != nil { //配置文件非空，直接返回，这里是非首次调用，即以及完成初始化。后面需要使用config调用这个函数直接返回
		return lgConfig.Conf
	} else { //首次调用，完成初始化
		lgConfig = newLangGoConfig() //为配置变量分配空间
		if confFile == "" {          //使用默认配置文件地址
			lgConfig.initLangGoConfig(confFilePath)
		} else { //使用传入配置文件地址
			lgConfig.initLangGoConfig(confFile)
		}
	}
	return lgConfig.Conf
}

// 根据yaml文件给LangGoConfig.Conf赋值
func (lg *LangGoConfig) initLangGoConfig(confFile string) {
	lg.Once.Do(
		func() {
			initConfig(lg.Conf, confFile)
		},
	)
}

func initConfig(conf *config.Configuration, confFile string) {
	pflag.StringVarP(&configPath, "conf", "", filepath.Join(rootPath, confFile),
		"config path, eg: --conf config.yaml")
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(rootPath, configPath)
	}

	//lgLogger.Logger.Info("load config:" + configPath)
	fmt.Println("load config:" + configPath)

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		//lgLogger.Logger.Error("read config failed: ", zap.String("err", err.Error()))
		fmt.Println("read config failed: ", zap.String("err", err.Error()))
		panic(err)
	}

	if err := v.Unmarshal(&conf); err != nil {
		//lgLogger.Logger.Error("config parse failed: ", zap.String("err", err.Error()))
		fmt.Println("config parse failed: ", zap.String("err", err.Error()))
	}

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		//lgLogger.Logger.Info("", zap.String("config file changed:", in.Name))
		fmt.Println("", zap.String("config file changed:", in.Name))
		defer func() {
			if err := recover(); err != nil {
				//lgLogger.Logger.Error("config file changed err:", zap.Any("err", err))
				fmt.Println("config file changed err:", zap.Any("err", err))
			}
		}()
		if err := v.Unmarshal(&conf); err != nil {
			//lgLogger.Logger.Error("config parse failed: ", zap.String("err", err.Error()))
			fmt.Println("config parse failed: ", zap.String("err", err.Error()))
		}
	})
	lgConfig.Conf = conf
}
