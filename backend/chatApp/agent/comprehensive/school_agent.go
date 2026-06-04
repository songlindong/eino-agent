package comprehensive

import (
	"context"
	"eino-agent/backend/chatApp/chat"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
)

func NewSchoolAgent() (adk.Agent, error) {
	ctx := context.Background()

	model, err := chat.CreatOpenAiChatModel(ctx)
	if err != nil {
		log.Fatalf("创建大模型失败: %v", err)
	}

	baseAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "SchoolAgent",
		Description:   "校招综合面试官智能体，全面评估应届毕业生的综合能力",
		Instruction:   SchoolComprehensiveAgentInstruction,
		Model:         model,
		MaxIterations: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create school comprehensive agent: %w", err)
	}

	return baseAgent, nil
}
