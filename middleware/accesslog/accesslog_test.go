package accesslog

import (
	"fmt"
	"testing"

	"github.com/nkweb"
)

func Test_Sever(t *testing.T) {
	s := nkweb.NewHttpServer()
	s.Use(NewBuilder().Build())
	s.Get("/", func(ctx *nkweb.Context) {
		ctx.RespJson(200, nkweb.H{"msg": "hello world"})
	})

	s.Get("/user", func(ctx *nkweb.Context) {
		ctx.RespJson(200, nkweb.H{"msg": "hello world user"})
	})

	s.Post("/user", func(ctx *nkweb.Context) {
		ctx.RespJson(200, nkweb.H{"msg": "hello world user post"})
	})

	s.Get("/user/:name", func(ctx *nkweb.Context) {
		ctx.RespJson(200, nkweb.H{
			"msg": fmt.Sprintf("hello, user : %s, post", ctx.PathParams["name"]),
		})
	})

	err := s.Start(":8080")
	fmt.Println(err)
}
