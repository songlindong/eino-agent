package main

import (
	"bufio"
	"context"
	"eino-agent/backend/chatApp/agent/rag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// ctx := context.Background()

	// fmt.Println("===1. 生成消息列表 ======")
	// messages := test.MessageTemplate()

	// fmt.Println("====2.创建大模型实例=====")
	// chatModel := chat.CreatOpenAiChatModel(ctx)

	// fmt.Println("====3. 流式输出AI回复====")
	// steam := test.Steam(ctx, chatModel, messages)

	// test.ReportSteam(steam)

	agent, toolset, err := rag.NewRAGAgent()
	if err != nil {
		log.Fatalf("智能体创建失败：%v", err)
	}
	defer func() {
		if err := toolset.Close(); err != nil {
			log.Printf("关闭 MCP 客户端失败：%v", err)
		}
	}()

	AgentTest(agent)
}

func AgentTest(agent adk.Agent) {
	ctx := context.Background()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})

	fmt.Println("========Agent 交互测试 ===========")
	fmt.Println("输入问题进行测试，输入'exit' 或 'quit' 退出")
	fmt.Println("============================")

	processInput := func(input string) {
		message := []*schema.Message{
			schema.UserMessage(input),
		}

		//启动智能体
		iter := runner.Run(ctx, message)

		fmt.Print("Agent:")

		for {
			event, ok := iter.Next()

			if !ok {
				break
			}

			if event.Err != nil {
				log.Fatalf("智能体运行失败: %v", event.Err)
				break
			}

			// 打印智能体输出
			if event.Output != nil && event.Output.MessageOutput != nil {
				content := event.Output.MessageOutput.Message.Content
				if content != "" {
					fmt.Printf("%s", content)
				}
			}
		}
		fmt.Println()
	}

	autoInput := strings.TrimSpace(os.Getenv("RAG_TEST_QUERY"))
	if autoInput == "" {
		autoInput = "请先调用 rag.retrieve 检索知识库内容，再根据检索结果给出总结。"
	}

	if autoInput != "" {
		fmt.Printf("自动输入: %s\n", autoInput)
		processInput(autoInput)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n 用户输入：")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("用户输入失败: %v", err)
			continue
		}
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("退出测试")
			break
		}
		processInput(input)
	}

}
