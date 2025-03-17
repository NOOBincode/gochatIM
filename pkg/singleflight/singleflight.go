package singleflight

import (
	"golang.org/x/sync/singleflight"
)

// Group 是 singleflight 组的包装
type Group struct {
	g *singleflight.Group
}

// NewGroup 创建一个新的 singleflight 组
func NewGroup() *Group {
	return &Group{
		g: &singleflight.Group{},
	}
}

// Do 执行函数，确保对于同一个 key 只会执行一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	return g.g.Do(key, fn)
}

// DoChan 异步执行函数，返回一个 channel
func (g *Group) DoChan(key string, fn func() (interface{}, error)) <-chan singleflight.Result {
	return g.g.DoChan(key, fn)
}

// Forget 使 key 对应的进行中的 Do 操作失效
func (g *Group) Forget(key string) {
	g.g.Forget(key)
}

// 为了向后兼容，保留全局 Group 变量
var defaultGroup = NewGroup()

// Do 使用默认组执行函数
func Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	return defaultGroup.Do(key, fn)
}

// DoChan 使用默认组异步执行函数
func DoChan(key string, fn func() (interface{}, error)) <-chan singleflight.Result {
	return defaultGroup.DoChan(key, fn)
}

// Forget 使默认组中 key 对应的进行中的 Do 操作失效
func Forget(key string) {
	defaultGroup.Forget(key)
}