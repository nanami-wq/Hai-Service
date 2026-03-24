package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    any         `json:"code"`
	ErrCode any         `json:"err_code"`
	Data    interface{} `json:"data,omitempty"`
	Msg     string      `json:"message,omitempty"`
}

func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 处理请求

		// 获取处理结果的状态码
		status := c.Writer.Status()

		// 如果响应已经写入，直接返回
		if c.Writer.Written() {
			return
		}

		// 获取原始响应数据
		var data interface{}
		if c.Keys != nil {
			data = c.Keys["data"]
		}
		msg := c.Keys["message"]
		code := c.Keys["code"]
		if code == nil {
			code = status
		}
		//codeInt := code.(int)

		if status == 404 && msg == nil {
			msg = "Not Found"
		}

		var errCode any
		errCode = code
		if codeInt, ok := code.(int); ok && codeInt == 200 {
			errCode = 200
		}

		// 构建统一响应结构体
		response := Response{
			Code:    code,
			ErrCode: errCode,
			Data:    data,
			Msg:     fmt.Sprintf("%v", msg),
		}

		// 以 JSON 形式返回响应
		c.JSON(status, response)
	}
}
