package nkweb

type Middleware func(next HandleFunc) HandleFunc
