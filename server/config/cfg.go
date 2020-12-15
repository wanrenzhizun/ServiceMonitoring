package config

import (
	"flag"
	"github.com/go-ini/ini"
)

var Cfg *ini.File
var UserFilePath = "./config/user.ini"

func init() {
	users := flag.String("u", "./config/user.ini", "指定用户配置文件")
	flag.Parse()
	InitConfig(*users)
}

func InitConfig(userPath string) {
	if len(userPath) != 0 {
		UserFilePath = userPath
	}
	Cfg, _ = ini.Load(UserFilePath)

}
