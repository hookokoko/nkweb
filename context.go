package nkweb

import (
	"encoding/json"
	"net/http"
)

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	PathParams map[string]string
	// 命中的路由
	MatchedRoute string

	// 缓存下 响应
	RespStatusCode int
	RespData       []byte
}

type H map[string]any

func (ctx *Context) RespJson(code int, val any) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ctx.RespStatusCode = code
	ctx.RespData = bs
	return nil
}
