package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	tail_n    = "n"
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
		pathMd5, tailN := c.DefaultQuery(param_md5, logFiles.first), c.DefaultQuery(tail_n, "")
		data := struct {
			FS map[string]*LogFile
			F  *LogFile
			TN string
		}{
			FS: logFiles.m,
			TN: tailN,
		}
		if logFile, ok := logFiles.m[pathMd5]; ok {
			data.F = logFile
		} else {
			data.F = &LogFile{}
		}
		c.HTML(http.StatusOK, html_log, data)
	})
	// tail日志输出websocket地址
	router.GET("/ws", func(c *gin.Context) {
		pathMd5, tailN := c.DefaultQuery(param_md5, logFiles.first), c.DefaultQuery(tail_n, "")
		tailOptions := make([]string, 0)
		_, err := strconv.Atoi(tailN)
		if err == nil {
			tailOptions = append(tailOptions, "-n", tailN)
		}
		if l, ok := logFiles.m[pathMd5]; ok {
			serveWs(c.Writer, c.Request, l.Path, tailOptions)
		} else {
			serveWs(c.Writer, c.Request, "", tailOptions)
		}
	})

	router.Run(*addr)
}
