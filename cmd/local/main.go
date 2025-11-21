package main

import (
	"context"
	"flag"
	"fmt"
	"gasinsight/internal/database"
	"gasinsight/internal/detect"
	services "gasinsight/internal/detect"
	fetcher "gasinsight/internal/fetch"
	model "gasinsight/internal/model"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む（存在しない場合はスキップ）
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .envファイルが見つかりません。環境変数を直接使用します。")
	}

	mode := flag.String("mode", "fetch", "モード")
	dbPath := flag.String("db", "./data/gasinsight.db", "DBパス")
	useScraping := flag.Bool("scrape", false, "スクレイピングを使用")
	useMock := flag.Bool("mock", true, "モック使用")
	useMockAnalysis := flag.Bool("mock-analysis", true, "モック分析を使用（Gemini APIの代わり）")
	detectChange := flag.Bool("detect", true, "変動検知を有効化")
	mockDate := flag.String("mock-date", "", "モックデータの日付 (例: 2025-11-06)")

	flag.Parse()

	log.Println("🚀 GasInsight ローカル実行版")

	db, err := database.NewSQLiteClient(*dbPath)
	if err != nil {
		log.Fatalf("❌ エラー: %v", err)
	}
	defer db.Close()

	switch *mode {
	case "fetch":
		fetchGasPrice(db, *useScraping, *useMock, *detectChange, *mockDate)
	case "fetch-exchange":
		fetchExchangeRate(db, *useMock, *detectChange)
	case "fetch-all":
		fetchGasPrice(db, *useScraping, *useMock, *detectChange, *mockDate)
		fetchExchangeRate(db, *useMock, *detectChange)
	case "list":
		listGasPrices(db)
	case "list-exchange":
		listExchangeRates(db)
	case "latest":
		latestGasPrice(db)
	case "latest-exchange":
		latestExchangeRate(db)
	case "fetch-news":
		fetchNews(db, *useMock, *useMockAnalysis)
	case "list-news":
		listNews(db)
	case "latest-news":
		latestNews(db)
	case "analyze-fluctuation":
		analyzeFluctuation(db)
	default:
		log.Fatalf("❌ 不正なモード: %s", *mode)
	}

	log.Println("✅ 処理完了")
}

func fetchGasPrice(db *database.SQLiteClient, useScraping bool, useMock bool, detectChange bool, mockDate string) {
	log.Println("⛽ ガソリン価格を取得中...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var data *fetcher.GasPriceData
	var err error

	if useScraping {
		manager := fetcher.NewScraperManager()
		data, err = manager.ScrapeWithFallback(ctx, useMock)
	} else {
		f := fetcher.NewMockGasPriceFetcher()
		data, err = f.FetchLatestPrice()
	}

	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}

	if mockDate != "" {
		data.Date = mockDate
	}

	price := model.NewGasPrice(data.Date, data.Region,
		data.RegularPrice, data.PremiumPrice, data.DieselPrice)

	if err := db.SaveGasPrice(price); err != nil {
		log.Fatalf("❌ 保存エラー: %v", err)
	}

	printGasPrice(price)

	if detectChange {
		if err := services.DetectPriceChanges("./data/gasinsight.db", 2.0); err != nil {
			log.Printf("⚠️  ガソリン価格変動検知エラー: %v", err)
		}
	}
}

