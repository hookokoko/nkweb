package nkweb

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Context struct {
	Resp       http.ResponseWriter
	Req        *http.Request
	PathParams map[string]string
	// 命中的路由
	MatchedRoute string

	// 缓存下 响应
	RespStatusCode int
	RespData       []byte

	tplEngine TemplateEngine

	queryValues url.Values
}

type H map[string]any

func (ctx *Context) RespJSON(code int, val any) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ctx.RespStatusCode = code
	ctx.RespData = bs
	return nil
}

func (ctx *Context) RespOK(msg string) error {
	ctx.RespStatusCode = http.StatusOK
	ctx.RespData = []byte(msg)
	return nil
}

func (ctx *Context) RespJSONOK(val any) error {
	return ctx.RespJSON(http.StatusOK, val)
}

func (ctx *Context) BindJSON(val any) error {
	if ctx.Req.Body == nil {
		return errors.New("web: body 为 nil")
	}
	// bs, _:= io.ReadAll(c.Req.Body)
	// json.Unmarshal(bs, val)
	decoder := json.NewDecoder(ctx.Req.Body)
	// useNumber => 数字就是用 Number 来表示
	// 否则默认是 float64
	// if jsonUseNumber {
	// 	decoder.UseNumber()
	// }

	// 如果要是有一个未知的字段，就会报错
	// 比如说你 User 只有 Name 和 Email 两个字段
	// JSON 里面额外多了一个 Age 字段，那么就会报错
	// decoder.DisallowUnknownFields()
	return decoder.Decode(val)
}

func (ctx *Context) RespServerError(msg string) error {
	ctx.RespData = []byte(msg)
	ctx.RespStatusCode = 500
	return nil
}

func (ctx *Context) Render(tplName string, data any) error {
	var err error
	ctx.RespData, err = ctx.tplEngine.Render(ctx.Req.Context(), tplName, data)
	if err != nil {
		ctx.RespStatusCode = http.StatusInternalServerError
		return err
	}
	ctx.RespStatusCode = http.StatusOK
	return nil
}

// QueryValue Query和Form比起来，它的底层实现没有缓存
// 因此，我们在context中放入一个queryValues字段缓存住它
// 这样每次调用时便不用再解析一遍了
func (ctx *Context) QueryValue(key string) (string, error) {
	if ctx.queryValues == nil {
		ctx.queryValues = ctx.Req.URL.Query()
	}

	vals, ok := ctx.queryValues[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return vals[0], nil
}

func (ctx *Context) FormValue(key string) (string, error) {
	if err := ctx.Req.ParseForm(); err != nil {
		return "", err
	}
	return ctx.Req.FormValue(key), nil
}

func (ctx *Context) PathValue(key string) (string, error) {
	val, ok := ctx.PathParams[key]
	if !ok {
		return "", errors.New("web: key 不存在")
	}
	return val, nil
}
