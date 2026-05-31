package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
	"github.com/samson/tidal-weather-clock/internal/scoring"
	"github.com/samson/tidal-weather-clock/web/templates"
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

func (h *DayHandler) Show(c *gin.Context) {
	date := parseDate(c.Param("date"))
	user := auth.GetUser(c)

	location, ok := h.resolveLocation(c, user)
	if !ok {
		return
	}

	start := truncateToDay(date.UTC())
	end := start.Add(24 * time.Hour)

	dbHours, err := h.db.GetHourlyDataForLocation(c.Request.Context(), db.GetHourlyDataForLocationParams{
		LocationID: location.ID,
		Time:       start,
		Time_2:     end,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	hours := dbHoursToDomain(dbHours)

	var windows []scoring.ScoredWindow
	if user != nil && len(hours) > 0 {
		dbActivities, err := h.db.ListActivities(c.Request.Context(), user.ID)
		if err != nil {
			c.String(http.StatusInternalServerError, "internal error")
			return
		}
		activities, err := dbActivitiesToDomain(dbActivities)
		if err != nil {
			c.String(http.StatusInternalServerError, "internal error")
			return
		}
		windows = scoring.ScoreDay(activities, hours)
	}

	data := templates.DayViewData{
		Date:     date,
		PrevDate: date.AddDate(0, 0, -1).Format("2006-01-02"),
		NextDate: date.AddDate(0, 0, 1).Format("2006-01-02"),
		Hours:    hours,
		Windows:  windows,
		User:     user,
		HasData:  len(hours) > 0,
	}

	if c.GetHeader("Content-Type") == "application/json" {
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, data)
		return
	}

	c.Header("Content-Type", "text/html")
	templates.DayView(data).Render(c.Request.Context(), c.Writer)
}

// resolveLocation returns the user's location if logged in, or the first
// available location for anonymous users. Returns false if it handled the
// response itself (redirect or error).
func (h *DayHandler) resolveLocation(c *gin.Context, user *domain.User) (db.Location, bool) {
	if user != nil {
		loc, err := h.db.GetLocationByUser(c.Request.Context(), user.ID)
		if errors.Is(err, sql.ErrNoRows) {
			c.Redirect(http.StatusSeeOther, "/setup/location")
			return db.Location{}, false
		}
		if err != nil {
			c.String(http.StatusInternalServerError, "internal error")
			return db.Location{}, false
		}
		return loc, true
	}

	locs, err := h.db.GetAllLocations(c.Request.Context())
	if err != nil || len(locs) == 0 {
		return db.Location{}, true // no location, HasData will be false
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

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func dbHoursToDomain(rows []db.HourlyDatum) []domain.HourlyData {
	hours := make([]domain.HourlyData, len(rows))
	for i, r := range rows {
		hours[i] = domain.HourlyData{
			ID:          r.ID,
			LocationID:  r.LocationID,
			Time:        r.Time,
			WindSpeedMS: r.WindSpeedMs,
			WindDirDeg:  r.WindDirDeg,
			Weather:     domain.WeatherCode(r.WeatherCode),
			TideHeightM: r.TideHeightM,
			FetchedAt:   r.FetchedAt,
		}
	}
	return hours
}
