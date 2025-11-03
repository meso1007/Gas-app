package models

import "time"

type GasPrice struct {
	ID           string  `json:"id"`
	Date         string  `json:"date"`
	RegularPrice float64 `json:"regular_price"`
	PremiumPrice float64 `json:"premium_price"`
	DieselPrice  float64 `json:"diesel_price"`
	Region       string  `json:"region"`
	Source       string  `json:"source"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
}

func NewGasPrice(date, region string, regular, premium, diesel float64) *GasPrice {
	now := time.Now().Unix()
	return &GasPrice{
		ID:           date,
		Date:         date,
		RegularPrice: regular,
		PremiumPrice: premium,
		DieselPrice:  diesel,
		Region:       region,
		Source:       "e-nenpi",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
