package main

import (
	"myGin"
	"net/http"
)

// 写一个v2专用的中间件来测试,阻塞v2路由组
/*func onlyForV2() myGin.HandlerFunc {
	return func(context *myGin.Context) {
		t := time.Now()
		context.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", context.StatusCode, context.Req.RequestURI, time.Since(t))
	}
}*/

// 测试框架--little test demo
func main() {
	// 获取myGin HTTP服务引擎
	var app = myGin.Default()
	// 写一个简单的http GET api
	app.GET("/", func(context *myGin.Context) {
		context.String(http.StatusOK, "Hello, it is myGin!")
	})
	// 测试Recovery中间件捕获panic错误
	app.GET("/panic", func(context *myGin.Context) {
		names := []string{"cjm"}
		context.String(http.StatusOK, names[100])
	})

	// 路由分组
	users := app.Group("/users")
	{
		// /users/login
		users.POST("/login", func(context *myGin.Context) {
			// ...todo code here

			// return json response
			context.JSON(http.StatusOK, myGin.H{
				"code": http.StatusOK,
				"msg":  "login successfully",
				"data": context.PostForm("username"),
			})
		})
		// /users/userInfo?username=xxx&age=xxx
		users.GET("/userInfo", func(context *myGin.Context) {
			// ... todo code here

			// 获取url query参数并返回
			context.JSON(http.StatusOK, myGin.H{
				"username": context.Query("username"),
				"age":      context.Query("age"),
			})
		})
		// 测试动态路由:
		// /users/check/jack  /users/check/mike
		users.GET("/check/:username", func(context *myGin.Context) {
			// ... todo code here

			// 获取url param参数并返回
			context.JSON(http.StatusOK, myGin.H{
				"code":     http.StatusOK,
				"msg":      "checked",
				"username": context.Param("username"),
			})
		})
	}

	// 启动HTTP服务 localhost:9000
	app.Run(":9000")
}
