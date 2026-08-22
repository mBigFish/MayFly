package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResult 分页结果
type PageResult struct {
	List    interface{} `json:"list"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
}

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// OKMsg 自定义消息的成功响应
func OKMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: msg,
		Data:    data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
	})
}

// FailWithDetail 失败响应（带详情）
func FailWithDetail(c *gin.Context, msg string, detail interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    1,
		Message: msg,
		Data:    detail,
	})
}

// OKPage 分页成功响应
func OKPage(c *gin.Context, list interface{}, total int64, page, perPage int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageResult{
			List:    list,
			Total:   total,
			Page:    page,
			PerPage: perPage,
		},
	})
}

// ParsePage 从请求中解析分页参数
func ParsePage(c *gin.Context) (page, perPage int) {
	page = 1
	perPage = 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			perPage = n
		}
	}
	return
}
