package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_addRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},

		// 通配符测试用例
		{
			method: http.MethodGet,
			path:   "/order/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/*",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc",
		},
		{
			method: http.MethodGet,
			path:   "/*/abc/*",
		},
		{
			method: http.MethodGet,
			path:   "/a/*",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则路由
		{
			method: http.MethodPost,
			path:   "/reg/:id(\\d+)",
		},
		{
			method: http.MethodPost,
			path:   "/reg/:id(\\d+)/bc",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*routeNode{
			http.MethodGet: {
				path: "/",
				starChild: &routeNode{
					path:    "*",
					handler: mockHandler,
					starChild: &routeNode{
						path:    "*",
						handler: mockHandler,
					},
					children: map[string]*routeNode{
						"abc": {
							path:    "abc",
							handler: mockHandler,
							starChild: &routeNode{
								path:    "*",
								handler: mockHandler,
							},
						},
					},
				},
				children: map[string]*routeNode{
					"user": {
						path: "user",
						children: map[string]*routeNode{
							"home": {
								path:    "home",
								handler: mockHandler,
							},
						}, handler: mockHandler},
					"order": {
						path: "order",
						starChild: &routeNode{
							path:    "*",
							handler: mockHandler,
						},
						children: map[string]*routeNode{
							"detail": {
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
					"param": {
						path: "param",
						paramChild: &routeNode{
							path: ":id",
							starChild: &routeNode{
								path:    "*",
								handler: mockHandler,
							},
							children: map[string]*routeNode{"detail": {path: "detail", handler: mockHandler}},
							handler:  mockHandler,
						},
					},
					"a": {
						path: "a",
						starChild: &routeNode{
							path:    "*",
							handler: mockHandler,
						},
					},
				},
				handler: mockHandler,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*routeNode{
					"order": {
						path: "order",
						children: map[string]*routeNode{
							"create": {
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": {
						path:    "login",
						handler: mockHandler,
					},
					"reg": {
						path: "reg",
						paramChild: &routeNode{
							path:       ":id(\\d+)",
							handler:    mockHandler,
							regKey:     "id",
							regPattern: "\\d+",
							children: map[string]*routeNode{
								"bc": {
									path:    "bc",
									handler: mockHandler,
								},
							},
						},
					},
				},
			},
		},
	}
	msg, ok := wantRouter.equal(*r)
	assert.True(t, ok, msg)

	//非法用例
	//r = newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 路由是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	//r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/]", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突[/a/b/c]", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})
}

func (r *router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 里面没有方法 %s 的路由树", k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

func (rn *routeNode) equal(y *routeNode) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if rn.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", rn.path, rn.path, y.path), false
	}

	nhv := reflect.ValueOf(rn.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", rn.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if len(rn.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", rn.path), false
	}
	if len(rn.children) == 0 {
		return "", true
	}

	for k, v := range rn.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", rn.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return rn.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/delete/check",
		},
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		{
			method: http.MethodGet,
			path:   "/a/*",
		},
		// 正则路由
		{
			method: http.MethodPost,
			path:   "/reg/:id(\\d+)",
		},
		{
			method: http.MethodPost,
			path:   "/reg/:id(\\d+)/bc",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name     string
		method   string
		path     string
		found    bool
		wantNode *routeNode
		mi       *matchInfo
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "/",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "user",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path: "order",
				},
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "create",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "order/*",
			method: http.MethodPost,
			path:   "/order/aaa",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "user/*/home",
			method: http.MethodGet,
			path:   "/user/1/home",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "order/delete/check",
			method: http.MethodPost,
			path:   "/order/delete/check",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "check",
					handler: mockHandler,
				},
			},
		},
		// 参数匹配
		{
			// 命中 /param/:id
			name:   ":id",
			method: http.MethodGet,
			path:   "/param/123",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    ":id",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /param/:id/*
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/abc",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "*",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},

		{
			// 命中 /param/:id/detail
			name:   ":id*",
			method: http.MethodGet,
			path:   "/param/123/detail",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			name:   "/a/*",
			method: http.MethodGet,
			path:   "/a/b/c/d",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			name:   "/reg/id()",
			method: http.MethodPost,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:       "/reg/:id(\\d+)",
					handler:    mockHandler,
					regKey:     "id",
					regPattern: "\\d+",
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			name:   "/reg/id()/bc",
			method: http.MethodPost,
			path:   "/reg/123/bc",
			found:  true,
			mi: &matchInfo{
				node: &routeNode{
					path:       "/reg/:id(\\d+)",
					handler:    mockHandler,
					regKey:     "id",
					regPattern: "\\d+",
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			n := mi.node
			wantVal := reflect.ValueOf(tc.mi.node.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
