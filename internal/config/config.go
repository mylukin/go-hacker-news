package config

// PodcastTitle is the title of the podcast
const PodcastTitle = "Hacker News Daily Podcast"

// PodcastDescription is the description of the podcast
const PodcastDescription = "An AI-generated daily podcast that summarizes popular articles from Hacker News for software developers and tech enthusiasts."

// KeepDays is the number of days to keep podcast content
const KeepDays = 30

// DefaultOpenAIBaseURL is the default URL for OpenAI API
const DefaultOpenAIBaseURL = "https://api.openai.com/v1"

// DefaultHTTPTimeout is the default timeout for HTTP requests in seconds
const DefaultHTTPTimeout = 60

// Config holds application configuration
type Config struct {
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	OpenAIModel     string
	OpenAIMaxTokens int
	JinaAPIKey      string
	MaxStories      int
	Development     bool
	HTTPTimeout     int // HTTP请求超时时间（秒）
}

// NewDefaultConfig creates a new config with default values
func NewDefaultConfig() *Config {
	return &Config{
		OpenAIBaseURL:   DefaultOpenAIBaseURL,
		OpenAIModel:     "gpt-4o",
		OpenAIMaxTokens: 16384,
		MaxStories:      10,
		Development:     false,
		HTTPTimeout:     DefaultHTTPTimeout,
	}
}
