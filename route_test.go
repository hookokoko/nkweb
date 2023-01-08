package nkweb

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_addRoute_new(t *testing.T) {
	mockHandler := func(ctx *Context) {}
	//mockHandler_diff := func(ctx *Context) {}
	r := newRouter()
	assert.PanicsWithValue(t, "path 为空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})
	assert.PanicsWithValue(t, "path 必须以 / 开头", func() {
		r.addRoute(http.MethodGet, "no_split_line_illegal", mockHandler)
	})
	assert.PanicsWithValue(t, "path 不能以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/behind_illegal/", mockHandler)
	})
	assert.PanicsWithValue(t, "path 不能有连续的 /", func() {
		r.addRoute(http.MethodGet, "/multi_line_illegal///123", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 路由冲突 [order]", func() {
		r.addRoute(http.MethodGet, "/order", mockHandler)
		r.addRoute(http.MethodGet, "/order", mockHandler)
	})
	//不支持不同正则同一位置重复注册，即使参数相同，也会报错
	assert.PanicsWithValue(t, "web: 路由冲突 [:id(\\d)]", func() {
		r.addRoute(http.MethodGet, `/order/:id(\d+)`, mockHandler)
		r.addRoute(http.MethodGet, `/order/:id(\d)`, mockHandler)
	})
	// 不支持不同参数同一位置重复注册，但是参数匹配若参数相同，则可覆盖注册
	assert.PanicsWithValue(t, "web: 路由冲突 [:name]", func() {
		r.addRoute(http.MethodGet, `/order/xxx/:id`, mockHandler)
		r.addRoute(http.MethodGet, `/order/xxx/:name`, mockHandler)
	})

	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/a/*", mockHandler)
		r.addRoute(http.MethodGet, "/a/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有参数路由。不允许同时注册参数路由和通配符路由 [*]", func() {
		r.addRoute(http.MethodGet, "/a/b/:id", mockHandler)
		r.addRoute(http.MethodGet, "/a/b/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有参数路由。不允许同时注册参数路由和通配符路由 [*]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有参数路由。不允许同时注册参数路由和正则路由 [:id(^\\d)]", func() {
		r.addRoute(http.MethodGet, "/:id", mockHandler)
		r.addRoute(http.MethodGet, `/:id(^\d)`, mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [:id]", func() {
		r.addRoute(http.MethodGet, `/:id(^\d)`, mockHandler)
		r.addRoute(http.MethodGet, "/:id", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有正则路由。不允许同时注册正则路由和通配符路由 [*]", func() {
		r.addRoute(http.MethodGet, `/:id(^\d)`, mockHandler)
		r.addRoute(http.MethodGet, "/*", mockHandler)
	})
	r = newRouter()
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [:id(^\\d)]", func() {
		r.addRoute(http.MethodGet, "/*", mockHandler)
		r.addRoute(http.MethodGet, `/:id(^\d)`, mockHandler)
	})

	testRoutes := []struct {
		method  string
		path    string
		handler HandleFunc
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
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		//{
		//	method: http.MethodGet,
		//	path:   "/param/:name/details",
		//},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则测试用例
		{
			method: http.MethodGet,
			path:   `/param/1/:id(^\d)`,
		},
		{
			method: http.MethodGet,
			path:   `/param/1/:id(^\d)/detail`,
		},
		{
			method: http.MethodGet,
			path:   `/param/1/:id(^\d)/*`,
		},
	}

	r = newRouter()
	for _, tr := range testRoutes {
		if tr.handler != nil {
			r.addRoute(tr.method, tr.path, tr.handler)
			continue
		}
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {path: "user", children: map[string]*node{
						"home": {path: "home", handler: mockHandler},
					}, handler: mockHandler},
					"order": {path: "order", children: map[string]*node{
						"detail": {path: "detail", handler: mockHandler},
					}, starChild: &node{path: "*", handler: mockHandler}},
					"param": {
						path: "param",
						paramChild: &node{
							path: ":id",
							starChild: &node{
								path:    "*",
								handler: mockHandler,
							},
							children: map[string]*node{"detail": {path: "detail", handler: mockHandler}},
							handler:  mockHandler,
						},
						children: map[string]*node{
							"1": &node{
								path: "1",
								regexChild: &node{
									path: `:id(^\d)`,
									children: map[string]*node{
										"detail": &node{path: "detail", handler: mockHandler},
									},
									starChild: &node{path: "*", handler: mockHandler}, handler: mockHandler}},
						},
					},
				},
				starChild: &node{
					path: "*",
					children: map[string]*node{
						"abc": {
							path:      "abc",
							starChild: &node{path: "*", handler: mockHandler},
							handler:   mockHandler},
					},
					starChild: &node{path: "*", handler: mockHandler},
					handler:   mockHandler},
				handler: mockHandler},

			http.MethodPost: {path: "/", children: map[string]*node{
				"order": {path: "order", children: map[string]*node{
					"create": {path: "create", handler: mockHandler},
				}},
				"login": {path: "login", handler: mockHandler},
			}},
		},
	}
	msg, ok := wantRouter.equal(*r)
	assert.True(t, ok, msg)
}

func Test_findRoute_new(t *testing.T) {
	mockHandler := func(ctx *Context) {}
	mockHandler_diff := func(ctx *Context) {}
	testRoutes := []struct {
		method  string
		path    string
		handler HandleFunc
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
			method:  http.MethodPost,
			path:    "/order/create",
			handler: mockHandler_diff,
		},
		{
			method: http.MethodGet,
			path:   "/user/*/home",
		},
		{
			method: http.MethodPost,
			path:   "/order/*",
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
			method: http.MethodGet,
			path:   `/param/reg/:id(\d+)/details`,
		},
		// 正则路由
		{
			method: http.MethodGet,
			path:   `/param/reg1/:id(user_name_\d+)`,
		},
		// 同时存在正则和参数
		{
			method: http.MethodGet,
			path:   "/param1/:name/:id(details_\\d+)/abc",
		},
		// 正则
		{
			method: http.MethodDelete,
			path:   "/reg/:id(.*)",
		},
		{
			method: http.MethodDelete,
			path:   "/:id([0-9]+)/home",
		},
	}

	r := newRouter()
	for _, tc := range testRoutes {
		if tc.handler != nil {
			r.addRoute(tc.method, tc.path, tc.handler)
		} else {
			r.addRoute(tc.method, tc.path, mockHandler)
		}
	}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		mi     *matchInfo
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
				n: &node{
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
				n: &node{
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
				n: &node{
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
				n: &node{
					path:    "create",
					handler: mockHandler_diff,
				},
			},
		},
		// 通配符匹配
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中/order/*
			name:   "star match",
			method: http.MethodPost,
			path:   "/order/delete/a/b/c",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			// 命中通配符在中间的
			// /user/*/home
			name:   "star in middle",
			method: http.MethodGet,
			path:   "/user/Tom/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "home",
					handler: mockHandler,
				},
			},
		},
		{
			// 比 /order/* 多了一段
			name:   "overflow",
			method: http.MethodPost,
			path:   "/order/delete/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "*",
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
				n: &node{
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
				n: &node{
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
				n: &node{
					path:    "detail",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},

		{
			// 命中 /param/reg/:id(\d+)/details
			name:   ":id(\\d+)",
			method: http.MethodGet,
			path:   "/param/reg/123/details",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "details",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},

		{
			// 命中 `/param/reg1/:id(user_name_\d+)`
			name:   `:id(user_name_\d+)`,
			method: http.MethodGet,
			path:   "/param/reg1/user_name_123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    `:id(user_name_\d+)`,
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "user_name_123"},
			},
		},

		{
			name:   "/param1/:name/:id(details_\\d+)/abc",
			method: http.MethodGet,
			path:   "/param1/reg/details_123/abc",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    "abc",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "details_123", "name": "reg"},
			},
		},

		{
			// 命中 /reg/:id(.*)
			name:   ":id(.*)",
			method: http.MethodDelete,
			path:   "/reg/123",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 命中 /:id([0-9]+)/home
			name:   ":id([0-9]+)",
			method: http.MethodDelete,
			path:   "/123/home",
			found:  true,
			mi: &matchInfo{
				n: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
				pathParams: map[string]string{"id": "123"},
			},
		},
		{
			// 未命中 /:id([0-9]+)/home
			name:   "not :id([0-9]+)",
			method: http.MethodDelete,
			path:   "/abc/home",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, found := r.findRouter(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			assert.Equal(t, tc.mi.pathParams, mi.pathParams)
			n := mi.n
			wantVal := reflect.ValueOf(tc.mi.n.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}

func Test_Mini1(t *testing.T) {
	mockHandler := func(ctx *Context) {}

	testRoutes := []struct {
		method  string
		path    string
		handler HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/param/:name/details",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id", // 这里会注册失败，因为只有一个paramchild,若支持需要改成map[string]*因为只有一个paramchild， 这种形式
		},
	}
	r := newRouter()
	for _, tr := range testRoutes {
		if tr.handler != nil {
			r.addRoute(tr.method, tr.path, tr.handler)
			continue
		}
		r.addRoute(tr.method, tr.path, mockHandler)
	}
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"param": {
						path: "param",
						paramChild: &node{
							path: ":name",
							children: map[string]*node{
								"details": &node{
									path:    "details",
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

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}
	//if len(n.children) == 0 {
	//	return "", true
	//}

	// 普通节点比较
	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	// 通配符比较
	if n.starChild != nil || y.starChild != nil {
		strStar, ok := n.starChild.equal(y.starChild)
		if !ok {
			return n.path + "-" + strStar, ok
		}
	}

	// 参数比较
	if n.paramChild != nil && y.paramChild != nil {
		paramStar, ok := n.paramChild.equal(y.paramChild)
		if !ok {
			return n.path + "-" + paramStar, ok
		}
	}
	// 正则比较
	if n.regexChild != nil && y.regexChild != nil {
		paramReg, ok := n.regexChild.equal(y.regexChild)
		if !ok {
			return n.path + "-" + paramReg, ok
		}
	}
	return "", true
}

func Benchmark_FindRoute_Simple(b *testing.B) {
	r := getTestRoute()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.findRouter(http.MethodGet, "/a/b/c/d")
	}
}

func Benchmark_FindRoute_Star(b *testing.B) {
	r := getTestRoute()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.findRouter(http.MethodGet, "/order/a/b/c/d")
	}
}

func Benchmark_FindRoute_Param(b *testing.B) {
	r := getTestRoute()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.findRouter(http.MethodGet, "/param/123/detail")
	}
}

func Benchmark_FindRoute_Regex(b *testing.B) {
	r := getTestRoute()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.findRouter(http.MethodGet, "/param/1/234/a/c/b")
	}
}

