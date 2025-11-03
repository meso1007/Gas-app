.PHONY: deps fetch fetch-scrape fetch-exchange fetch-all list list-exchange latest latest-exchange clean-db

deps:
	@echo "📦 依存パッケージをインストール中..."
	go mod download
	go mod tidy
	@echo "✅ 完了"

fetch:
	@echo "⛽ ガソリン価格を取得（モックモード）..."
	go run cmd/local/main.go -mode=fetch

fetch-scrape:
	@echo "🌐 ガソリン価格を取得（スクレイピングモード）..."
	go run cmd/local/main.go -mode=fetch -scrape=true -mock=true

fetch-exchange:
	@echo "💱 為替レートを取得..."
	go run cmd/local/main.go -mode=fetch-exchange -mock=false

fetch-all:
	@echo "🚀 全データを取得（ガソリン価格 + 為替レート）..."
	go run cmd/local/main.go -mode=fetch-all -scrape=true -mock=false

list:
	@echo "📋 ガソリン価格一覧を表示..."
	go run cmd/local/main.go -mode=list

list-exchange:
	@echo "💱 為替レート一覧を表示..."
	go run cmd/local/main.go -mode=list-exchange

latest:
	@echo "🔍 最新のガソリン価格を表示..."
	go run cmd/local/main.go -mode=latest

latest-exchange:
	@echo "💱 最新の為替レートを表示..."
	go run cmd/local/main.go -mode=latest-exchange

clean-db:
	@echo "🗑️  データベースを削除..."
	rm -f data/gasinsight.db
	@echo "✅ 完了"

help:
	@echo "利用可能なコマンド:"
	@echo "  make deps            - 依存パッケージをインストール"
	@echo "  make fetch           - ガソリン価格を取得（モック）"
	@echo "  make fetch-scrape    - ガソリン価格を取得（スクレイピング）"
	@echo "  make fetch-exchange  - 為替レートを取得"
	@echo "  make fetch-all       - 全データを取得"
	@echo "  make list            - ガソリン価格一覧"
	@echo "  make list-exchange   - 為替レート一覧"
	@echo "  make latest          - 最新ガソリン価格"
	@echo "  make latest-exchange - 最新為替レート"
	@echo "  make clean-db        - データベースを削除"
