package main

import (
	"bufio"
	"chatGptGo/logUtility"
	"context"
	"fmt"
	"github.com/eiannone/keyboard"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"strings"
)

func getAPIKey() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入 API 密钥:")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("读取 API 密钥时出错：", err)
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		log.Fatalln("API 密钥不能为空")
	}
	return apiKey
}

// 从配置文件中读取模型列表
func getModels() []string {
	// TODO: 从配置文件中读取模型列表
	return []string{
		"gpt-3.5-turbo-0301",
		"gpt-3.5-turbo",
		"gpt-4-32k-0314",
		"gpt-4-32k",
		"gpt-4-0314",
		"gpt-4",
		"text-davinci-003",
		"text-davinci-002",
		"text-curie-001",
	}
}

// 选择模型
func chooseModel(models []string) (string, error) {
	fmt.Println("请选择一个模型：")
	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, model)
	}

	fmt.Print("输入模型序号：")
	for {
		char, _, err := keyboard.GetSingleKey()
		if err != nil {
			return "", fmt.Errorf("获取键盘输入时出错: %v", err)
		}

		modelIndex := int(char - '1')
		if modelIndex >= 0 && modelIndex < len(models) {
			chosenModel := models[modelIndex]
			fmt.Printf("\r\n选择的模型是: %s\n", chosenModel)
			return chosenModel, nil
		}

		fmt.Print("\r\n无效的输入，请输入有效的模型序号：")
	}
}

func truncateQuestion(question string, maxTokens int) string {
	tokens := len(strings.Split(question, ""))
	if tokens > maxTokens {
		return strings.Join(strings.Split(question, "")[:maxTokens], "")
	}
	return question
}

// 调用 GPT-3 生成答案
// 调用 GPT-3 生成答案
func getAnswer(apiKey, question, model string) (bool, string) {
	conf := openai.DefaultConfig(apiKey)
	conf.HTTPClient = &http.Client{}
	gptClient := openai.NewClientWithConfig(conf)

	req := openai.ChatCompletionRequest{
		Model:     model,
		MaxTokens: 300,
		Stream:    false,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "assistant",
				Content: truncateQuestion(question, 300),
			},
		},
	}
	ctx := context.Background()
	resp, err := gptClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return false, fmt.Sprintf("chatGpt 返回错误: %v", err)
	}

	if len(resp.Choices) > 0 {
		return true, strings.TrimSpace(resp.Choices[0].Message.Content)
	} else {
		return false, "chatGpt 服务器没有任何返回"
	}
}

func main() {
	logUtility.SetLogrus("chatGptGo")
	apiKey := getAPIKey()
	models := getModels()
	if len(models) == 0 {
		log.Fatalln("没有可用的模型")
	}

	model, err := chooseModel(models)
	if err != nil {
		log.Fatalln("选择模型时出错：", err)
	}
	if model == "" {
		return
	}
	for {
		fmt.Print("请输入您的问题:")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			log.Fatalf("读取输入时出错：%v", scanner.Err())
		}
		question := scanner.Text()

		if question == "q" {
			fmt.Println("退出程序")
			os.Exit(0)
		}

		question = strings.TrimSpace(question)
		if question == "" {
			continue
		}
		logrus.Info(question)
		ok, answer := getAnswer(apiKey, question, model)
		if ok {
			resStr := "ChatGpt: " + answer
			fmt.Println(resStr)
			logrus.Info(resStr)
		} else {
			fmt.Println(answer)
			logrus.Error(answer)
		}
	}
}