func fetchNews(db *database.SQLiteClient, useMockNews bool, useMockAnalysis bool) {
	log.Println("📰 ニュース取得中...")

	var articles []fetcher.NewsArticle
	var err error

	if useMockNews {
		// モックニュースを使用
		mockFetcher := fetcher.NewMockNewsFetcher()
		articles, err = mockFetcher.FetchTopNews("")
	} else {
		// 実際のNewsAPIを使用
		apiKey := os.Getenv("NEWSAPI_KEY")
		if apiKey == "" {
			log.Println("⚠️  NEWSAPI_KEYが設定されていません。環境変数を確認してください。")
			log.Println("💡 ヒント: .envファイルに NEWSAPI_KEY=your_key を追加してください")
			log.Println("💡 取得先: https://newsapi.org/register")
			return
		}

		log.Printf("🔑 APIキー: %s...%s (長さ: %d)", apiKey[:4], apiKey[len(apiKey)-4:], len(apiKey))

		newsFetcher := fetcher.NewNewsFetcher(apiKey)

		// 英語のクエリを使用（NewsAPIは英語の方が安定）
		articles, err = newsFetcher.FetchTopNews("oil OR gasoline OR economy")
	}

	if err != nil {
		log.Printf("❌ ニュース取得エラー: %v", err)
		if !useMockNews {
			log.Println("💡 ヒント:")
			log.Println("  1. NewsAPIキーが正しいか確認")
			log.Println("  2. https://newsapi.org/account でAPIキーの状態を確認")
			log.Println("  3. 無料プランは過去1ヶ月のニュースのみ取得可能")
			log.Println("  4. モックモードで試す: make fetch-news (デフォルトでモック使用)")
		}
		return
	}

	if len(articles) == 0 {
		log.Println("📭 ニュースが見つかりませんでした")
		return
	}

	log.Printf("📊 %d件のニュースを取得しました", len(articles))
	successCount := 0

	for i, a := range articles {
		log.Printf("[%d/%d] 分析中: %s", i+1, len(articles), a.Title)

		var analyzed *detect.AnalyzedNews
		var err error

		if useMockAnalysis {
			// モック分析を使用（API不要）
			analyzed, err = detect.MockAnalyzeNews(a)
		} else {
			// Gemini APIで分析
			// レート制限回避のため、前のリクエストから時間を空ける（無料枠対策）
			if i > 0 {
				log.Printf("⏳ APIレート制限回避のため 10秒待機中...")
				time.Sleep(10 * time.Second)
			}

			analyzed, err = detect.AnalyzeNewsWithGemini(a)
		}

		if err != nil {
			log.Printf("⚠️  分析エラー: %v", err)
			// 429エラーの場合は長めに待機してリトライを促すなどの処理が可能だが、
			// ここでは単純に次の記事へ進む
			continue
		}

		if err := db.SaveNews(analyzed); err != nil {
			log.Printf("⚠️  保存エラー: %v", err)
			continue
		}

		log.Printf("✅ 保存: %s (%s)", analyzed.Title, analyzed.Sentiment)
		successCount++
	}

	log.Printf("🎉 完了: %d/%d 件のニュースを保存しました", successCount, len(articles))
}

func fetchExchangeRate(db *database.SQLiteClient, useMock bool, detectChange bool) {
	log.Println("💱 為替レートを取得中...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var data *fetcher.ExchangeRateData
	var err error

	if useMock {
		f := fetcher.NewMockExchangeRateFetcher()
		data, err = f.Fetch(ctx)
	} else {
		f := fetcher.NewExchangeRateFetcher()
		data, err = f.Fetch(ctx)
	}

	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}

	rate := model.NewExchangeRate(data.Date, data.USDJPY, data.EURJPY, data.GBPJPY, data.CNYJPY)

	if err := db.SaveExchangeRate(rate); err != nil {
		log.Fatalf("❌ 保存エラー: %v", err)
	}

	printExchangeRate(rate)

	// --- 変動検知 ---
	if detectChange {
		if err := services.DetectPriceChanges("./data/gasinsight.db", 2.0); err != nil {
			log.Printf("⚠️  ガソリン価格変動検知エラー: %v", err)
		}
	}

}

func listGasPrices(db *database.SQLiteClient) {
	prices, err := db.GetAllGasPrices()
	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}

	if len(prices) == 0 {
		fmt.Println("📭 データがありません")
		return
	}

	fmt.Printf("\n📊 ガソリン価格データ一覧（%d件）\n\n", len(prices))
	for i, p := range prices {
		fmt.Printf("[%d] %s - レギュラー:%.2f円 ハイオク:%.2f円 軽油:%.2f円 (%s)\n",
			i+1, p.Date, p.RegularPrice, p.PremiumPrice, p.DieselPrice, p.Region)
	}
}

func listExchangeRates(db *database.SQLiteClient) {
	rates, err := db.GetAllExchangeRates()
	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}

	if len(rates) == 0 {
		fmt.Println("📭 データがありません")
		return
	}

	fmt.Printf("\n💱 為替レートデータ一覧(%d件) \n\n", len(rates))
	for i, r := range rates {
		fmt.Printf("[%d] %s - USD:%.2f EUR:%.2f GBP:%.2f CNY:%.2f\n",
			i+1, r.Date, r.USDJPY, r.EURJPY, r.GBPJPY, r.CNYJPY)
	}
}

