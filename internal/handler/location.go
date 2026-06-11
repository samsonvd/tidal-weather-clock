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
	"github.com/samson/tidal-weather-clock/internal/domain"
	"github.com/samson/tidal-weather-clock/internal/fetcher"
)

type LocationHandler struct {
	db      *db.Queries
	fetcher *fetcher.Service
}

func NewLocationHandler(queries *db.Queries, svc *fetcher.Service) *LocationHandler {
	return &LocationHandler{db: queries, fetcher: svc}
}

func (h *LocationHandler) Get(c *gin.Context) {
	user := auth.GetUser(c)
	loc, err := h.db.GetLocationByUser(c.Request.Context(), user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusOK, nil)
		return
	}
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, domain.Location{
		ID:            loc.ID,
		Name:          loc.Name,
		Lat:           loc.Lat,
		Lon:           loc.Lon,
		TideStationID: loc.TideStationID,
		CreatedAt:     loc.CreatedAt,
	})
}

func (h *LocationHandler) Set(c *gin.Context) {
	user := auth.GetUser(c)
	var body struct {
		StationID string `json:"station_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.StationID == "" {
		apiError(c, http.StatusBadRequest, "station_id required")
		return
	}

	stations, err := fetcher.FetchStations(c.Request.Context())
	if err != nil {
		internalError(c)
		return
	}

	var chosen *fetcher.Station
	for i := range stations {
		if stations[i].ID == body.StationID {
			chosen = &stations[i]
			break
		}
	}
	if chosen == nil {
		apiError(c, http.StatusBadRequest, "unknown station id")
		return
	}

	ctx := c.Request.Context()
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
			internalError(c)
			return
		}
		isNew = true
	} else if err != nil {
		internalError(c)
		return
	}

	if err := h.db.DeleteUserLocations(ctx, user.ID); err != nil {
		internalError(c)
		return
	}
	if err := h.db.LinkUserLocation(ctx, db.LinkUserLocationParams{
		UserID:     user.ID,
		LocationID: loc.ID,
	}); err != nil {
		internalError(c)
		return
	}

	if isNew {
		go func() {
			fetchCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := h.fetcher.FetchForLocation(fetchCtx, loc); err != nil {
				log.Printf("ad-hoc fetch for %s: %v", loc.Name, err)
			}
		}()
	}

	c.JSON(http.StatusOK, domain.Location{
		ID:            loc.ID,
		Name:          loc.Name,
		Lat:           loc.Lat,
		Lon:           loc.Lon,
		TideStationID: loc.TideStationID,
		CreatedAt:     loc.CreatedAt,
	})
}

func (h *LocationHandler) Stations(c *gin.Context) {
	stations, err := fetcher.FetchStations(c.Request.Context())
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, fetcher.GroupByCountry(stations))
}
