package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/fetcher"
	"github.com/samson/tidal-weather-clock/internal/handler"
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
	fetchSvc := fetcher.NewService(queries)
	m := &mailer.LogMailer{}
	authHandler := auth.NewHandler(queries, m)
	activityHandler := handler.NewActivityHandler(queries)
	locationHandler := handler.NewLocationHandler(queries, fetchSvc)
	dayHandler := handler.NewDayHandler(queries)

	r := gin.Default()
	r.Use(auth.SessionMiddleware(queries))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/login", authHandler.RequestLink)
	r.GET("/auth/verify", authHandler.VerifyToken)
	r.POST("/auth/logout", authHandler.Logout)

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/stations", locationHandler.Stations)

		protected := apiV1.Group("/", auth.RequireAuth)
		{
			protected.GET("/me", authHandler.Me)
			protected.GET("/locations", locationHandler.Get)
			protected.PUT("/locations", locationHandler.Set)

			protected.GET("/activities", activityHandler.List)
			protected.POST("/activities", activityHandler.Create)
			protected.PUT("/activities/:id", activityHandler.Update)
			protected.DELETE("/activities/:id", activityHandler.Delete)

			protected.GET("/day/:date", dayHandler.Show)
		}

	}

	r.Static("/assets", "frontend/dist/assets")
	r.StaticFile("/favicon.svg", "frontend/dist/favicon.svg")
	r.StaticFile("/icons.svg", "frontend/dist/icons.svg")
	r.NoRoute(func(c *gin.Context) {
		c.File("frontend/dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
