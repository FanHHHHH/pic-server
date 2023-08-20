package utils

import "github.com/gin-gonic/gin"

type ResponseApi struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewResponseApi(code int, message string, data interface{}) ResponseApi {
	return ResponseApi{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func SendJsonResponse(c *gin.Context, code int, msg string, data interface{}) ResponseApi {
	c.JSON(code, NewResponseApi(code, msg, data))
	return NewResponseApi(code, msg, data)
}
