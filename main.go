package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/not-empty/grit/app/database"
	"github.com/not-empty/grit/app/router"

	_ "github.com/not-empty/grit/app/router/domains"
	_ "github.com/not-empty/grit/app/router/registry"
	_ "github.com/not-empty/grit/app/router/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8001"
	}

	dbConfig := database.LoadConfigFromEnv()
	db := database.Init(dbConfig)
	router.RegisterRoutes(db)

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, http.DefaultServeMux))
}
