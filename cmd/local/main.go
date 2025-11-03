package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"gasinsight/internal/database"
	"gasinsight/internal/fetcher"
	"gasinsight/internal/models"
	"gasinsight/internal/services"
)

func main() {
	mode := flag.String("mode", "fetch", "モード")
	dbPath := flag.String("db", "./data/gasinsight.db", "DBパス")
	useScraping := flag.Bool("scrape", false, "スクレイピングを使用")
	useMock := flag.Bool("mock", true, "モック使用")
	detectChange := flag.Bool("detect", true, "変動検知を有効化")
	flag.Parse()

	log.Println("🚀 GasInsight ローカル実行版")

	db, err := database.NewSQLiteClient(*dbPath)
	if err != nil {
		log.Fatalf("❌ エラー: %v", err)
	}
	defer db.Close()

	switch *mode {
	case "fetch":
		fetchGasPrice(db, *useScraping, *useMock, *detectChange)
	case "fetch-exchange":
		fetchExchangeRate(db, *useMock, *detectChange)
	case "fetch-all":
		fetchGasPrice(db, *useScraping, *useMock, *detectChange)
		fetchExchangeRate(db, *useMock, *detectChange)
	case "list":
		listGasPrices(db)
	case "list-exchange":
		listExchangeRates(db)
	case "latest":
		latestGasPrice(db)
	case "latest-exchange":
		latestExchangeRate(db)
	default:
		log.Fatalf("❌ 不正なモード: %s", *mode)
	}

	log.Println("✅ 処理完了")
}

func fetchGasPrice(db *database.SQLiteClient, useScraping bool, useMock bool, detectChange bool) {
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

	price := models.NewGasPrice(data.Date, data.Region,
		data.RegularPrice, data.PremiumPrice, data.DieselPrice)

	if err := db.SaveGasPrice(price); err != nil {
		log.Fatalf("❌ 保存エラー: %v", err)
	}

	printGasPrice(price)

	// 変動検知
	if detectChange {
		detector := services.NewChangeDetector(db)
		changes, err := detector.DetectGasPriceChanges(price)
		if err != nil {
			log.Printf("⚠️  変動検知エラー: %v", err)
		} else if len(changes) > 0 {
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📈 価格変動サマリー")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
		}
	}
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

	rate := models.NewExchangeRate(data.Date, data.USDJPY, data.EURJPY, data.GBPJPY, data.CNYJPY)

	if err := db.SaveExchangeRate(rate); err != nil {
		log.Fatalf("❌ 保存エラー: %v", err)
	}

	printExchangeRate(rate)

	// 変動検知
	if detectChange {
		detector := services.NewChangeDetector(db)
		changes, err := detector.DetectExchangeRateChanges(rate)
		if err != nil {
			log.Printf("⚠️  変動検知エラー: %v", err)
		} else if len(changes) > 0 {
			fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("📈 為替変動サマリー")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
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

	fmt.Printf("\n💱 為替レートデータ一覧（%d件）\n\n", len(rates))
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

func printGasPrice(p *models.GasPrice) {
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

func printExchangeRate(r *models.ExchangeRate) {
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
