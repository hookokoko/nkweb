package repeatbody

import (
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/nkweb"
)

func Test_Sever(t *testing.T) {
	s := nkweb.NewHttpServer()
	s.Use(RepeatBody())

	s.Post("/user/:name", func(ctx *nkweb.Context) {
		body1, err := io.ReadAll(ctx.Req.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println("body1: ", string(body1))

		body2, err := io.ReadAll(ctx.Req.Body)
		if err != nil {
			log.Println(err)
		}
		log.Println("body2: ", string(body2))

		ctx.RespJson(200, nkweb.H{
			"msg": fmt.Sprintf("hello, user : %s, post", ctx.PathParams["name"]),
		})
	})

	err := s.Start(":8080")
	fmt.Println(err)
}
