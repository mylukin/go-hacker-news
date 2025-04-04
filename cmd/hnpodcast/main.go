package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/lukin/hacker-news/go/internal/ai"
	"github.com/lukin/hacker-news/go/internal/config"
	"github.com/lukin/hacker-news/go/internal/fetcher"
	"github.com/lukin/hacker-news/go/internal/logger"
	"github.com/lukin/hacker-news/go/internal/prompt"
)

// getEnvInt 从环境变量中读取整数值，如果不存在或无法解析则返回默认值
func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func main() {
	// 获取默认配置
	defaultConfig := config.NewDefaultConfig()

	// 解析命令行参数
	var (
		date        = flag.String("date", time.Now().Format("2006-01-02"), "Date to fetch stories for (YYYY-MM-DD)")
		outputDir   = flag.String("output", "output", "Directory to store output files")
		openAIKey   = flag.String("openai-key", os.Getenv("OPENAI_API_KEY"), "OpenAI API key")
		openAIURL   = flag.String("openai-url", os.Getenv("OPENAI_BASE_URL"), "OpenAI API base URL")
		openAIModel = flag.String("openai-model", os.Getenv("OPENAI_MODEL"), "OpenAI model to use")
		jinaKey     = flag.String("jina-key", os.Getenv("JINA_KEY"), "Jina API key")
		maxStories  = flag.Int("max-stories", defaultConfig.MaxStories, "Maximum number of stories to process")
		maxTokens   = flag.Int("max-tokens", defaultConfig.OpenAIMaxTokens, "Maximum number of tokens for OpenAI responses")
		dev         = flag.Bool("dev", defaultConfig.Development, "Run in development mode")
		debug       = flag.Bool("debug", false, "Enable debug logging")
		httpTimeout = flag.Int("timeout", getEnvInt("HTTP_TIMEOUT", defaultConfig.HTTPTimeout), "HTTP request timeout in seconds")
	)

	flag.Parse()

	// 初始化日志系统
	logger.Init(*debug)

	// 尝试从多个位置加载.env文件
	envPaths := []string{
		".env",       // 当前目录
		"go/.env",    // go子目录
		"../go/.env", // 父目录的go子目录
		filepath.Join(os.Getenv("HOME"), "Projects/hacker-news/go/.env"), // 绝对路径
	}

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			logger.Info("加载环境变量文件: %s", path)
			err := godotenv.Load(path)
			if err != nil {
				logger.Error("加载环境变量文件出错 %s: %v", path, err)
			} else {
				break
			}
		}
	}

	// 创建配置
	cfg := &config.Config{
		OpenAIAPIKey:    *openAIKey,
		OpenAIBaseURL:   *openAIURL,
		OpenAIModel:     *openAIModel,
		OpenAIMaxTokens: *maxTokens,
		JinaAPIKey:      *jinaKey,
		MaxStories:      *maxStories,
		Development:     *dev,
		HTTPTimeout:     *httpTimeout,
	}

	// 设置默认值（如果未提供）
	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = defaultConfig.OpenAIModel
	}

	if cfg.OpenAIBaseURL == "" {
		cfg.OpenAIBaseURL = defaultConfig.OpenAIBaseURL
	}

	// 验证必需的字段
	if cfg.OpenAIAPIKey == "" {
		logger.Error("OpenAI API密钥是必需的。请在.env文件中设置或通过-openai-key标志提供")
		os.Exit(1)
	}

	// 创建输出目录（如果不存在）
	err := os.MkdirAll(*outputDir, 0755)
	if err != nil {
		logger.Error("创建输出目录失败: %v", err)
		os.Exit(1)
	}

	// 创建OpenAI客户端
	openAIClient := ai.NewOpenAIClient(
		cfg.OpenAIAPIKey,
		cfg.OpenAIBaseURL,
		cfg.OpenAIModel,
		cfg.OpenAIMaxTokens,
		cfg.HTTPTimeout,
	)

	// 开始工作流
	logger.Info("开始为 %s 生成Hacker News播客", *date)
	logger.Debug("使用OpenAI API URL: %s", cfg.OpenAIBaseURL)
	logger.Debug("使用OpenAI模型: %s", cfg.OpenAIModel)

	// 获取热门故事
	storyTimer := logger.NewTimer("获取热门故事")
	stories, err := fetcher.GetHackerNewsTopStories(*date, cfg.JinaAPIKey, cfg.HTTPTimeout)
	storyTimer.Stop()

	if err != nil {
		logger.Error("获取热门故事失败: %v", err)
		os.Exit(1)
	}

	if len(stories) == 0 {
		logger.Error("未找到任何故事")
		os.Exit(1)
	}

	// 限制故事数量
	if len(stories) > cfg.MaxStories {
		stories = stories[:cfg.MaxStories]
	}

	logger.Info("找到 %d 个故事", len(stories))
	logger.Debug("HTTP超时设置为 %d 秒", cfg.HTTPTimeout)

	// 格式化日期字符串，用于文件名
	dateFormatted := strings.ReplaceAll(*date, "-", "")

	// 处理每个故事
	var allStories []string
	var allStoriesMarkdown strings.Builder

	// 添加Markdown标题和日期
	allStoriesMarkdown.WriteString(fmt.Sprintf("# Hacker News 故事摘要 - %s\n\n", *date))

	for i, story := range stories {
		logger.Info("[%d/%d] 处理故事: %s", i+1, len(stories), story.Title)

		// 获取故事内容和评论
		contentTimer := logger.NewTimer("获取故事内容: " + story.Title)
		storyContent, err := fetcher.GetHackerNewsStory(story, cfg.OpenAIMaxTokens, cfg.JinaAPIKey, cfg.HTTPTimeout)
		contentTimer.Stop()

		if err != nil {
			logger.Error("获取故事 %s 失败: %v", story.ID, err)
			continue
		}

		logger.Debug("故事内容长度: %d 字符", len(storyContent))

		// 总结故事
		summarizeTimer := logger.NewTimer("总结故事: " + story.Title)
		summarizeStorySystemPrompt := prompt.SummarizeStoryPrompt()
		text, err := openAIClient.GenerateText(summarizeStorySystemPrompt, storyContent)
		summarizeTimer.Stop()

		if err != nil {
			logger.Error("总结故事 %s 失败: %v", story.ID, err)
			continue
		}

		logger.Info("成功总结故事: %s", story.Title)
		allStories = append(allStories, text)

		// 添加故事到Markdown文件
		allStoriesMarkdown.WriteString(fmt.Sprintf("## %s\n\n", story.Title))
		allStoriesMarkdown.WriteString(fmt.Sprintf("- 原文链接: [%s](%s)\n", story.Title, story.URL))
		allStoriesMarkdown.WriteString(fmt.Sprintf("- HN链接: [Hacker News讨论](%s)\n\n", story.HackerNewsURL))
		allStoriesMarkdown.WriteString(text + "\n\n---\n\n")

		// 在开发模式下让AI休息一下
		if cfg.Development {
			time.Sleep(2 * time.Second)
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	// 创建播客内容
	logger.Info("创建播客内容")
	podcastTimer := logger.NewTimer("生成播客内容")
	podcastSystemPrompt := prompt.SummarizePodcastPrompt(config.PodcastTitle, *date)
	podcastContent, err := openAIClient.GenerateText(podcastSystemPrompt, strings.Join(allStories, "\n\n---\n\n"))
	podcastTimer.Stop()

	if err != nil {
		logger.Error("创建播客内容失败: %v", err)
		os.Exit(1)
	}

	// 创建介绍内容
	logger.Info("创建简介内容")
	introTimer := logger.NewTimer("生成简介内容")
	introSystemPrompt := prompt.IntroPrompt()
	introContent, err := openAIClient.GenerateText(introSystemPrompt, podcastContent)
	introTimer.Stop()

	if err != nil {
		logger.Error("创建简介内容失败: %v", err)
		os.Exit(1)
	}

	// 将介绍内容插入到Markdown文件的标题下方
	markdownWithIntro := strings.Builder{}
	markdownWithIntro.WriteString(fmt.Sprintf("# Hacker News 故事摘要 - %s\n\n", *date))
	markdownWithIntro.WriteString("## 今日概述\n\n")
	markdownWithIntro.WriteString(introContent)
	markdownWithIntro.WriteString("\n\n---\n\n")
	markdownWithIntro.WriteString(allStoriesMarkdown.String()[len(fmt.Sprintf("# Hacker News 故事摘要 - %s\n\n", *date)):])

	// 保存内容到文件
	podcastFilename := filepath.Join(*outputDir, fmt.Sprintf("podcast-%s.txt", dateFormatted))
	markdownFilename := filepath.Join(*outputDir, fmt.Sprintf("stories-%s.md", dateFormatted))

	err = os.WriteFile(podcastFilename, []byte(podcastContent), 0644)
	if err != nil {
		logger.Error("写入播客内容失败: %v", err)
		os.Exit(1)
	}

	// 保存Markdown文件
	err = os.WriteFile(markdownFilename, []byte(markdownWithIntro.String()), 0644)
	if err != nil {
		logger.Error("写入故事摘要失败: %v", err)
		os.Exit(1)
	}

	logger.Info("成功为 %s 生成播客内容", *date)
	logger.Info("文件保存到:")
	logger.Info("  - 播客: %s", podcastFilename)
	logger.Info("  - 故事摘要: %s", markdownFilename)
}
