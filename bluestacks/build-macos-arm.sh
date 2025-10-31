#!/bin/bash

# 编译 macOS ARM64 架构的二进制文件
# 使用方法: ./build-macos-arm.sh

set -e

echo "🔨 开始编译 macOS ARM64 二进制文件..."

# 设置输出文件名
OUTPUT_NAME="bluestacks-monitor"

# 设置目标平台和架构
export GOOS=darwin
export GOARCH=arm64

# 编译
echo "📦 目标平台: ${GOOS}"
echo "🏗️  目标架构: ${GOARCH}"
echo "📝 输出文件: ${OUTPUT_NAME}"

go build -o "${OUTPUT_NAME}" -ldflags="-s -w" .

# 检查编译结果
if [ -f "${OUTPUT_NAME}" ]; then
    echo "✅ 编译成功！"
    echo "📊 文件信息:"
    ls -lh "${OUTPUT_NAME}"
    echo ""
    echo "🎯 可以使用以下命令运行:"
    echo "   ./${OUTPUT_NAME}"
else
    echo "❌ 编译失败！"
    exit 1
fi

