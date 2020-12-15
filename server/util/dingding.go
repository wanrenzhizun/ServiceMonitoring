package util

import (
	"github.com/blinkbean/dingtalk"
)


func SendDingDingMsg(){
	// 单个机器人有单位时间内消息条数的限制，如果有需要可以初始化多个token，发消息时随机发给其中一个机器人。
	var dingToken = []string{"06e40f02f9672b1d7cc5c494af9612ddc997d3aeb64588118788c969ebc8e916"}
	cli := dingtalk.InitDingTalk(dingToken, "admin")
	cli.SendTextMessage("哈哈哈")
}
