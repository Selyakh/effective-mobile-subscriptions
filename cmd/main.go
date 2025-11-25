package main

import (
	"database/sql"
	_ "effective-mobile-subscriptions/docs"
	"effective-mobile-subscriptions/internal/config"
	"effective-mobile-subscriptions/internal/handler"
	"effective-mobile-subscriptions/internal/repository"
	"effective-mobile-subscriptions/internal/service"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// @title Subscription Aggregation API
// @version 1.0
// @description REST service for aggregating user subscriptions.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

func main() {
	// загрузка конфигурации
	cfg, err := config.LoadConfig("./internal/config")
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}

	// инициализация бд
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open connection to DB: %v", err)
	}
	defer db.Close()

	// проверка соединения
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")

	// инициализация слоев
	subRepo := repository.NewSubscriptionRepository(db)
	subService := service.NewSubscriptionService(subRepo)
	subHandler := handler.NewSubscriptionHandler(subService)

	// настройка Роутера
	r := mux.NewRouter()

	r.HandleFunc("/subscriptions", subHandler.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions", subHandler.ListSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions/analytics", subHandler.GetCostAnalytics).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", subHandler.GetSubscriptionByID).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", subHandler.UpdateSubscription).Methods("PUT")
	r.HandleFunc("/subscriptions/{id}", subHandler.DeleteSubscription).Methods("DELETE")
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs"))))

	// запуск HTTP-сервера
	addr := ":" + cfg.Server.Port
	log.Printf("The service is running at the address %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server startup error: %v", err)
	}
}
