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

// 生成标题，使用AI对播客内容进行总结
func generateTitle(date string, openAIClient *ai.OpenAIClient, podcastContent string) string {
	// 从YYYY-MM-DD格式转换为年.月.日格式
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		logger.Error("解析日期失败: %v", err)
		return fmt.Sprintf("%s HN精选：科技热点回顾", date)
	}

	// 格式化为年.月.日作为前缀
	datePrefix := fmt.Sprintf("%d.%d.%d", t.Year(), t.Month(), t.Day())

	// 使用AI生成标题(英文提示词)
	titleSystemPrompt := "You are a title generation expert. Create a concise and engaging Chinese title for the following podcast content. The title should be no more than 15 Chinese characters, should not include dates, and should avoid generic words like '总结' (summary) or '回顾' (review). Focus on the key tech topics or trends mentioned in the content."
	aiTitle, err := openAIClient.GenerateText(titleSystemPrompt, podcastContent)
	if err != nil {
		logger.Error("生成AI标题失败: %v", err)
		return fmt.Sprintf("%s HN精选：科技热点探讨", datePrefix)
	}

	// 清理生成的标题（去除可能的引号、多余空格、标点等）
	aiTitle = strings.TrimSpace(aiTitle)
	aiTitle = strings.Trim(aiTitle, "\"'.,，。")

	// 如果AI生成的标题为空或过长，使用默认标题
	if aiTitle == "" || len([]rune(aiTitle)) > 20 {
		// 为Hacker News播客生成更合适的默认标题
		defaultTitles := []string{
			"HN精选：科技创新前沿",
			"科技热点：程序员视角",
			"硅谷观察：技术与创业",
			"开发者热议：科技动态",
			"技术前沿：HN每日精选",
		}

		// 基于日期选择一个默认标题，确保每天都有所不同
		defaultTitle := defaultTitles[t.Day()%len(defaultTitles)]
		return fmt.Sprintf("%s %s", datePrefix, defaultTitle)
	}

	// 组合日期和AI生成的标题
	return fmt.Sprintf("%s %s", datePrefix, aiTitle)
}