func getTestRoute() *router {
	r := newRouter()
	mockHandler := func(ctx *Context) {}
	//mockHandler_diff := func(ctx *Context) {}
	testRoutes := []struct {
		method  string
		path    string
		handler HandleFunc
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
		// 参数路由
		{
			method: http.MethodGet,
			path:   "/param/:id",
		},
		//{
		//	method: http.MethodGet,
		//	path:   "/param/:name/details",
		//},
		{
			method: http.MethodGet,
			path:   "/param/:id/detail",
		},
		{
			method: http.MethodGet,
			path:   "/param/:id/*",
		},
		// 正则测试用例
		{
			method: http.MethodGet,
			path:   `/param/1/:id(^\d)`,
		},
		{
			method: http.MethodGet,
			path:   `/param/1/:id(^\d)/detail`,
		},
		{
			method: http.MethodGet,
			path:   `/param/1/:id(^\d)/*`,
		},
	}

	for _, tr := range testRoutes {
		if tr.handler != nil {
			r.addRoute(tr.method, tr.path, tr.handler)
			continue
		}
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	return r
}

/**
goos: darwin
goarch: arm64
pkg: geektime/web/v1
Benchmark_FindRoute_Simple-8   	36832098	        82.97 ns/op	      80 B/op	       2 allocs/op
Benchmark_FindRoute_Star-8     	33547755	       103.5 ns/op	      96 B/op	       2 allocs/op
Benchmark_FindRoute_Param-8    	19717522	       169.2 ns/op	     400 B/op	       4 allocs/op
Benchmark_FindRoute_Regex-8    	 1200590	      2918 ns/op	    4979 B/op	      53 allocs/op
PASS
ok  	geektime/web/v1	17.244s
*/

// benchmark总结：
// 1. 简单匹配、通配符匹配、参数匹配的相差不是很大，但是正则匹配的性能最差,无论是cpu还是内存
// 2. 通过profile查看，最耗时的地方在regxp.MustComile方法上，通过火焰图发现 regxp.MustComile这个方法被调用了两次，于是优化了这个方法为调用一次。实际在 path必须是这种形式 :name(.*)
//    这个形式的验证改正则为字符串分割，不进行正则匹配，正则路由性能提高了2倍。以下是优化后结果：

/**
goos: darwin
goarch: arm64
pkg: geektime/web/v1
Benchmark_FindRoute_Simple-8   	13562884	        82.97 ns/op	      80 B/op	       2 allocs/op
Benchmark_FindRoute_Star-8     	11574576	       103.1 ns/op	      96 B/op	       2 allocs/op
Benchmark_FindRoute_Param-8    	 6617998	       179.4 ns/op	     400 B/op	       4 allocs/op
Benchmark_FindRoute_Regex-8    	 1234760	       976.9 ns/op	    1521 B/op	      19 allocs/op
*/
