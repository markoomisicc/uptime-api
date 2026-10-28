package main

import (
	"context"
	"log"
	"net/http"
	"time"
	_ "time/tzdata"

	"github.com/markoomisicc/uptime-api/internal/api"
	"github.com/markoomisicc/uptime-api/internal/config"
	"github.com/markoomisicc/uptime-api/internal/database"
	"github.com/markoomisicc/uptime-api/internal/mail"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Println("Connected to database")

	mailer := mail.New(cfg.SMTP)
	server := api.NewServer(cfg, pool, mailer)

	log.Printf("API listening on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
