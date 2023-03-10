package nkweb

import (
	"html/template"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoginPage(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	require.NoError(t, err)
	engine := &GoTemplateEngine{
		T: tpl,
	}

	h := NewHttpServer(ServerWithTemplateEngine(engine))
	h.Get("/login", func(ctx *Context) {
		err := ctx.Render("login.gohtml", nil)
		if err != nil {
			log.Println(err)
		}
	})
	h.Start(":8081")
}
