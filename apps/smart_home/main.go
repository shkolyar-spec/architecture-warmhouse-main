package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smarthome/db"
	"smarthome/handlers"
	"smarthome/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/smarthome")
	tempAPIURL := getEnv("TEMPERATURE_API_URL", "http://temperature-api:8081")

	// init DB
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer pool.Close()

	database := &db.DB{Pool: pool}

	// init services
	tempService := services.NewTemperatureService(tempAPIURL)
	sensorHandler := handlers.NewSensorHandler(database, tempService)

	// init router
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	sensorHandler.RegisterRoutes(api)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// run server in goroutine
	go func() {
		log.Println("smart_home listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v\n", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
