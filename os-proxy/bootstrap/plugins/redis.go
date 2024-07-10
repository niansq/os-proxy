package plugins

import (
	"context"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"os-proxy/bootstrap"
	"os-proxy/config"
	"sync"
)

var lgRedis = new(LangGoRedis)

type LangGoRedis struct {
	Once        *sync.Once //确保初始化函数只被执行一次
	RedisClient *redis.Client
}

func newLangGoRedis() *LangGoRedis {
	return &LangGoRedis{
		RedisClient: &redis.Client{},
		Once:        &sync.Once{},
	}
}

func (lg *LangGoRedis) NewRedis() *redis.Client {
	if lgRedis.RedisClient != nil {
		return lgRedis.RedisClient
	} else {
		return lg.New().(*redis.Client)
	}
}

// 将插件注册到plugins，方便后续注册
func init() {
	p := &LangGoRedis{}
	RegisterdPlugin(p)
}

// redis一定会用到
func (lg *LangGoRedis) Flag() bool {
	return true
}

func (lg *LangGoRedis) Name() string {
	return "Redis"
}

// 这个New函数会初始化redis,主要是在bootstrap/plugin中使用
func (lg *LangGoRedis) New() interface{} {
	lgRedis = newLangGoRedis()
	lgRedis.initRedis(bootstrap.NewConfig(""))
	return lgRedis.RedisClient
}

// 初始化redis,使用sync.Once确保只会执行一次
func (lg *LangGoRedis) initRedis(conf *config.Configuration) {
	lg.Once.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr: conf.Redis.Host + ":" + conf.Redis.Port,
			DB:   conf.Redis.DB,
		})
		client.AddHook(redisotel.TracingHook{}) // redis链路追踪相关
		lgRedis.RedisClient = client
	})
}

func (lg *LangGoRedis) Health() {
	if err := lgRedis.RedisClient.Ping(context.Background()).Err(); err != nil {
		bootstrap.NewLogger().Logger.Error("redis connect failed, err:", zap.Any("err", err))
		panic(err)
	}
}

func (lg *LangGoRedis) Close() {
	if lg.RedisClient == nil {
		return
	} else {
		if err := lg.RedisClient.Close(); err != nil {
			bootstrap.NewLogger().Logger.Error("redis close failed, err:", zap.Any("err", err))
		}
	}
}

/*
New() 函数：用于插件的初始化过程，确保插件在系统启动时被正确初始化。
NewRedis() 函数：用于在系统运行时获取 Redis 客户端实例，确保在整个应用中只使用一个 Redis 客户端实例。
*/
