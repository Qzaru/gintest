package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

const (
	ERROR   = 7
	SUCCESS = 0
)

func Result(code int, data interface{}, msg string, c *gin.Context) {
	// 开始时间
	c.JSON(http.StatusOK, Response{
		code,
		data,
		msg,
	})
}

func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", c)
}

func OkWithMessage(message string, c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

func OkWithData(data interface{}, c *gin.Context) {
	Result(SUCCESS, data, "成功", c)
}

func OkWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS, data, message, c)
}

func Fail(c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, "操作失败", c)
}

func FailWithMessage(message string, c *gin.Context) {
	Result(ERROR, map[string]interface{}{}, message, c)
}

func NoAuth(message string, c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		7,
		nil,
		message,
	})
}

func FailWithDetailed(data interface{}, message string, c *gin.Context) {
	Result(ERROR, data, message, c)
}

// 新弄一个
const (
	ERRORNOTFOUND                 = 404
	ERRORBADREQUEST               = 400
	SUCCESS1                      = 200
	SUCCESSMESSAGE                = "OK"
	ERRORCODE_INVALID_PARAMETER   = "INVALID_PARAMETER"
	ERRORCODE_NOT_FOUND           = "NOT_FOUND"
	ERRORMESSAGE_WRONGPRODUCTCODE = "不正な商品識別子です。"
	ERRORMESSAGE_WRONGPARAMETER1  = "pageパラメータは数値で指定してください。"
	ERRORMESSAGE_WRONGPARAMETER2  = "pageパラメータは1以上で指定してください。"
	ERRORMESSAGE_WRONGPARAMETER3  = "limitパラメータは1から100の間で指定してください。"
	ERRORMESSAGE_WRONGSORT        = "不正なsortパラメータです。('newest', 'oldest', 'most_helpful' のいずれかを指定。"
	ERRORMESSAGE_NOTFOUND         = "商品が見つかりません。"
	MESSAGE_FAVORITED             = "该商品已经收藏过了。"
	ERRORMESSAGE_FAVORITENOTFOUND = "お気に入りが見つかりません。"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target,omitempty"`
}

func Result2(code int, message string, target string, c *gin.Context) {
	c.JSON(http.StatusOK, ErrorResponse{
		ErrorDetail{code,
			message,
			target,
		},
	})
}
func FailWithDetailed2(code int, message string, target string, c *gin.Context) {
	Result2(code, message, target, c)
}
func OkWithDetailed2(data interface{}, message string, c *gin.Context) {
	Result(SUCCESS1, data, message, c)
}
