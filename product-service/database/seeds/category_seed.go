package seeds

import (
	"log"
	"product-service/internal/core/domain/model"

	"gorm.io/gorm"
)

// SeedCategory inserts the starter categories. FirstOrCreate keeps it
// idempotent: running the server twice does not duplicate rows.
func SeedCategory(db *gorm.DB) {
	categories := []model.Category{
		{Name: "Elektronik", Slug: "elektronik", Status: true, Description: "Perangkat elektronik"},
		{Name: "Pakaian", Slug: "pakaian", Status: true, Description: "Busana pria dan wanita"},
		{Name: "Makanan", Slug: "makanan", Status: true, Description: "Makanan dan minuman"},
	}

	for _, category := range categories {
		if err := db.FirstOrCreate(&category, model.Category{Slug: category.Slug}).Error; err != nil {
			log.Fatalf("[SeedCategory-1] %v", err)
		} else {
			log.Printf("[SeedCategory-2] Category %s ready", category.Slug)
		}
	}
}
