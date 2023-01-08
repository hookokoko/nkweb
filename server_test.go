package nkweb

import (
	"fmt"
	"net/http"
	"testing"
)

func Test_Sever(t *testing.T) {
	s := NewHttpServer()
	//var sI Server
	//sI = s
	s.Get("/", func(ctx *Context) {
		_, _ = ctx.Resp.Write([]byte("hello world"))
	})

	s.Get("/user", func(ctx *Context) {
		_, _ = ctx.Resp.Write([]byte("hello world user"))
	})

	s.Post("/user", func(ctx *Context) {
		_, _ = ctx.Resp.Write([]byte("hello world user post"))
	})

	s.Post("/user/:name", func(ctx *Context) {
		_, _ = ctx.Resp.Write([]byte(fmt.Sprintf("hello, user : %s, post", ctx.PathParams["name"])))
	})

	err := s.Start(":8080")
	fmt.Println(err)
}

func Test_Group(t *testing.T) {
	s := NewHttpServer()
	g := s.Group("/v1")
	g.AddRoute(http.MethodGet, "/user/:name/detail", func(ctx *Context) {
		ctx.Resp.Write([]byte(fmt.Sprintf("hello, %s, get details", ctx.PathParams["name"])))
	})
	err := s.Start(":8080")
	fmt.Println(err)
}
