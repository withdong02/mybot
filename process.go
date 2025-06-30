package main

import (
	"context"
	"fmt"
	"github.com/bregydoc/gtranslate"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
	"golang.org/x/text/language"
	"strings"
)

// Processor is a struct to process message
type Processor struct {
	api openapi.OpenAPI
}

// ProcessGroupMessage 回复群消息
func (p *Processor) ProcessGroupMessage(input string, data *dto.WSGroupATMessageData) error {
	if strings.HasPrefix(input, "查询 ") {
		word := strings.TrimPrefix(input, "查询 ") // 提取单词（如 "apple"）
		meaning, err := queryWordMeaning(word)
		if err != nil {
			meaning = "翻译失败: " + err.Error()
		}

		reply := &dto.MessageToCreate{
			Content: fmt.Sprintf("单词 '%s' 的中文意思是：%s", word, meaning),
			MsgType: dto.TextMsg,
		}

		_, err = p.api.PostGroupMessage(context.Background(), data.GroupID, reply)
		return err
	}
	return nil
}
func queryWordMeaning(word string) (string, error) {
	translated, err := gtranslate.Translate(
		word,
		language.English, // 源语言：英语
		language.Chinese, // 目标语言：中文
	)
	if err != nil {
		return "", fmt.Errorf("翻译失败: %v", err)
	}
	return translated, nil
}
