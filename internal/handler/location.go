package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/fetcher"
	"github.com/samson/tidal-weather-clock/web/templates"
)

type LocationHandler struct {
	db *db.Queries
}

func NewLocationHandler(queries *db.Queries) *LocationHandler {
	return &LocationHandler{db: queries}
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

	if _, err := h.db.CreateLocation(c.Request.Context(), db.CreateLocationParams{
		UserID:        user.ID,
		Name:          chosen.Name,
		Lat:           chosen.Lat,
		Lon:           chosen.Lon,
		TideStationID: chosen.ID,
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}
