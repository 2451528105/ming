package consts

type ErrorCode string

const (
	ErrorCode_Success    ErrorCode = "G000000" // 成功
	ErrorCode_RequestErr ErrorCode = "G000001" // 请求错误
)
