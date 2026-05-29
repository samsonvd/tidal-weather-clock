package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/mailer"
)

func main() {
	database, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	queries := db.New(database)
	m := &mailer.LogMailer{}
	authHandler := auth.NewHandler(queries, m)

	r := gin.Default()
	r.Use(auth.SessionMiddleware(queries))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/login", func(c *gin.Context) {
		c.String(200, "login page — sent=%s error=%s", c.Query("sent"), c.Query("error"))
	})
	r.POST("/login", authHandler.RequestLink)
	r.GET("/auth/verify", authHandler.VerifyToken)
	r.POST("/auth/logout", authHandler.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
