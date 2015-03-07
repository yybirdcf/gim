package main

//定义错误常量
const (
	//客户端错误 1000开始
	//客户端消息错误 1000~1099
	ERR_CLIENT_MSG_FORMAT      = 1000 //消息格式错误
	ERR_CLIENT_MSG_UNKNOW_TYPE = 1001 //未知的消息类型

	//客户端认证错误 1100~1199
	ERR_CLIENT_AUTH_FAILED = 1100 //认证失败
	ERR_CLIENT_AUTH_NOT    = 1003 //未认证

//服务器错误 2000开始
//系统错误 3000开始
)
