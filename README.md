# Hacker News Podcast Maker

This Go program gets top stories from Hacker News and makes podcasts from them using OpenAI.

## What It Does

- Gets top stories from Hacker News
- Gets article content and comments
- Uses OpenAI to make short summaries
- Makes three files:
  - Podcast script (TXT)
  - Short intro (TXT)
  - Blog post (MD)

## Tools We Use

- Go language
- OpenAI API
- HTML parsing tools
- Env file loader

## What You Need

- Go 1.21 or newer
- OpenAI API key
- Jina API key (optional, helps get web content)

## How to Install

1. Get the code:

```bash
git clone https://github.com/yourusername/hacker-news-podcast
cd hacker-news-podcast/go
```

2. Get the tools:

```bash
go mod download
```

## Setup

Make a `.env` file in the `go` folder with:

```env
OPENAI_API_KEY=your_openai_api_key
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o
JINA_KEY=your_jina_api_key  # Optional
```

The program will look for this file in:

- Current folder (`.env`)
- `go/.env`
- `../go/.env`
- Your home folder

## Default Settings

If you don't change them, the program uses:

- OpenAI API URL: `https://api.openai.com/v1`
- OpenAI Model: `gpt-4o`
- Max stories: 10
- Max tokens: 16384

## How to Use

Build and run:

```bash
cd go
go build -o hnpodcast ./cmd/hnpodcast
./hnpodcast
```

### Options

- `-date`: Which day to get stories for (YYYY-MM-DD, default: today)
- `-output`: Where to save files (default: "output")
- `-openai-key`: Your OpenAI key
- `-openai-url`: OpenAI API address
- `-openai-model`: Which AI model to use
- `-jina-key`: Your Jina key
- `-max-stories`: How many stories to get (default: 10)
- `-max-tokens`: Max tokens for AI responses (default: 16384)
- `-dev`: Run in test mode
- `-debug`: Show more detailed logs

Example:

```bash
./hnpodcast -date 2023-11-01 -output ./my_podcasts -max-stories 5
```

## What You Get

The program makes three files:

1. `podcast-YYYYMMDD.txt`: The podcast script
2. `intro-YYYYMMDD.txt`: A short intro
3. `blog-YYYYMMDD.md`: A blog post

## License

See the LICENSE file.
