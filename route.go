package nkweb

import (
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	trees map[string]*node
}
type node struct {
	// 到达该节点的完整路径
	route string

	path       string
	children   map[string]*node
	starChild  *node
	paramChild *node
	regexChild *node
	handler    HandleFunc
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
}

func newRouter() *router {
	return &router{
		make(map[string]*node, 0),
	}
}

func (r *router) addRoute(method string, path string, handler HandleFunc) {
	if path == "" {
		panic("path 为空字符串")
	}
	if path != "/" && !strings.HasPrefix(path, "/") {
		panic("path 必须以 / 开头")
	}
	if path != "/" && strings.HasSuffix(path, "/") {
		panic("path 不能以 / 结尾")
	}
	pathList := strings.Split(path[1:], "/")
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}
	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic(fmt.Sprintf("路由冲突 [%s]", root.path))
		}
		root.handler = handler
		root.route = "/"
		return
	}
	cur := root
	for _, seg := range pathList {
		if seg == "" {
			panic("path 不能有连续的 /")
		}
		cur = cur.childOrCreate(seg)
	}
	// 不支持同一个path位置的handler覆盖
	if cur.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突 [%s]", cur.path))
	}
	cur.route = path
	cur.handler = handler
}

func (r *router) findRouter(method string, path string) (*matchInfo, bool) {
	mi := &matchInfo{}
	root, ok := r.trees[method]
	if !ok {
		return nil, ok
	}
	if root.path == path && path == "/" {
		mi.n = root
		return mi, true
	}
	pathList := strings.Split(path[1:], "/")
	cur := root
	for _, seg := range pathList {
		var (
			matchParam      bool
			matchParamRegex bool
		)
		cur, matchParam, matchParamRegex = cur.childOf(seg)
		if cur == nil {
			return nil, false
		}

		if matchParamRegex {
			p, ok := match(seg, cur.path)
			if !ok {
				return mi, false
			}
			if mi.pathParams == nil {
				mi.pathParams = p
			}
			for k, v := range p {
				mi.pathParams[k] = v
			}
		}

		//mi.n = root
		if matchParam {
			if mi.pathParams == nil {
				mi.pathParams = make(map[string]string)
			}
			mi.pathParams[cur.path[1:]] = seg
		}

		if cur.path == "*" && cur.starChild == nil &&
			cur.paramChild == nil && cur.children == nil && cur.regexChild == nil {
			break
		}

	}
	mi.n = cur

	return mi, true
}

// node
// 参数匹配成功
// 正则匹配成功
func (n *node) childOf(path string) (*node, bool, bool) {
	node, ok := n.children[path]
	if !ok {
		// 正则匹配优先
		if n.regexChild != nil {
			return n.regexChild, false, true
		}
		// 参数匹配其次
		if n.paramChild != nil {
			return n.paramChild, true, false
		}
		// 通配符最后
		if n.starChild != nil {
			return n.starChild, false, false
		}

		return nil, false, false
	}
	return node, false, false
}

func (n *node) childOrCreate(path string) *node {
	// 命中通配符匹配
	if path == "*" {
		if n.paramChild != nil {
			panic("web: 非法路由，已有参数路由。不允许同时注册参数路由和通配符路由 [" + path + "]")
		}
		if n.regexChild != nil {
			panic("web: 非法路由，已有正则路由。不允许同时注册正则路由和通配符路由 [" + path + "]")
		}
		if n.starChild == nil {
			n.starChild = &node{
				path: "*",
			}
		}
		return n.starChild
	}

	// 命中参数匹配
	if strings.HasPrefix(path, ":") {
		// 命中参数匹配，优先正则
		if _, _, ok := isMatch(path); ok {
			if n.paramChild != nil {
				panic("web: 非法路由，已有参数路由。不允许同时注册参数路由和正则路由 [" + path + "]")
			}
			if n.starChild != nil {
				panic("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [" + path + "]")
			}
			if n.regexChild != nil && n.regexChild.path != path && n.regexChild.handler != nil {
				panic(fmt.Sprintf("web: 路由冲突 [%s]", path))
			}
			if n.regexChild == nil {
				n.regexChild = &node{
					path: path,
				}
			}
			return n.regexChild
		}

		if n.starChild != nil {
			panic("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [" + path + "]")
		}
		if n.regexChild != nil {
			panic("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [" + path + "]")
		}
		if n.paramChild != nil && n.paramChild.path != path && n.paramChild.handler != nil {
			panic(fmt.Sprintf("web: 路由冲突 [%s]", path))
		}
		if n.paramChild == nil {
			n.paramChild = &node{
				path: path,
			}
		}
		return n.paramChild
	}

	// 普通匹配
	child, ok := n.children[path]
	if !ok {
		if n.children == nil {
			n.children = make(map[string]*node)
		}
		child = &node{
			path: path,
		}
		n.children[path] = child
	}

	return child
}

// 路径正则合法判断
func isMatch(path string) (string, string, bool) {
	// path必须是这种形式 :name(.*)
	r := regexp.MustCompile(`(:\w+)(\(.*\))`)
	match := r.FindStringSubmatch(path)
	if len(match) <= 2 {
		return "", "", false
	}
	pathParam := match[1]
	regexPath := match[2]
	return pathParam, regexPath, regexPath != ""
}

// 判断正则是不是匹配；提取正则参数
func match(path string, pattern string) (map[string]string, bool) {
	// path必须是这种形式 :name(.*)
	// 没有必要这里再来一次正则匹配，直接字符串分割，提高性能
	params := strings.SplitN(pattern, "(", 2)
	param := params[0]
	paramRegex := strings.TrimSuffix(params[1], ")")
	r := regexp.MustCompile(paramRegex)
	match := r.FindStringSubmatch(path)
	if len(match) < 1 {
		return nil, false
	}
	p := make(map[string]string, 1)
	p[param[1:]] = match[0]
	return p, true
}
