package rag

import (
	"context"
	"fmt"

	"eino-agent/backend/chatApp/chat"
	ragmcp "eino-agent/backend/chatApp/mcp"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
)

func NewRAGAgent() (adk.Agent, *ragmcp.Toolset, error) {
	ctx := context.Background()

	model, err := chat.CreatOpenAiChatModel(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
	}

	mcpCfg, err := ragmcp.LoadConfigFromEnv()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load rag mcp config: %w", err)
	}

	toolset, err := ragmcp.NewToolset(ctx, mcpCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize rag mcp tools: %w", err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "RAGRetrieveAgent",
		Description: "一个通过 MCP 调用 rag-retrievalOps 检索知识库内容的智能体",
		Instruction: `你是一个知识库检索与总结助手。

工作要求：
- 当用户的问题需要查询知识库时，优先调用 rag.retrieve 工具。
- 如果用户没有明确给出 kb_ids，也可以直接调用工具，让 MCP 使用默认知识库。
- 基于工具返回的 items 进行总结，优先引用高分结果。
- 如果工具返回空结果，要明确告诉用户“当前知识库未检索到相关内容”。
- 如果工具调用失败，直接说明失败原因，并提示检查 RAG_BASE_URL、RAG_API_KEY、知识库 ID 或知识库内容。

回答要求：
- 使用中文回答。
- 先给结论，再给依据。
- 不要编造知识库中不存在的事实。`,
		Model: model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: toolset.Tools,
			},
		},
		MaxIterations: 4,
	})
	if err != nil {
		_ = toolset.Close()
		return nil, nil, fmt.Errorf("failed to create rag agent: %w", err)
	}

	return agent, toolset, nil
}
