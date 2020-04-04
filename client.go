package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/gorilla/websocket"
)

const (
	controlC = "^C"

	// websocket消息类型
	msg_kind_sys = "sys"
	msg_kind_log = "log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	// The websocket connection.
	conn *websocket.Conn
	// Context撤销函数，用于关闭日志文件的读取
	ctxCancelFunc context.CancelFunc
	// 需要读取的日志文件
	logFile string
	// tail 参数
	tailOptions []string
}

// read 读取客户端websocket消息
func (c *Client) read() {
	defer func() {
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			c.ctxCancelFunc()
			break
		}
		message = bytes.TrimSpace(message)
		log.Println("来自客户端消息：", string(message))
		if string(message) == controlC {
			c.ctxCancelFunc()
			break
		}
	}
	log.Println("退出client.read()")
}

// write 向客户端写websocket消息
func (c *Client) write(ctx context.Context) {
	defer func() {
		c.conn.Close()
	}()
	if c.logFile == "" {
		c.conn.WriteJSON(genWsMsg(msg_kind_sys, "未指定日志文件或日志文件不存在"))
		return
	}
	if !FileExists(c.logFile) {
		c.conn.WriteJSON(genWsMsg(msg_kind_sys, fmt.Sprintf("日志文件[%s]不存在", c.logFile)))
		return
	}
	c.tailOptions = append(c.tailOptions, "-f", c.logFile)
	cmd := exec.CommandContext(ctx, "tail", c.tailOptions...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.conn.WriteJSON(genWsMsg(msg_kind_sys, fmt.Sprintf("调用tail命令出错: %v", err)))
		return
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			log.Println("退出tail命令")
			break
		}
		c.conn.WriteJSON(genWsMsg(msg_kind_log, line))
	}
	cmd.Wait()
	log.Println("退出client.write()")
}

// 创建websocket json消息
func genWsMsg(kind, msg string) map[string]string {
	return map[string]string{
		"kind": kind,
		"msg":  msg,
	}
}

// serveWs 处理相应客户端的websocket请求
func serveWs(w http.ResponseWriter, r *http.Request, logFile string, tailOptions []string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	client := &Client{conn: conn, ctxCancelFunc: cancelFunc, logFile: logFile, tailOptions: tailOptions}

	go client.write(ctx)
	go client.read()
}
