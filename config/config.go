package config

import (
	"errors"
	"github.com/unknwon/goconfig"
	"log"
	"os"
)

const configFile = "/conf/conf.ini"

var File *goconfig.ConfigFile

// 加载会初始化方法
func init() {
	//
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	//fmt.Println(currentDir)
	configpath := currentDir + configFile

	if !fileExist(configpath) {
		panic(errors.New("配置文件不存在！"))
	}

	//判断  msserver.exe D:/123 会将后边目录添加
	len := len(os.Args)
	if len > 1 {
		dir := os.Args[1]
		if dir != "" {
			configpath = dir + configpath
		}
	}
	//读取文件
	File, err = goconfig.LoadConfigFile(configpath)
	if err != nil {
		log.Fatal("读取配置文件出错！", err)
	}
	//fmt.Println(File)
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename) //检查文件是否存在
	return err == nil || os.IsExist(err)
}
