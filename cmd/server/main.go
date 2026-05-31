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
	"github.com/samson/tidal-weather-clock/web/templates"
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
	settingsHandler := handler.NewSettingsHandler(queries)

	r := gin.Default()
	r.Use(auth.SessionMiddleware(queries))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/", dayHandler.Show)
	r.GET("/day/:date", dayHandler.Show)

	r.GET("/login", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		templates.Login(c.Query("sent") == "true", c.Query("error") != "").Render(c.Request.Context(), c.Writer)
	})
	r.POST("/login", authHandler.RequestLink)
	r.GET("/auth/verify", authHandler.VerifyToken)
	r.POST("/auth/logout", authHandler.Logout)

	protected := r.Group("/", auth.RequireAuth)
	{
		protected.GET("/setup/location", locationHandler.SetupPage)
		protected.POST("/setup/location", locationHandler.SaveLocation)

		protected.GET("/settings/activities", settingsHandler.ListActivities)
		protected.GET("/settings/activities/new", settingsHandler.NewActivity)
		protected.POST("/settings/activities", settingsHandler.CreateActivity)
		protected.GET("/settings/activities/:id/edit", settingsHandler.EditActivity)
		protected.POST("/settings/activities/:id", settingsHandler.UpdateActivity)
		protected.DELETE("/settings/activities/:id", settingsHandler.DeleteActivity)

		protected.GET("/settings/constraint/row", settingsHandler.ConstraintRowFragment)
		protected.GET("/settings/constraint/fields", settingsHandler.ConstraintFieldsFragment)

		// Legacy JSON API kept for potential future use.
		protected.GET("/activities", activityHandler.List)
		protected.POST("/activities", activityHandler.Create)
		protected.PUT("/activities/:id", activityHandler.Update)
		protected.DELETE("/activities/:id", activityHandler.Delete)
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
