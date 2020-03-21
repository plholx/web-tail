package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	param_md5 = "md5"
	html_log  = "log.html"
)

var addr = flag.String("addr", ":8765", "http service address")

func main() {
	flag.Parse()

	// 定时解析日志配置文件
	go ParseLogConfigFileLiPeriod()

	router := gin.Default()
	// 加载模版文件
	router.LoadHTMLGlob("tmpls/*")
	// 日志查看页面
	router.GET("/log", func(c *gin.Context) {
		pathMd5 := c.DefaultQuery(param_md5, logFiles.first)
		data := struct {
			FS map[string]*LogFile
			F  *LogFile
		}{
			FS: logFiles.m,
		}
		data.F, _ = logFiles.m[pathMd5]
		c.HTML(http.StatusOK, html_log, data)
	})
	// tail日志输出websocket地址
	router.GET("/ws", func(c *gin.Context) {
		pathMd5 := c.DefaultQuery(param_md5, logFiles.first)
		if l, ok := logFiles.m[pathMd5]; ok {
			serveWs(c.Writer, c.Request, l.Path)
		} else {
			serveWs(c.Writer, c.Request, "")
		}
	})

	router.Run(*addr)
}
