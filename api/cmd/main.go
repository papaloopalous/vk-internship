package main

import (
	_ "api/internal/load_config"
	"api/internal/logger"
	"api/internal/router"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

func main() {
	apiPort := viper.GetString("api.port")
	router := router.CreateNewRouter()

	srv := &http.Server{
		Addr:    apiPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server is starting on %s", apiPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	defer logger.Sync()

	log.Println("Server gracefully stopped")
}
