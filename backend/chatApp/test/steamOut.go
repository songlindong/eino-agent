package test

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func ReportSteam(reader *schema.StreamReader[*schema.Message]) {
	defer reader.Close()

	for {
		message, err := reader.Recv()

		if err == io.EOF {
			return
		}

		if err != nil {
			log.Fatalf("读取流消息失败：%v", err)
		}

		fmt.Print(message.Content)
	}
}

func Steam(ctx context.Context, model model.ToolCallingChatModel, messages []*schema.Message) *schema.StreamReader[*schema.Message] {
	result, err := model.Stream(ctx, messages)

	if err != nil {
		log.Fatalf("模型生成流式回复失败：%v", err)
	}

	return result
}
