package web

import (
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	trees map[string]*routeNode
}

type handleFunc func(ctx *Context)

func newRouter() *router {
	return &router{
		trees: map[string]*routeNode{},
	}
}

func (r *router) addRoute(method string, path string, handler handleFunc) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	root, ok := r.trees[method]
	// 这是一个全新的 HTTP 方法，创建根节点
	if !ok {
		// 创建根节点
		root = &routeNode{path: "/"}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突[/]")
		}
		root.handler = handler
		return
	}

	segs := strings.Split(path[1:], "/")
	// 开始一段段处理
	for _, seg := range segs {
		if seg == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.findChildOrCreate(seg)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handler
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return &matchInfo{node: root}, true
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	mi := &matchInfo{}
	for _, s := range segs {
		var matchParam bool
		root, matchParam, ok = root.findChild(s)
		if !ok {
			return nil, false
		}
		if matchParam {
			// 先检查是否匹配正则
			if root.regPattern != "" {
				reg, _ := regexp.Compile(root.regPattern)
				if reg.MatchString(s) {
					mi.addValue(root.regKey, s)
				} else {
					mi.addValue(root.path[1:], s)
				}
			} else {
				mi.addValue(root.path[1:], s)
			}
		}
	}
	mi.node = root
	return mi, true
}

type routeNode struct {
	path       string
	children   map[string]*routeNode
	handler    handleFunc
	starChild  *routeNode
	paramChild *routeNode
	isEndStar  bool
	regPattern string
	regKey     string
}

func (rn *routeNode) findChild(seg string) (*routeNode, bool, bool) {
	// 如果通配符在最后一段，那么匹配后面多段路由。例如 /a/b/* 可以匹配 /a/b/c/d/e/f
	if rn.isEndStar {
		return rn, false, true
	}
	if rn.starChild != nil &&
		rn.starChild.children == nil &&
		rn.starChild.starChild == nil &&
		rn.starChild.paramChild == nil {
		rn.starChild.isEndStar = true
		return rn.starChild, false, true
	}

	if rn.children == nil {
		if rn.paramChild != nil {
			return rn.paramChild, true, true
		}
		return rn.starChild, false, rn.starChild != nil
	}
	res, ok := rn.children[seg]
	if !ok {
		if rn.paramChild != nil {
			return rn.paramChild, true, true
		}
		return rn.starChild, false, rn.starChild != nil
	}
	return res, false, ok
}

// findChildOrCreate 查找子节点，如果子节点不存在就创建一个
// 并且将子节点放回去了 children 中
func (rn *routeNode) findChildOrCreate(seg string) *routeNode {
	if seg == "*" {
		if rn.starChild == nil {
			rn.starChild = &routeNode{path: seg}
		}
		return rn.starChild
	}

	// 以 : 开头，我们认为是参数路由/正则路由
	if seg[0] == ':' {
		if rn.starChild != nil {
			panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", seg))
		}
		if rn.paramChild != nil {
			if rn.paramChild.path != seg {
				panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", rn.paramChild.path, seg))
			}
		} else {
			rn.paramChild = &routeNode{path: seg}
			key, pattern := getRegParam(seg)
			if pattern != "" {
				rn.paramChild.regKey = key
				rn.paramChild.regPattern = pattern
			}
		}
		return rn.paramChild
	}

	if rn.children == nil {
		rn.children = make(map[string]*routeNode)
	}
	child, ok := rn.children[seg]
	if !ok {
		child = &routeNode{path: seg}
		rn.children[seg] = child
	}
	return child
}

func getRegParam(seg string) (key string, pattern string) {
	leftBracketIndex := strings.Index(seg, "(")
	if leftBracketIndex > 0 && seg[len(seg)-1] == ')' {
		key = seg[1:leftBracketIndex]
		pattern = seg[leftBracketIndex+1 : len(seg)-1]
	}
	return
}

type matchInfo struct {
	node       *routeNode
	pathParams map[string]string
}

func (m *matchInfo) addValue(key string, value string) {
	if m.pathParams == nil {
		// 大多数情况，参数路径只会有一段
		m.pathParams = map[string]string{key: value}
	}
	m.pathParams[key] = value
}
