package test

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func createChatTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		// 系统提示词
		schema.SystemMessage("你是{role},用{style}的语气回答,帮助用户回答面试上的问题"),
		schema.UserMessage("问题：{question}"),
	)
}

// 给大模型输入
func MessageTemplate() []*schema.Message {
	template := createChatTemplate()

	messages, err := template.Format(context.Background(), map[string]any{
		"role":     "经验丰富的大厂开发面试专家",
		"style":    "温和且专业，条理清晰",
		"question": "什么是Go语言？请从面试角度重点讲解核心特点",
	})

	if err != nil {
		log.Fatalf("格式化消息模板失败，%v", err)
	}

	return messages
}
