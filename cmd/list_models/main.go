package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	// .envファイルを読み込む
	if err := godotenv.Load("../../.env"); err != nil {
		// カレントディレクトリも試す
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ .envファイルが見つかりません")
		}
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("🔍 利用可能なGeminiモデル一覧:")
	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s\n", m.Name)
	}
}
