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

# 计算前一天的日期，格式为YYYY-MM-DD
YESTERDAY=$(date -d "yesterday" +"%Y-%m-%d")

# 运行hnpodcast，使用前一天的日期
./dist/hnpodcast -date "$YESTERDAY"

# 执行make commit并将输出重定向到日志文件
make commit > /var/log/hacker-news.log 2>&1

