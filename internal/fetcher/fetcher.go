package fetcher

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/lukin/hacker-news/go/internal/logger"
)

// Story represents a Hacker News story
type Story struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	HackerNewsURL string `json:"hackerNewsUrl"`
}

// GetHackerNewsTopStories fetches the top stories from Hacker News for a given date
func GetHackerNewsTopStories(today string, jinaKey string, timeoutSeconds int) ([]Story, error) {
	// url := fmt.Sprintf("https://news.ycombinator.com/front?day=%s", today)
	url := "https://news.ycombinator.com/news"

	logger.Info("获取 %s 的热门故事，来源: %s", today, url)

	// Create client with timeout
	client := req.C().SetTimeout(time.Duration(timeoutSeconds) * time.Second)

	// Setup request
	request := client.R().
		SetHeader("X-Retain-Images", "none").
		SetHeader("X-Return-Format", "html")

	if jinaKey != "" {
		request.SetHeader("Authorization", fmt.Sprintf("Bearer %s", jinaKey))
	}

	// Send request
	resp, err := request.Get("https://r.jina.ai/" + url)
	if err != nil {
		return nil, fmt.Errorf("获取故事失败: %w", err)
	}

	logger.Debug("热门故事请求结果: %s", resp.Status)

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("意外的状态码: %d", resp.StatusCode)
	}

	// Parse HTML
	html, err := resp.ToString()
	if err != nil {
		return nil, fmt.Errorf("获取响应内容失败: %w", err)
	}

	// Debug log - truncate to reasonable length
	logHtml := html
	if len(logHtml) > 500 {
		logHtml = logHtml[:500] + "..."
	}
	logger.Debug("收到HTML内容(截断): %s", logHtml)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	var stories []Story

	// Find all story items
	doc.Find(".athing.submission").Each(func(i int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if !exists {
			return
		}

		titleElem := s.Find(".titleline > a")
		title := titleElem.Text()
		url, _ := titleElem.Attr("href")

		story := Story{
			ID:            id,
			Title:         title,
			URL:           url,
			HackerNewsURL: fmt.Sprintf("https://news.ycombinator.com/item?id=%s", id),
		}

		stories = append(stories, story)
	})

	// Filter out invalid stories
	var validStories []Story
	for _, story := range stories {
		if story.ID != "" && story.URL != "" {
			validStories = append(validStories, story)
		}
	}

	return validStories, nil
}

// CleanHTML removes problematic HTML tags and content
func CleanHTML(html string) string {
	// 使用regexp包进行正则表达式匹配
	scriptRegex := regexp.MustCompile(`(?s)<script.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	styleRegex := regexp.MustCompile(`(?s)<style.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	commentRegex := regexp.MustCompile(`(?s)<!--.*?-->`)
	html = commentRegex.ReplaceAllString(html, "")

	// 如果内容太长，截断它
	maxLen := 100000 // 设置一个合理的最大长度
	if len(html) > maxLen {
		logger.Debug("HTML内容过长，从 %d 字符截断至 %d 字符", len(html), maxLen)
		html = html[:maxLen]
	}

	return html
}

// GetHackerNewsStory fetches the content of a story and its comments
func GetHackerNewsStory(story Story, maxTokens int, jinaKey string, timeoutSeconds int) (string, error) {
	// Create client with timeout
	client := req.C().SetTimeout(time.Duration(timeoutSeconds) * time.Second)

	// Create request with common headers
	reqHeaders := map[string]string{
		"X-Retain-Images": "none",
	}

	if jinaKey != "" {
		reqHeaders["Authorization"] = fmt.Sprintf("Bearer %s", jinaKey)
	}

	// Use goroutines and channels for concurrent requests
	type result struct {
		content string
		err     error
	}

	articleChan := make(chan result)
	commentsChan := make(chan result)

	// Fetch article
	go func() {
		articleReq := client.R().
			SetHeaders(reqHeaders)

		articleTimer := logger.NewTimer("获取文章: " + story.Title)
		resp, err := articleReq.Get("https://r.jina.ai/" + story.URL)
		articleTimer.Stop()

		if err != nil {
			articleChan <- result{"", fmt.Errorf("获取文章失败: %w", err)}
			return
		}

		if !resp.IsSuccess() {
			articleChan <- result{"", nil}
			return
		}

		content, err := resp.ToString()
		if err != nil {
			articleChan <- result{"", fmt.Errorf("读取文章内容失败: %w", err)}
			return
		}

		// 清理HTML内容
		content = CleanHTML(content)

		// 记录内容前100个字符用于调试
		if len(content) > 0 {
			preview := content
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			logger.Debug("文章内容预览: %s", preview)
		}

		articleChan <- result{content, nil}
	}()

	// Fetch comments
	go func() {
		commentsReq := client.R().
			SetHeaders(reqHeaders).
			SetHeader("X-Remove-Selector", ".navs").
			SetHeader("X-Target-Selector", "#pagespace + tr")

		commentsURL := fmt.Sprintf("https://r.jina.ai/https://news.ycombinator.com/item?id=%s", story.ID)

		commentsTimer := logger.NewTimer("获取评论: " + story.Title)
		resp, err := commentsReq.Get(commentsURL)
		commentsTimer.Stop()

		if err != nil {
			commentsChan <- result{"", fmt.Errorf("获取评论失败: %w", err)}
			return
		}

		if !resp.IsSuccess() {
			commentsChan <- result{"", nil}
			return
		}

		content, err := resp.ToString()
		if err != nil {
			commentsChan <- result{"", fmt.Errorf("读取评论内容失败: %w", err)}
			return
		}

		// 清理HTML内容
		content = CleanHTML(content)

		// 记录内容前100个字符用于调试
		if len(content) > 0 {
			preview := content
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			logger.Debug("评论内容预览: %s", preview)
		}

		commentsChan <- result{content, nil}
	}()

	// Wait for results
	articleResult := <-articleChan
	if articleResult.err != nil {
		return "", articleResult.err
	}

	commentsResult := <-commentsChan
	if commentsResult.err != nil {
		return "", commentsResult.err
	}

	// Format the final response
	var parts []string

	if story.Title != "" {
		parts = append(parts, fmt.Sprintf("<title>\n%s\n</title>", story.Title))
	}

	if articleResult.content != "" {
		// 对于GPT-4o，我们可以处理更多内容，将乘数从4增加到10
		article := articleResult.content
		if len(article) > maxTokens*10 {
			logger.Debug("文章内容过长，从 %d 字符截断至 %d 字符", len(article), maxTokens*10)
			article = article[:maxTokens*10]
		}
		parts = append(parts, fmt.Sprintf("<article>\n%s\n</article>", article))
	}

	if commentsResult.content != "" {
		// 对于GPT-4o，我们可以处理更多内容，将乘数从4增加到10
		comments := commentsResult.content
		if len(comments) > maxTokens*10 {
			logger.Debug("评论内容过长，从 %d 字符截断至 %d 字符", len(comments), maxTokens*10)
			comments = comments[:maxTokens*10]
		}
		parts = append(parts, fmt.Sprintf("<comments>\n%s\n</comments>", comments))
	}

	// 构建完整内容
	finalContent := strings.Join(parts, "\n\n---\n\n")

	// 记录最终内容的长度和前100个字符
	logger.Debug("最终内容长度: %d 字符", len(finalContent))
	preview := finalContent
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}
	logger.Debug("最终内容预览: %s", preview)

	return finalContent, nil
}
