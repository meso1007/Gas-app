package detect

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DetectPriceChanges opens the given sqlite DB file, compares the most recent two dates
// of gas_price per region, calculates percent change, and logs/inserts flagged changes.
// dbPath: path to sqlite DB (e.g. "data/gasinsight.db")
// thresholdPct: absolute percent threshold, e.g. 2.0 for ±2%
func DetectPriceChanges(dbPath string, thresholdPct float64) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// ensure change table exists
	if err := ensurePriceChangeTable(db); err != nil {
		return fmt.Errorf("ensure table: %w", err)
	}

	// get two most recent distinct dates
	dates, err := getLatestTwoDates(db)
	if err != nil {
		return fmt.Errorf("get dates: %w", err)
	}
	if len(dates) < 2 {
		log.Printf("[change_detector] not enough dates for comparison (need 2), found %d", len(dates))
		return nil
	}
	dateNew := dates[0] // newest
	dateOld := dates[1] // previous

	// iterate regions
	regions, err := getRegions(db)
	if err != nil {
		return fmt.Errorf("get regions: %w", err)
	}

	for _, region := range regions {
		priceNew, okNew, err := getPriceForRegionDate(db, region, dateNew)
		if err != nil {
			log.Printf("[change_detector] error fetching new price region=%s date=%s: %v", region, dateNew, err)
			continue
		}
		priceOld, okOld, err := getPriceForRegionDate(db, region, dateOld)
		if err != nil {
			log.Printf("[change_detector] error fetching old price region=%s date=%s: %v", region, dateOld, err)
			continue
		}
		if !okNew || !okOld {
			// missing data for this region/date - skip
			log.Printf("[change_detector] skipping region=%s because missing price at one of dates (%s/%s)", region, dateNew, dateOld)
			continue
		}

		// percent change: ((new - old) / old) * 100
		if priceOld == 0 {
			log.Printf("[change_detector] old price zero for region=%s/date=%s, skipping to avoid div0", region, dateOld)
			continue
		}
		pct := ((priceNew - priceOld) / priceOld) * 100
		flagged := math.Abs(pct) >= thresholdPct

		// log result
		log.Printf("[change_detector] region=%s date_new=%s price_new=%.2f date_old=%s price_old=%.2f pct=%.3f flagged=%v",
			region, dateNew, priceNew, dateOld, priceOld, pct, flagged,
		)

		// insert record of this comparison
		if err := insertPriceChange(db, region, dateNew, priceNew, dateOld, priceOld, pct, flagged); err != nil {
			log.Printf("[change_detector] insert error region=%s: %v", region, err)
		}

		// if flagged -> you can call notification here or collect flagged items for later notification
		if flagged {
			// keep it simple: just log; integration (Slack/LINE) should read from price_change table or
			// you can directly notify here by calling a notifier.
			log.Printf("[change_detector] ALERT: region=%s pct_change=%.3f%% (threshold %.2f%%)", region, pct, thresholdPct)
		}
	}

	return nil
}

func ensurePriceChangeTable(db *sql.DB) error {
	create := `
	CREATE TABLE IF NOT EXISTS price_change (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		region TEXT,
		date_new TEXT,
		price_new REAL,
		date_old TEXT,
		price_old REAL,
		pct_change REAL,
		flagged INTEGER,
		created_at DATETIME DEFAULT (datetime('now'))
	);
	`
	_, err := db.Exec(create)
	return err
}

func getLatestTwoDates(db *sql.DB) ([]string, error) {
	// distinct dates in gas_price table ordered desc
	rows, err := db.Query(`SELECT DISTINCT date FROM gas_prices ORDER BY date DESC LIMIT 2`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		dates = append(dates, d)
	}
	return dates, rows.Err()
}

func getRegions(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT region FROM gas_prices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regs []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		regs = append(regs, r)
	}
	return regs, rows.Err()
}

func getPriceForRegionDate(db *sql.DB, region, date string) (float64, bool, error) {
	row := db.QueryRow(`SELECT regular_price FROM gas_prices WHERE region = ? AND date = ? LIMIT 1`, region, date)
	var p sql.NullFloat64
	if err := row.Scan(&p); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	if !p.Valid {
		return 0, false, nil
	}
	return p.Float64, true, nil
}

func insertPriceChange(db *sql.DB, region, dateNew string, priceNew float64, dateOld string, priceOld float64, pct float64, flagged bool) error {
	stmt := `INSERT INTO price_change (region, date_new, price_new, date_old, price_old, pct_change, flagged, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(stmt, region, dateNew, priceNew, dateOld, priceOld, pct, boolToInt(flagged), time.Now().UTC())
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
