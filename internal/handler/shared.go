package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
)

func apiError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func dbActivityToDomain(row db.Activity) (domain.Activity, error) {
	var constraints []domain.Constraint
	if err := json.Unmarshal(row.Constraints, &constraints); err != nil {
		return domain.Activity{}, err
	}
	if constraints == nil {
		constraints = []domain.Constraint{}
	}
	return domain.Activity{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		DurationHrs: int(row.DurationHrs),
		WindowStart: int(row.WindowStart),
		WindowEnd:   int(row.WindowEnd),
		Constraints: constraints,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func dbActivitiesToDomain(rows []db.Activity) ([]domain.Activity, error) {
	activities := make([]domain.Activity, 0, len(rows))
	for _, row := range rows {
		a, err := dbActivityToDomain(row)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}
	return activities, nil
}

func dbHoursToDomain(rows []db.HourlyDatum) []domain.HourlyData {
	hours := make([]domain.HourlyData, len(rows))
	for i, r := range rows {
		hours[i] = domain.HourlyData{
			ID:                 r.ID,
			LocationID:         r.LocationID,
			Time:               r.Time,
			WindSpeedMS:        r.WindSpeedMs,
			WindDirDeg:         r.WindDirDeg,
			Weather:            domain.WeatherCode(r.WeatherCode),
			TideHeightM:        r.TideHeightM,
			TemperatureCelsius: r.TemperatureCelsius,
			FetchedAt:          r.FetchedAt,
		}
	}
	return hours
}

func internalError(c *gin.Context) {
	apiError(c, http.StatusInternalServerError, "internal error")
}
