package main

import (
	"context"
	"eino-agent/backend/chatApp/chat"
	"eino-agent/backend/chatApp/test"
	"fmt"
)

func main() {
	ctx := context.Background()

	fmt.Println("===1. 生成消息列表 ======")
	messages := test.MessageTemplate()

	fmt.Println("====2.创建大模型实例=====")
	chatModel := chat.CreatOpenAiChatModel(ctx)

	fmt.Println("====3. 流式输出AI回复====")
	steam := test.Steam(ctx, chatModel, messages)

	test.ReportSteam(steam)
}
