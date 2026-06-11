package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
	"github.com/samson/tidal-weather-clock/internal/scoring"
)

var londonTZ *time.Location

func init() {
	var err error
	londonTZ, err = time.LoadLocation("Europe/London")
	if err != nil {
		londonTZ = time.UTC
	}
}

type DayHandler struct {
	db *db.Queries
}

func NewDayHandler(queries *db.Queries) *DayHandler {
	return &DayHandler{db: queries}
}

type dayResponse struct {
	Date     string               `json:"date"`
	Location *domain.Location     `json:"location"`
	Hours    []domain.HourlyData  `json:"hours"`
	Windows  []scoring.ScoredWindow `json:"windows"`
}

func (h *DayHandler) Show(c *gin.Context) {
	date := parseDate(c.Param("date"))
	user := auth.GetUser(c)

	loc, ok := h.resolveLocation(c, user)
	if !ok {
		return
	}

	start := date
	end := date.Add(24 * time.Hour)

	dbHours, err := h.db.GetHourlyDataForLocation(c.Request.Context(), db.GetHourlyDataForLocationParams{
		LocationID: loc.ID,
		Time:       start,
		Time_2:     end,
	})
	if err != nil {
		internalError(c)
		return
	}

	hours := dbHoursToDomain(dbHours)

	var windows []scoring.ScoredWindow
	if user != nil && len(hours) > 0 {
		dbActivities, err := h.db.ListActivities(c.Request.Context(), user.ID)
		if err != nil {
			internalError(c)
			return
		}
		activities, err := dbActivitiesToDomain(dbActivities)
		if err != nil {
			internalError(c)
			return
		}
		windows = scoring.ScoreDay(activities, hours)
	}
	if windows == nil {
		windows = []scoring.ScoredWindow{}
	}

	domainLoc := domain.Location{
		ID:            loc.ID,
		Name:          loc.Name,
		Lat:           loc.Lat,
		Lon:           loc.Lon,
		TideStationID: loc.TideStationID,
		CreatedAt:     loc.CreatedAt,
	}

	c.JSON(http.StatusOK, dayResponse{
		Date:     date.Format("2006-01-02"),
		Location: &domainLoc,
		Hours:    hours,
		Windows:  windows,
	})
}

func (h *DayHandler) resolveLocation(c *gin.Context, user *domain.User) (db.Location, bool) {
	if user != nil {
		loc, err := h.db.GetLocationByUser(c.Request.Context(), user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, dayResponse{Date: parseDate(c.Param("date")).Format("2006-01-02"), Hours: []domain.HourlyData{}, Windows: []scoring.ScoredWindow{}})
			return db.Location{}, false
		}
		if err != nil {
			internalError(c)
			return db.Location{}, false
		}
		return loc, true
	}

	locs, err := h.db.GetAllLocations(c.Request.Context())
	if err != nil || len(locs) == 0 {
		c.JSON(http.StatusOK, dayResponse{Date: parseDate(c.Param("date")).Format("2006-01-02"), Hours: []domain.HourlyData{}, Windows: []scoring.ScoredWindow{}})
		return db.Location{}, false
	}
	return locs[0], true
}

func parseDate(s string) time.Time {
	if s != "" {
		if t, err := time.ParseInLocation("2006-01-02", s, londonTZ); err == nil {
			return t
		}
	}
	return time.Now().In(londonTZ)
}