func latestGasPrice(db *database.SQLiteClient) {
	p, err := db.GetLatestGasPrice()
	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}
	printGasPrice(p)
}

func latestExchangeRate(db *database.SQLiteClient) {
	r, err := db.GetLatestExchangeRate()
	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}
	printExchangeRate(r)
}

func printGasPrice(p *model.GasPrice) {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⛽ ガソリン価格")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("日付:       %s\n", p.Date)
	fmt.Printf("地域:       %s\n", p.Region)
	fmt.Printf("レギュラー: %.2f円\n", p.RegularPrice)
	fmt.Printf("ハイオク:   %.2f円\n", p.PremiumPrice)
	fmt.Printf("軽油:       %.2f円\n", p.DieselPrice)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
}

func printExchangeRate(r *model.ExchangeRate) {
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💱 為替レート")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("日付:    %s\n", r.Date)
	fmt.Printf("USD/JPY: %.2f円\n", r.USDJPY)
	fmt.Printf("EUR/JPY: %.2f円\n", r.EURJPY)
	fmt.Printf("GBP/JPY: %.2f円\n", r.GBPJPY)
	fmt.Printf("CNY/JPY: %.2f円\n", r.CNYJPY)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
}

func listNews(db *database.SQLiteClient) {
	newsList, err := db.GetAllNews()
	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}

	if len(newsList) == 0 {
		fmt.Println("📭 ニュースがありません")
		return
	}

	fmt.Printf("\n📰 ニュース一覧（%d件）\n\n", len(newsList))
	for i, n := range newsList {
		fmt.Printf("[%d] %s\n", i+1, n.Title)
		fmt.Printf("    日付: %s | 感情: %s\n", n.Date, n.Sentiment)
		fmt.Printf("    要約: %s\n", truncateString(n.Summary, 100))
		fmt.Printf("    URL:  %s\n\n", n.URL)
	}
}

func latestNews(db *database.SQLiteClient) {
	newsList, err := db.GetLatestNews(5)
	if err != nil {
		log.Fatalf("❌ 取得エラー: %v", err)
	}

	if len(newsList) == 0 {
		fmt.Println("📭 ニュースがありません")
		return
	}

	fmt.Printf("\n📰 最新ニュース（%d件）\n", len(newsList))
	for i, n := range newsList {
		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("[%d] %s\n", i+1, n.Title)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("日付:   %s\n", n.Date)
		fmt.Printf("感情:   %s\n", n.Sentiment)
		fmt.Printf("要約:\n%s\n", n.Summary)
		fmt.Printf("URL:    %s\n", n.URL)
	}
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
}

func truncateString(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func analyzeFluctuation(db *database.SQLiteClient) {
	log.Println("📉 価格変動分析を実行中...")

	// 1. ニュースを取得（実際はDBからその日のニュースを取得するが、今回はAPIから取得）
	apiKey := os.Getenv("NEWSAPI_KEY")
	if apiKey == "" {
		log.Println("⚠️ NEWSAPI_KEYが設定されていません")
		return
	}

	newsFetcher := fetcher.NewNewsFetcher(apiKey)
	// テスト用に3件取得
	articles, err := newsFetcher.FetchTopNews("oil OR gasoline OR economy")
	if err != nil {
		log.Printf("❌ ニュース取得エラー: %v", err)
		return
	}

	if len(articles) > 3 {
		articles = articles[:3]
	}

	// 2. ダミーの価格変動データ
	oldPrice := 160
	newPrice := 165
	priceDiff := newPrice - oldPrice

	log.Printf("📊 検知された価格変動: %d円 -> %d円 (%+d円)", oldPrice, newPrice, priceDiff)
	log.Printf("📰 関連ニュース数: %d件", len(articles))
	log.Println("🤖 Geminiによる分析を開始します...")

	// 3. 分析実行
	ctx := context.Background()
	analysis, err := detect.AnalyzePriceChange(ctx, priceDiff, oldPrice, newPrice, articles)
	if err != nil {
		log.Printf("❌ 分析エラー: %v", err)
		return
	}

	// 4. 結果表示
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🤖 Geminiの分析レポート")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(analysis)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
