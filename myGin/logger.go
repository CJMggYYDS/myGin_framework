package myGin

import (
	"log"
	"time"
)

// Logger 定义日志中间件
func Logger() HandlerFunc {
	return func(context *Context) {
		t := time.Now()
		// 交给下一个中间件
		context.Next()
		// 最后打印这次的信息
		log.Printf("[%d] %s in %v", context.StatusCode, context.Req.RequestURI, time.Since(t))
	}
}
