package repeatbody

import (
	"bytes"
	"io"

	"github.com/nkweb"
)

//func RepeatBody() nkweb.Middleware {
//	return func(next nkweb.HandlerFunc) nkweb.HandlerFunc {
//		return func(ctx *nkweb.Context) {
//			ctx.Req.Body = io.NopCloser(ctx.Req.Body)
//			next(ctx)
//		}
//	}
//}

func RepeatBody() nkweb.Middleware {
	return func(next nkweb.HandlerFunc) nkweb.HandlerFunc {
		return func(ctx *nkweb.Context) {
			// 把request的内容读取出来
			var bodyBytes []byte
			if ctx.Req.Body != nil {
				bodyBytes, _ = io.ReadAll(ctx.Req.Body)
				_ = ctx.Req.Body.Close()
			}
			// 把刚刚读出来的再写进去
			ctx.Req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			next(ctx)
		}
	}
}
