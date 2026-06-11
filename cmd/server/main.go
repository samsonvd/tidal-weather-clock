package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/fetcher"
	"github.com/samson/tidal-weather-clock/internal/handler"
	"github.com/samson/tidal-weather-clock/internal/mailer"
)

var isDevelopment = os.Getenv("APP_ENV") == "development"

func main() {
	database, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		log.Fatalf("API_URL not set")
	}

	appUrl := os.Getenv("APP_URL")
	if appUrl == "" {
		log.Fatalf("APP_URL not set")
	}
	if !strings.HasSuffix(appUrl, "#") {
		appUrl += "#"
	}

	queries := db.New(database)
	fetchSvc := fetcher.NewService(queries)

	var mailSvc mailer.Mailer = mailer.NewResendMailer(os.Getenv("RESEND_API_KEY"), apiUrl)
	if isDevelopment {
		mailSvc = mailer.NewLogMailer(apiUrl)
	}
	authHandler := auth.NewHandler(queries, mailSvc, appUrl)

	activityHandler := handler.NewActivityHandler(queries)
	locationHandler := handler.NewLocationHandler(queries, fetchSvc)
	dayHandler := handler.NewDayHandler(queries)

	r := gin.Default()

	var allowedOriginsList []string
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOriginsList = []string{appUrl} //localhost:5173"
	} else {
		allowedOriginsList = strings.Split(allowedOrigins, ",")
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOriginsList,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
