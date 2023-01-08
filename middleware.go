package nkweb

type Middleware func(next HandlerFunc) HandlerFunc
