package controller

// MyCode 响应状态码类型
type MyCode int64

const (
	CodeSuccess         MyCode = 1000 // 成功
	CodeInvalidParam    MyCode = 1001 // 请求参数错误
	CodeUserExist       MyCode = 1002 // 用户已存在
	CodeUserNotExist    MyCode = 1003 // 用户不存在
	CodeInvalidPassword MyCode = 1004 // 用户名或密码错误
	CodeServerBusy      MyCode = 1005 // 服务繁忙

	CodeNeedLogin      MyCode = 1006 // 需要登录
	CodeInvalidToken   MyCode = 1007 // 无效的token
	CodeResourceExists MyCode = 1008 // 资源已存在
	CodeResourceNotExist MyCode = 1009 // 资源不存在
	CodePermissionDenied MyCode = 1010 // 权限不足
)

var codeMsgMap = map[MyCode]string{
	CodeSuccess:          "success",
	CodeInvalidParam:     "请求参数错误",
	CodeUserExist:        "用户已存在",
	CodeUserNotExist:     "用户不存在",
	CodeInvalidPassword:  "用户名或密码错误",
	CodeServerBusy:       "服务繁忙",
	CodeNeedLogin:        "需要登录",
	CodeInvalidToken:     "无效的token",
	CodeResourceExists:   "资源已存在",
	CodeResourceNotExist: "资源不存在",
	CodePermissionDenied: "权限不足",
}

// Msg 返回状态码对应的消息
func (c MyCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}