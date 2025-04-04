package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/lukin/hacker-news/go/internal/logger"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient is a client for the OpenAI API
type OpenAIClient struct {
	client    *openai.Client
	model     string
	maxTokens int
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, baseURL, model string, maxTokens int) *OpenAIClient {
	config := openai.DefaultConfig(apiKey)

	if baseURL != "" {
		config.BaseURL = baseURL
	}

	// Create a client with reasonable timeout
	config.HTTPClient.Timeout = 60 * time.Second

	client := openai.NewClientWithConfig(config)

	return &OpenAIClient{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}
}

// GenerateText generates text using the OpenAI API
func (c *OpenAIClient) GenerateText(systemPrompt, userPrompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// 记录提示词长度
	logger.Debug("系统提示词长度: %d 字符", len(systemPrompt))
	logger.Debug("用户提示词长度: %d 字符", len(userPrompt))

	// 限制用户提示词长度，粗略估计token数
	// 一般来说，对于英文文本，一个token大约是4个字符
	// GPT-4o支持128K上下文窗口，预留系统提示约1000个token和模型输出约16K个token
	// 那么用户提示词可以使用约100K个token，即400000个字符
	maxUserPromptChars := 400000
	if len(userPrompt) > maxUserPromptChars {
		logger.Debug("截断用户提示词，从 %d 字符截断至 %d 字符", len(userPrompt), maxUserPromptChars)
		userPrompt = userPrompt[:maxUserPromptChars]
		userPrompt += "\n\n[内容已截断...太长无法完全处理]"
	}

	// 截取用户提示词前100个字符用于调试
	userPromptPreview := userPrompt
	if len(userPromptPreview) > 100 {
		userPromptPreview = userPromptPreview[:100] + "..."
	}
	logger.Debug("用户提示词预览: %s", userPromptPreview)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	request := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   c.maxTokens,
		Temperature: 0.7,
	}

	// 尝试发送请求前记录日志
	logger.Debug("正在发送请求到OpenAI API (model: %s)", c.model)
	requestTimer := logger.NewTimer("OpenAI API请求")

	// 添加更详细的错误处理
	response, err := c.client.CreateChatCompletion(ctx, request)
	requestTimer.Stop()

	if err != nil {
		logger.Error("OpenAI API错误: %v", err)

		// 检查是否为API错误，尝试输出更多细节
		if apiErr, ok := err.(*openai.APIError); ok {
			logger.Error("OpenAI API错误详情 - 类型: %s, 代码: %s, 消息: %s",
				apiErr.Type, apiErr.Code, apiErr.Message)
		}

		return "", fmt.Errorf("生成文本失败: %w", err)
	}

	if len(response.Choices) == 0 {
		logger.Error("OpenAI API返回了空结果")
		return "", fmt.Errorf("响应中没有选择项")
	}

	logger.Debug("成功收到OpenAI API响应")

	// 截取响应前100个字符用于调试
	responseText := response.Choices[0].Message.Content
	responsePreview := responseText
	if len(responsePreview) > 100 {
		responsePreview = responsePreview[:100] + "..."
	}
	logger.Debug("响应预览: %s", responsePreview)

	return responseText, nil
}
