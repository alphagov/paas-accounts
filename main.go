package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alphagov/paas-accounts/api"
	"github.com/alphagov/paas-accounts/database"
)

var globalContext context.Context

func init() {
	ctx, shutdown := context.WithCancel(context.Background())
	globalContext = ctx
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		shutdown()
	}()
}

func Main() error {
	db, err := database.NewDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	if err := db.Init(); err != nil {
		return err
	}

	server := api.NewServer(api.Config{
		DB:                db,
		BasicAuthUsername: os.Getenv("BASIC_AUTH_USERNAME"),
		BasicAuthPassword: os.Getenv("BASIC_AUTH_PASSWORD"),
	})
	addr := fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT"))
	fmt.Println("server started at", addr)
	return api.ListenAndServe(globalContext, server, addr)
}

func main() {
	if err := Main(); err != nil {
		log.Fatal(err)
	}
}
