#!/bin/bash
# 加载用户环境变量，防止crontab执行时的环境变量问题
if [ -f ~/.bashrc ]; then
    source ~/.bashrc
elif [ -f ~/.bash_profile ]; then
    source ~/.bash_profile
elif [ -f ~/.profile ]; then
    source ~/.profile
fi

# 切换到脚本所在目录
cd "$(dirname "$0")"

# 加载.env文件
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# 检查是否提供了日期参数，未提供则使用昨天的日期
if [ -z "$1" ]; then
    TARGET_DATE=$(date -d "yesterday" +"%Y-%m-%d")
    echo "未提供日期参数，使用昨天的日期: $TARGET_DATE"
else
    # 验证日期格式是否为YYYY-MM-DD
    if [[ $1 =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
        TARGET_DATE=$1
        echo "使用指定日期: $TARGET_DATE"
    else
        echo "错误: 日期格式不正确，请使用YYYY-MM-DD格式"
        exit 1
    fi
fi

# 运行hnpodcast，使用目标日期
./dist/hnpodcast -date "$TARGET_DATE"

# 获取格式化后的日期字符串用于文件名
DATE_FORMATTED=$(date -d "$TARGET_DATE" +"%Y%m%d")

# 从title文件读取标题
TITLE=$(cat "output/title-${DATE_FORMATTED}.txt")
if [ -z "$TITLE" ]; then
    TITLE="${TARGET_DATE} HN精选：科技热点探讨"
fi

# 读取DICTOGO_API_TOKEN从.env文件
DICTOGO_API_TOKEN=${DICTOGO_API_TOKEN:-$OPENAI_API_KEY}

# 读取播客内容并转义双引号
PODCAST_CONTENT=$(cat "output/podcast-${DATE_FORMATTED}.txt" | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g')

# 检查播客内容是否为空
if [ -z "$PODCAST_CONTENT" ] || [ "$PODCAST_CONTENT" = "" ]; then
    echo "错误: 播客内容为空，无法提交到dictogo.app"
    exit 1
fi

# 发送到dictogo.app API
echo "正在发送播客内容到dictogo.app API..."
curl -s -X POST "https://api.dictogo.app/openapi/v1/article/create" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${DICTOGO_API_TOKEN}" \
-d @- << EOF
{
  "title": "${TITLE}",
  "content": "${PODCAST_CONTENT}",
  "toLanguage": "zh-hans",
  "albumId": "67f0d7d531701d15ea5397ad"
}
EOF
echo -e "\n文章已发送到dictogo.app"

# 执行make commit并将输出重定向到日志文件
make commit
