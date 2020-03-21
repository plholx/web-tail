package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	logConfigFile = "log.li"
	filePeriod    = 10 * time.Second
	commentSign   = "#"
	sep           = "="
)

var (
	logLiLastMod time.Time

	logFiles = struct {
		first string
		m     map[string]*LogFile
	}{
		m: make(map[string]*LogFile),
	}
)

// LogFile 日志文件对象
type LogFile struct {
	Path  string
	Alias string
	// 日志文件路径的md5值
	PathMd5 string
}

// ParseLogConfigFileLiPeriod 定时解析日志配置文件
func ParseLogConfigFileLiPeriod() {
	for {
		fi, err := os.Stat(logConfigFile)
		if err == nil || os.IsExist(err) {
			curLastMod := fi.ModTime()
			if fi.ModTime().After(logLiLastMod) {
				logLiLastMod = curLastMod
				log.Println("重新加载日志配置文件")
				InitLogFiles()
			}
		} else {
			log.Printf("日志配置文件[%s]不存在\n", logConfigFile)
		}
		fileTicker := time.NewTicker(filePeriod)
		<-fileTicker.C
	}
}

// InitLogFiles 实例化日志文件列表
func InitLogFiles() {
	logFiles.m = make(map[string]*LogFile)
	logFiles.first = ""

	file, err := os.Open(logConfigFile)
	if err != nil {
		log.Printf("日志配置文件[%s]打开异常\n", logConfigFile)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		fmt.Println(line)
		if len(line) > 0 && !strings.HasPrefix(line, commentSign) {
			li := strings.Split(line, sep)
			if len(li) == 1 {
				AddLogFile(line, line)
			} else if len(li) > 1 {
				AddLogFile(li[0], li[1])
			}
		}
	}
}

// AddLogFile 向日志文件列表对象添加日志文件
func AddLogFile(path, alias string) {
	if FileExists(path) {
		pathMd5 := fmt.Sprintf("%x", md5.Sum([]byte(path)))
		logFile := &LogFile{
			Path:    path,
			Alias:   alias,
			PathMd5: pathMd5,
		}
		logFiles.m[pathMd5] = logFile
		if logFiles.first == "" {
			logFiles.first = pathMd5
		}
	} else {
		log.Printf("日志文件[%s]不存在\n", path)
	}
}

// FileExists 判断文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}
