package nkweb

import (
	"log"
	"net/http"
	"strconv"
)

type Server interface {
	// Handler 必须保证继承 ServeHTTP
	http.Handler
	// Start 管理生命周期
	Start(addr string) error
	// AddRoute 又回看看了一遍视频，作为web框架的核心方法，这个还是需要支持
	// GET POST 等这些都是基于这个方法提供的语法糖
	// 是否使用私有, 待定
	AddRoute(method, path string, handler HandlerFunc)
}

type HandlerFunc func(ctx *Context)

type MyServer struct {
	*router
	// middleware
	ms []Middleware
}

var _ Server = &MyServer{}

// NewHttpServer 通过接口就能完成不同的server实现
func NewHttpServer() *MyServer {
	return &MyServer{
		router: newRouter(),
	}
}

func (s *MyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{
		Writer: w,
		Req:    r,
	}
	root := s.serve
	for i := len(s.ms) - 1; i >= 0; i-- {
		root = s.ms[i](root) // 这里并没有执行，只是把middleware从后往前串起来了
	}

	// 新加入的一个middleware，回写响应，使能够缓存住响应码和响应体
	var m Middleware = func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) {
			next(ctx)
			s.flush(ctx) // 因为是放在next之后执行的，所以最后才执行刷响应的操作
		}
	}
	// 这样就能放在链子的最前位置
	root = m(root)
	root(c) // 这里触发执行，最先执行的是最后串的middleware，也就是s.ms[0]
}

func (s *MyServer) flush(ctx *Context) {
	if ctx.RespStatusCode > 0 {
		ctx.Writer.WriteHeader(ctx.RespStatusCode)
	}
	ctx.Writer.Header().Set("Content-Length", strconv.Itoa(len(ctx.RespData)))
	if _, err := ctx.Writer.Write(ctx.RespData); err != nil {
		log.Fatal("web: 回写响应失败", err)
	}
}

// Use middleware注册
func (s *MyServer) Use(ms ...Middleware) {
	s.ms = ms
}

func (s *MyServer) Start(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *MyServer) AddRoute(method, path string, handler HandlerFunc) {
	s.router.addRoute(method, path, handler)
}

func (s *MyServer) serve(c *Context) {
	mi, ok := s.findRouter(c.Req.Method, c.Req.URL.Path)
	if !ok || mi.n == nil || mi.n.handler == nil {
		c.RespStatusCode = 404
		c.RespData = []byte("not found")
		return
	}
	c.PathParams = mi.pathParams
	c.MatchedRoute = mi.n.route
	mi.n.handler(c)
}

func (s *MyServer) Get(path string, handler HandlerFunc) {
	s.AddRoute(http.MethodGet, path, handler)
}

func (s *MyServer) Post(path string, handler HandlerFunc) {
	s.AddRoute(http.MethodPost, path, handler)
}

// Group 分组的功能，是一个非核心功能
// 因为用户能够根据Server轻易的封装一个group
// 从下面的实现一个可以看出，并不是很难，即创建一个struct包装一个server和一个前缀即可
// 如果需要应用中间件，可以加上一个中间间数组
type Group struct {
	prefix string
	server Server
}

func (s *MyServer) Group(prefix string) *Group {
	return &Group{
		prefix: prefix,
		server: s,
	}
}

func (g *Group) AddRoute(method, path string, handler HandlerFunc) {
	g.server.AddRoute(method, g.prefix+path, handler)
}
