package http

import "github.com/gin-gonic/gin"

const (
	SuccessCode = 200
	ErrorCode   = 500
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// Success 返回成功响应
func Success(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code: SuccessCode,
		Msg:  "success",
		Data: data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, msg string, data any) {
	c.JSON(200, Response{
		Code: ErrorCode,
		Msg:  msg,
		Data: data,
	})
}
