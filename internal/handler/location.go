package handler

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/fetcher"
	"github.com/samson/tidal-weather-clock/web/templates"
)

type LocationHandler struct {
	db      *db.Queries
	fetcher *fetcher.Service
}

func NewLocationHandler(queries *db.Queries, svc *fetcher.Service) *LocationHandler {
	return &LocationHandler{db: queries, fetcher: svc}
}

func (h *LocationHandler) SetupPage(c *gin.Context) {
	stations, err := fetcher.FetchStations(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "error fetching stations")
		return
	}
	groups := fetcher.GroupByCountry(stations)
	c.Header("Content-Type", "text/html")
	templates.SetupLocation(groups).Render(c.Request.Context(), c.Writer)
}

func (h *LocationHandler) SaveLocation(c *gin.Context) {
	user := auth.GetUser(c)
	stationID := c.PostForm("station_id")
	if stationID == "" {
		c.Redirect(http.StatusSeeOther, "/setup/location?error=no_station")
		return
	}

	stations, err := fetcher.FetchStations(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "error fetching stations")
		return
	}

	var chosen *fetcher.Station
	for i := range stations {
		if stations[i].ID == stationID {
			chosen = &stations[i]
			break
		}
	}
	if chosen == nil {
		c.Redirect(http.StatusSeeOther, "/setup/location?error=invalid_station")
		return
	}

	ctx := c.Request.Context()

	// Find existing location record for this station, or create one.
	isNew := false
	loc, err := h.db.GetLocationByStationID(ctx, chosen.ID)
	if errors.Is(err, sql.ErrNoRows) {
		loc, err = h.db.CreateLocation(ctx, db.CreateLocationParams{
			Name:          chosen.Name,
			Lat:           chosen.Lat,
			Lon:           chosen.Lon,
			TideStationID: chosen.ID,
		})
		if err != nil {
			c.String(http.StatusInternalServerError, "internal error")
			return
		}
		isNew = true
	} else if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	// Replace user's current location link.
	if err := h.db.DeleteUserLocations(ctx, user.ID); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	if err := h.db.LinkUserLocation(ctx, db.LinkUserLocationParams{
		UserID:     user.ID,
		LocationID: loc.ID,
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	// Populate data for newly added locations so the home page isn't empty.
	if isNew {
		fetchCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.fetcher.FetchForLocation(fetchCtx, loc); err != nil {
			log.Printf("ad-hoc fetch for location %s: %v", loc.Name, err)
			// Non-fatal: user can still see the page, data will arrive on next nightly run.
		}
	}

	c.Redirect(http.StatusSeeOther, "/")
}
