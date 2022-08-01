package myGin

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

/*
	myGin框架入口
*/

type HandlerFunc func(*Context)

// RouterGroup 定义路由组结构体，为路由添加分组功能
type RouterGroup struct {
	prefix      string        // 路由组前缀
	middlewares []HandlerFunc // 支持中间件功能,中间件与路由组绑定
	parent      *RouterGroup  // 当前分组的父分组
	engine      *Engine       // 通过engine来访问router,所有group共享同一个engine
}

// Engine 改变Engine结构，嵌入(相当于继承)RouterGroup，作为最顶层的路由分组，同时维护所有的RouterGroup
type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup     // 维护所有的路由组
	htmlTemplates *template.Template //html模板渲染,负责将模板加载进内存
	funcMap       template.FuncMap   //html模板渲染,所有的自定义模板渲染函数
}

// New 构造函数初始化Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Default 默认引入logger和recovery两个全局中间件的engine
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

/*
	由于Engine嵌入了RouterGroup,拥有了RouterGroup的所有特性
	故可以将和路由有关的函数交给RouterGroup实现而不用Engine
*/

// Group 建立新路由组
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// Use 将中间件应用到某个RouterGroup
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}

func (group *RouterGroup) DELETE(pattern string, handler HandlerFunc) {
	group.addRoute("DELETE", pattern, handler)
}

// Static 创建静态文件资源服务 api method为GET
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

// 创建静态文件资源handler函数, 负责解析请求的资源地址，其余任务交给http.FileServer处理
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absoluteFilePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absoluteFilePath, http.FileServer(fs))
	return func(context *Context) {
		file := context.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			context.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(context.Writer, context.Req)
	}
}

// SetFuncMap 用于设置自定义模板渲染函数
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// LoadHTMLGlob 用来加载模板
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// 当接收一个请求时，要判断该请求适用于哪些中间件，这里简单通过URL前缀来判断；之后初始化context
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

// Run http服务启动函数，通过Go标准库的net/http包实现
func (engine *Engine) Run(addr string) (err error) {
	fmt.Printf("HTTP Server Start Successfully at: %s\n", addr)
	return http.ListenAndServe(addr, engine)
}
