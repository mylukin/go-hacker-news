# Hacker News 播客生成器

这是一个Go语言编写的应用，可以自动获取Hacker News上的热门文章，使用OpenAI API生成摘要，并创建播客内容。

## 功能特点

- 获取指定日期的Hacker News热门故事
- 提取文章内容和评论
- 使用OpenAI生成简单易懂的故事摘要
- 生成两个文件：
  - 播客脚本（TXT格式）
  - 故事摘要（MD格式，包含今日概述）

## 技术栈

- Go语言
- OpenAI API
- HTML解析工具
- 环境变量加载器

## 运行环境要求

- Go 1.21或更高版本
- OpenAI API密钥
- Jina API密钥（可选，用于更好地获取网页内容）

## 安装方法

1. 克隆代码库：

```bash
git clone https://github.com/mylukin/go-hacker-news
cd go-hacker-news
```

2. 安装依赖：

```bash
go mod download
```

## 配置方法

在`go`目录下创建一个`.env`文件，包含以下内容：

```env
OPENAI_API_KEY=你的OpenAI密钥
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o
JINA_KEY=你的Jina密钥（可选）
HTTP_TIMEOUT=60  # HTTP请求超时时间（秒）
```

程序会在以下位置查找此文件：

- 当前目录（`.env`）
- `go/.env`
- `../go/.env`
- 你的主目录

## 默认设置

如果不做更改，程序会使用以下默认设置：

- OpenAI API地址：`https://api.openai.com/v1`
- OpenAI模型：`gpt-4o`
- 最大故事数：10
- 最大令牌数：16384
- HTTP请求超时：60秒

## 使用方法

编译并运行：

```bash
cd go
go build -o hnpodcast ./cmd/hnpodcast
./hnpodcast
```

也可以使用提供的Makefile简化操作（见下文）。

### 命令行选项

- `-date`：要获取故事的日期（YYYY-MM-DD格式，默认：今天）
- `-output`：输出文件保存位置（默认："output"）
- `-openai-key`：OpenAI密钥
- `-openai-url`：OpenAI API地址
- `-openai-model`：使用的AI模型
- `-jina-key`：Jina密钥
- `-max-stories`：获取的最大故事数（默认：10）
- `-max-tokens`：AI响应的最大令牌数（默认：16384）
- `-timeout`：HTTP请求超时时间（秒，默认：60）
- `-dev`：开发模式
- `-debug`：显示详细日志

示例：

```bash
./hnpodcast -date 2023-11-01 -output ./my_podcasts -max-stories 5
```

## 输出内容

程序会生成两个文件：

1. `podcast-YYYYMMDD.txt`：完整的播客脚本
2. `stories-YYYYMMDD.md`：包含今日概述和所有故事摘要的Markdown文件

## 使用Makefile

项目提供了Makefile来简化常见操作：

```bash
make build      # 编译程序
make run        # 运行程序（今天的故事）
make run-date   # 运行程序获取指定日期的故事（会提示输入日期）
make debug      # 以调试模式运行
make clean      # 清理编译文件
make test       # 运行测试
```

## 许可证

请查看LICENSE文件。
