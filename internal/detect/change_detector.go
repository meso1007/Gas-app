package detect

import "backend/internal/model"

type ChangeResult struct {
	FuelType string  // 例: "Regular", "Premium"
	OldPrice float64 // 前回価格
	NewPrice float64 // 今回価格
	Change   float64 // 変化量 (NewPrice - OldPrice)
	Status   string  // "up", "down", "same"
}

func DetectChanges(oldData, newData []model.GasPrice) []ChangeResult {
	results := []ChangeResult{}

	for _, newItem := range newData {
		for _, oldItem := range oldData {
			if newItem.FuelType == oldItem.FuelType {
				change := newItem.Price - oldItem.Price
				status := "same"
				if change > 0 {
					status = "up"
				} else if change < 0 {
					status = "down"
				}
				results = append(results, ChangeResult{
					FuelType: newItem.FuelType,
					OldPrice: oldItem.Price,
					NewPrice: newItem.Price,
					Change:   change,
					Status:   status,
				})
			}
		}
	}

	return results
}
