package myGin

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// Recovery panic错误处理中间件
func Recovery() HandlerFunc {
	return func(context *Context) {
		// 使用defer挂载上错误恢复的函数，在这个函数中调用recover()捕获panic
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				context.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		context.Next()
	}
}

// 使用Go的运行时(runtime)来获取触发panic的堆栈信息
func trace(message string) string {
	var pcs [32]uintptr
	// 使用runtime.Callers获取调用栈的程序计数器(跳过了前三个Caller)
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		// 调用runtime.FuncForPC获取对应的函数
		fn := runtime.FuncForPC(pc)
		// 调用Func.FileLine获取函数的文件名和行号
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
