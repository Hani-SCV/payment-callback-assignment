package main

import (
	"log"

	"github.com/Hani-SCV/payment-callback-assignment/internal/app"
	"github.com/Hani-SCV/payment-callback-assignment/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server started on :8080")

	if err := application.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}