func main() {
	// 获取默认配置
	defaultConfig := config.NewDefaultConfig()

	// 定义所有命令行参数，但暂不解析
	debug := flag.Bool("debug", false, "Enable debug logging")
	date := flag.String("date", time.Now().Format("2006-01-02"), "Date to fetch stories for (YYYY-MM-DD)")
	outputDir := flag.String("output", "output", "Directory to store output files")
	openAIKey := flag.String("openai-key", "", "OpenAI API key")
	openAIURL := flag.String("openai-url", "", "OpenAI API base URL")
	openAIModel := flag.String("openai-model", "", "OpenAI model to use")
	jinaKey := flag.String("jina-key", "", "Jina API key")
	maxStories := flag.Int("max-stories", defaultConfig.MaxStories, "Maximum number of stories to process")
	maxTokens := flag.Int("max-tokens", defaultConfig.OpenAIMaxTokens, "Maximum number of tokens for OpenAI responses")
	dev := flag.Bool("dev", defaultConfig.Development, "Run in development mode")
	httpTimeout := flag.Int("timeout", defaultConfig.HTTPTimeout, "HTTP request timeout in seconds")

	// 先解析所有命令行参数
	flag.Parse()

	// 初始化日志系统
	logger.Init(*debug)

	// 获取并显示当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		logger.Error("获取当前工作目录失败: %v", err)
	} else {
		logger.Info("当前工作目录: %s", workDir)
	}

	// 环境变量文件路径
	envPaths := []string{
		".env",       // 当前目录
		"go/.env",    // go子目录
		"../go/.env", // 父目录的go子目录
	}

	// 环境变量加载
	envLoaded := false
	for _, path := range envPaths {
		absPath, _ := filepath.Abs(path)
		logger.Debug("检查环境变量文件: %s (绝对路径: %s)", path, absPath)

		if _, err := os.Stat(path); err == nil {
			logger.Info("找到环境变量文件: %s", path)

			// 检查文件权限
			fileInfo, statErr := os.Stat(path)
			if statErr != nil {
				logger.Error("获取文件信息失败 %s: %v", path, statErr)
			} else {
				fileMode := fileInfo.Mode()
				logger.Debug("文件权限: %s (%o)", fileMode, fileMode.Perm())

				// 检查是否可读
				if fileMode.Perm()&0400 == 0 {
					logger.Warn("文件无读取权限: %s", path)
				}
			}

			err = godotenv.Load(path)
			if err != nil {
				logger.Error("加载环境变量文件出错 %s: %v", path, err)
			} else {
				logger.Info("成功加载环境变量文件: %s", path)
				envLoaded = true

				// 检查关键环境变量是否存在
				if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
					if len(apiKey) > 4 {
						logger.Debug("环境变量OPENAI_API_KEY已加载: %s", apiKey[:4]+"...")
					} else {
						logger.Debug("环境变量OPENAI_API_KEY已加载(值过短)")
					}
				} else {
					logger.Warn("环境变量OPENAI_API_KEY未设置或为空")
				}
				break
			}
		} else {
			logger.Debug("环境变量文件不存在: %s (%v)", path, err)
		}
	}

	if !envLoaded {
		logger.Warn("未成功加载任何环境变量文件，将使用系统环境变量")
	}

	// 显示所有可能与OpenAI API相关的环境变量
	checkEnvVars := []string{
		"OPENAI_API_KEY",
		"OPENAI_KEY",
		"OPENAI_SECRET_KEY",
		"OPENAI_BASE_URL",
		"OPENAI_MODEL",
		"HTTP_TIMEOUT",
	}

	logger.Debug("检查系统环境变量:")
	for _, varName := range checkEnvVars {
		val := os.Getenv(varName)
		if val != "" {
			// 不显示密钥的具体内容
			if strings.Contains(strings.ToLower(varName), "key") || strings.Contains(strings.ToLower(varName), "secret") {
				logger.Debug("- %s: ***已设置***", varName)
			} else {
				logger.Debug("- %s: %s", varName, val)
			}
		} else {
			logger.Debug("- %s: 未设置", varName)
		}
	}

	// 使用从环境变量获取的值更新命令行参数的默认值
	if *openAIKey == "" {
		*openAIKey = os.Getenv("OPENAI_API_KEY")
	}
	if *openAIURL == "" {
		*openAIURL = os.Getenv("OPENAI_BASE_URL")
	}
	if *openAIModel == "" {
		*openAIModel = os.Getenv("OPENAI_MODEL")
	}
	if *jinaKey == "" {
		*jinaKey = os.Getenv("JINA_KEY")
	}
	if envTimeout := getEnvInt("HTTP_TIMEOUT", defaultConfig.HTTPTimeout); *httpTimeout == defaultConfig.HTTPTimeout {
		*httpTimeout = envTimeout
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

	// 记录OpenAI API密钥来源
	if *openAIKey == "" {
		logger.Debug("命令行未提供OpenAI API密钥")
	} else if *openAIKey == os.Getenv("OPENAI_API_KEY") {
		logger.Debug("使用环境变量中的OpenAI API密钥")
	} else {
		logger.Debug("使用命令行参数提供的OpenAI API密钥")
	}

	// 尝试直接从环境变量获取密钥（以防flag解析发生在env加载之前）
	if cfg.OpenAIAPIKey == "" {
		directKey := os.Getenv("OPENAI_API_KEY")
		if directKey != "" {
			logger.Info("从环境变量直接加载OpenAI API密钥")
			cfg.OpenAIAPIKey = directKey
		}
	}

	// 验证必需的字段
	if cfg.OpenAIAPIKey == "" {
		logger.Error("OpenAI API密钥是必需的。请在.env文件中设置或通过-openai-key标志提供")
		os.Exit(1)
	}

	// 创建输出目录（如果不存在）
	err = os.MkdirAll(*outputDir, 0755)
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
	titleFilename := filepath.Join(*outputDir, fmt.Sprintf("title-%s.txt", dateFormatted))

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

	// 生成并保存标题
	title := generateTitle(*date, openAIClient, podcastContent)
	err = os.WriteFile(titleFilename, []byte(title), 0644)
	if err != nil {
		logger.Error("写入标题失败: %v", err)
		os.Exit(1)
	}

	logger.Info("成功为 %s 生成播客内容", *date)
	logger.Info("文件保存到:")
	logger.Info("  - 播客: %s", podcastFilename)
	logger.Info("  - 故事摘要: %s", markdownFilename)
	logger.Info("  - 标题: %s", titleFilename)
}
