package accesslog

import (
	"encoding/json"
	"io"
	"log"

	"github.com/nkweb"
)

type MiddlewareBuilder struct {
	logFunc func(accesslog []byte)
}

type accesslog struct {
	Method     string
	Body       string
	Path       string
	MatchRoute string
}

func NewBuilder() MiddlewareBuilder {
	return MiddlewareBuilder{
		logFunc: func(accesslog []byte) {
			log.Println(string(accesslog))
		},
	}
}

func (m MiddlewareBuilder) Build() nkweb.Middleware {
	return func(next nkweb.HandlerFunc) nkweb.HandlerFunc {
		return func(ctx *nkweb.Context) {
			defer func() {
				body, _ := io.ReadAll(ctx.Req.Body) // 拿到body的字节流数据，但是需要注意只能读取一次
				l := accesslog{
					Method:     ctx.Req.Method,
					Body:       string(body),
					MatchRoute: ctx.MatchedRoute,
					Path:       ctx.Req.URL.String(),
				}
				//ctx.Req.Body = io.NopCloser(bytes.NewBuffer(body)) // 用该方法重新将数据写入Body用于传递
				// 使用另一个repeat中间件之后可以注释掉上面
				bs, err := json.Marshal(l)
				if err == nil {
					m.logFunc(bs)
				}
			}()

			next(ctx)
		}
	}
}
