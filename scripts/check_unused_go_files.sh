#!/bin/bash

# 未使用Goファイル検出スクリプト
echo "=== 未使用Goファイル検出 ==="

PROJECT_ROOT="/Users/taktakeu/Local/GitHub/roi-agent"

echo "🔍 Goファイルの依存関係をチェック中..."

# data-senderディレクトリ
echo ""
echo "📁 data-sender ディレクトリ:"
cd "$PROJECT_ROOT/data-sender"
echo "  go.mod で定義されているモジュール:"
head -1 go.mod

echo "  Goファイル一覧:"
for file in *.go; do
    if [ -f "$file" ]; then
        echo "    $file"
        # main関数やexport関数があるかチェック
        if grep -q "^func main(" "$file"; then
            echo "      → main関数あり"
        fi
        if grep -q "^func [A-Z]" "$file"; then
            echo "      → export関数あり"
        fi
    fi
done

# agentディレクトリ
echo ""
echo "📁 agent ディレクトリ:"
cd "$PROJECT_ROOT/agent"
echo "  go.mod で定義されているモジュール:"
head -1 go.mod

echo "  Goファイル一覧:"
for file in *.go; do
    if [ -f "$file" ]; then
        echo "    $file"
        # main関数やexport関数があるかチェック
        if grep -q "^func main(" "$file"; then
            echo "      → main関数あり"
        fi
        if grep -q "^func [A-Z]" "$file"; then
            echo "      → export関数あり"
        fi
    fi
done

# debugディレクトリ
echo ""
echo "📁 debug ディレクトリ:"
cd "$PROJECT_ROOT/debug"
if [ -f go.mod ]; then
    echo "  go.mod で定義されているモジュール:"
    head -1 go.mod
    
    echo "  Goファイル一覧:"
    for file in *.go; do
        if [ -f "$file" ]; then
            echo "    $file"
            if grep -q "^func main(" "$file"; then
                echo "      → main関数あり"
            fi
        fi
    done
else
    echo "  go.mod なし"
fi

# windowsディレクトリ
echo ""
echo "📁 windows ディレクトリ:"
cd "$PROJECT_ROOT/windows"
if [ -f go.mod ]; then
    echo "  go.mod で定義されているモジュール:"
    head -1 go.mod
    
    echo "  Goファイル一覧:"
    for file in *.go; do
        if [ -f "$file" ]; then
            echo "    $file"
            if grep -q "^func main(" "$file"; then
                echo "      → main関数あり"
            fi
        fi
    done
else
    echo "  go.mod なし"
fi

echo ""
echo "🧐 推奨アクション:"
echo "  - 各ディレクトリで 'go mod tidy' を実行して未使用依存関係を削除"
echo "  - main関数のないファイルは統合を検討"
echo "  - backup拡張子のファイルは削除対象"
