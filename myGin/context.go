package myGin

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTP Context

// H 一个map,键(string),值(空接口) 空接口可以传入所有的类型,相当于map的value是一个泛型
type H map[string]interface{}

// Context 定义http上下文(Context)
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
	// middleware handlers 当前context需要执行的handler列表，包含业务逻辑和中间件操作
	handlers []HandlerFunc
	// 记录当前context执行到第几个中间件handler
	index int
	// Engine pointer 主要用来访问engine中的html模板
	engine *Engine
}

// 初始化context
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// Next 调用该方法会执行下一个中间件
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// PostForm 处理post请求的表单参数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// Query 处理请求中url里的键值参数, 例如/user?username=cjm
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Param 获取请求中url里的直接参数, 例如/user/cjm
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// Status 处理http响应码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// SetHeader 处理http请求头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String 处理字符串型返回数据
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 处理json格式返回数据
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 处理html模板
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
