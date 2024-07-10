package plugins

import "os-proxy/bootstrap"

var lgLocal = new(LangGoLocal)

type LangGoLocal struct {
}

// 先将LangGoLocal实现接口，然后调用register函数完成注册
func init() {
	p := &LangGoLocal{}
	RegisterdPlugin(p)
}

// 根据配置文件决定是否生成实例
func (lg *LangGoLocal) Flag() bool {
	return bootstrap.NewConfig("").Local.Enabled
}

func (lg *LangGoLocal) Name() string {
	return "Local"
}

func (lg *LangGoLocal) New() interface{} {
	return nil
}

func (lg *LangGoLocal) Health() {

}

func (lg *LangGoLocal) Close() {

}
