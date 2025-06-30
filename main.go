package main

import (
	"fmt"
	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/dto/message"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/interaction/webhook"
	"github.com/tencent-connect/botgo/token"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	host_ = "0.0.0.0"
	port_ = 9000
	path_ = "/qqbot"
)

// 消息处理器，持有 openapi 对象
var processor Processor

func main() {
	// 加载 appid 和 token
	content, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalln("load config file failed, err:", err)
	}
	//创建一个空的凭证结构体 QQBotCredentials；
	//使用 YAML 解析库将配置文件内容反序列化到该结构体中；
	credentials := &token.QQBotCredentials{
		AppID:     "",
		AppSecret: "",
	}
	if err = yaml.Unmarshal(content, &credentials); err != nil {
		log.Fatalln("parse config failed, err:", err)
	}
	log.Println("credentials:", credentials)

	tokenSource := token.NewQQBotTokenSource(credentials)

	// 创建一个可手动取消的上下文对象 ctx 和一个取消函数 cancel。
	//context 是 Go 语言中用于控制 函数调用生命周期 的工具
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() //释放刷新协程
	if err = token.StartRefreshAccessToken(ctx, tokenSource); err != nil {
		log.Fatalln(err)
	}
	// 初始化 openapi，正式环境
	api := botgo.NewOpenAPI(credentials.AppID, tokenSource).WithTimeout(5 * time.Second).SetDebug(true)
	processor = Processor{api: api}
	// 注册处理函数
	_ = event.RegisterHandlers(
		// ***********消息事件***********
		// 群@机器人消息事件
		GroupATMessageEventHandler(),
	)
	http.HandleFunc(path_, func(writer http.ResponseWriter, request *http.Request) {
		// 验证签名
		handleValidation(writer, request, credentials.AppSecret)
		webhook.HTTPHandler(writer, request, credentials)

	})
	if err = http.ListenAndServe(fmt.Sprintf("%s:%d", host_, port_), nil); err != nil {
		log.Fatal("setup server fatal:", err)
	}
}

// GroupATMessageEventHandler 实现处理 at 消息的回调
func GroupATMessageEventHandler() event.GroupATMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSGroupATMessageData) error {
		input := strings.ToLower(message.ETLInput(data.Content))
		return processor.ProcessGroupMessage(input, data)
	}
}
