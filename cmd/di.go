package main

import (
	"fmt"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/adapters/repository/sqliter"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/config"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/domain/model"
	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/core/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Container struct{}

func BuildContainer(cfg *config.Config) *Container {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("Fail to connect database")
	}

	fmt.Println("Connected to database")

	err = db.AutoMigrate(&model.Book{})
	if err != nil {
		panic("Fail to migratetion")
	}

	// repository
	bookRepository := sqliter.NewBookRepositorySqlite(db)

	// service
	bookService := service.NewBookService(bookRepository)

	fmt.Println(bookService)

	fmt.Println("kuykuykuy")

	return nil
}
