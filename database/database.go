package database

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	DB   *gorm.DB
	once sync.Once // Ensure connection is established only once
)

func ConnectDB(databaseURL string) {
	once.Do(func() {
		config := &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // Important for your schema
			},
		}
		var err error

		DB, err = gorm.Open(postgres.Open(databaseURL), config)

		if err != nil {
			log.Fatal("failed to connect database:", err)
		}

		fmt.Println("Successfully connected to the database!")

	})
}
