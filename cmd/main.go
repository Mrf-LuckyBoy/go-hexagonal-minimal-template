package main

import (
	"fmt"

	"github.com/Mrf-LuckyBoy/go-hexagonal-minimal-template/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	cfg, err := config.Load("")
	if err != nil {
		panic(err)
	}

	BuildContainer(cfg)

	fmt.Println(cfg.App.EncryptKey)

	fmt.Println("hello world")
}
