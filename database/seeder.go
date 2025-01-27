package database

import (
	"fmt"

	"review_product_tokopedia_be/database/seeds"

	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) error {
	fmt.Println("Seeding User")
	if err := seeds.UserSeeder(db); err != nil {
		return err
	}

	return nil
}
