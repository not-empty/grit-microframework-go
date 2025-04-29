package app

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/not-empty/grit/app/config"
	"github.com/not-empty/grit/app/database"
	"github.com/not-empty/grit/app/router"

	_ "github.com/not-empty/grit/app/router/domains"
	_ "github.com/not-empty/grit/app/router/registry"
	_ "github.com/not-empty/grit/app/router/routes"
)

func Bootstrap() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	config.LoadConfig()

	dbConfig := database.LoadDatabaseConfig()
	db := database.Init(dbConfig)
	router.RegisterRoutes(db)
}

func StartServer() {
	port := config.AppConfig.AppPort

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, http.DefaultServeMux))
}
