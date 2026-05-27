package chat

import (
	"context"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func CreatOpenAiChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	APIkey := "ark-b1e3f513-33e7-4285-8395-c6d6ee1389d6-7db00"
	ModelName := "doubao-seed-1-6-flash-250828"
	BaseUrl := "https://ark.cn-beijing.volces.com/api/v3"

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  APIkey,
		Model:   ModelName,
		BaseURL: BaseUrl,
	})

	if err != nil {
		log.Fatalf("创建大模型失败: %v", err)
	}

	return chatModel, nil
}
