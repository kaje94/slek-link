package main

import (
	"log"
	"os"

	"github.com/go-gormigrate/gormigrate/v2"
	gormModels "github.com/kaje94/slek-link/gorm/pkg"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	envVal := os.Getenv("POSTGRESQL_DSN")
	if envVal == "" {
		log.Fatalf("POSTGRESQL_DSN env value is required")
	}

	db, err := gorm.Open(postgres.Open(envVal), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// your migrations here
		// Refer https://github.com/go-gormigrate/gormigrate on how to add migrations if needed
		{
			// drop `country_name` column from `link_country_clicks` table
			ID: "202503172200",
			Migrate: func(tx *gorm.DB) error {
				// when table already exists, define only columns that are about to change
				type LinkCountryClicks struct {
					CountryName string `gorm:"countryName"`
				}
				return tx.Migrator().DropColumn(&LinkCountryClicks{}, "CountryName")
			},
			Rollback: func(tx *gorm.DB) error {
				type LinkCountryClicks struct {
					CountryName string `gorm:"countryName"`
				}
				return db.Migrator().AddColumn(&LinkCountryClicks{}, "CountryName")
			},
		},
	})

	m.InitSchema(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(&gormModels.User{}, &gormModels.Link{}, &gormModels.LinkMonthlyClicks{}, &gormModels.LinkCountryClicks{})
		if err != nil {
			return err
		}

		return nil
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration did run successfully")
}
