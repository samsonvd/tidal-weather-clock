package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
	"github.com/samson/tidal-weather-clock/web/templates"
)

type SettingsHandler struct {
	db *db.Queries
}

func NewSettingsHandler(queries *db.Queries) *SettingsHandler {
	return &SettingsHandler{db: queries}
}

func (h *SettingsHandler) ListActivities(c *gin.Context) {
	user := auth.GetUser(c)
	rows, err := h.db.ListActivities(c.Request.Context(), user.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	activities, err := dbActivitiesToDomain(rows)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.Header("Content-Type", "text/html")
	templates.ActivityList(user, activities).Render(c.Request.Context(), c.Writer)
}

func (h *SettingsHandler) NewActivity(c *gin.Context) {
	user := auth.GetUser(c)
	c.Header("Content-Type", "text/html")
	templates.ActivityForm(user, nil, false).Render(c.Request.Context(), c.Writer)
}

func (h *SettingsHandler) CreateActivity(c *gin.Context) {
	user := auth.GetUser(c)
	constraints, err := parseConstraints(c)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid constraints")
		return
	}
	raw, _ := json.Marshal(constraints)

	if _, err := h.db.CreateActivity(c.Request.Context(), db.CreateActivityParams{
		UserID:      user.ID,
		Name:        c.PostForm("name"),
		DurationHrs: parseInt32(c.PostForm("duration_hrs")),
		WindowStart: parseInt32(c.PostForm("window_start")),
		WindowEnd:   parseInt32(c.PostForm("window_end")),
		Constraints: raw,
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.Redirect(http.StatusSeeOther, "/settings/activities")
}

func (h *SettingsHandler) EditActivity(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}
	row, err := h.db.GetActivity(c.Request.Context(), db.GetActivityParams{ID: id, UserID: user.ID})
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	activity, err := dbActivityToDomain(row)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.Header("Content-Type", "text/html")
	templates.ActivityForm(user, &activity, true).Render(c.Request.Context(), c.Writer)
}

func (h *SettingsHandler) UpdateActivity(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}
	constraints, err := parseConstraints(c)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid constraints")
		return
	}
	raw, _ := json.Marshal(constraints)

	if _, err := h.db.UpdateActivity(c.Request.Context(), db.UpdateActivityParams{
		ID:          id,
		UserID:      user.ID,
		Name:        c.PostForm("name"),
		DurationHrs: parseInt32(c.PostForm("duration_hrs")),
		WindowStart: parseInt32(c.PostForm("window_start")),
		WindowEnd:   parseInt32(c.PostForm("window_end")),
		Constraints: raw,
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.Redirect(http.StatusSeeOther, "/settings/activities")
}

func (h *SettingsHandler) DeleteActivity(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.db.DeleteActivity(c.Request.Context(), db.DeleteActivityParams{
		ID:     id,
		UserID: user.ID,
	}); err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SettingsHandler) ConstraintRowFragment(c *gin.Context) {
	index, _ := strconv.Atoi(c.Query("index"))
	c.Header("Content-Type", "text/html")
	templates.ConstraintRow(index, nil).Render(c.Request.Context(), c.Writer)
}

func (h *SettingsHandler) ConstraintFieldsFragment(c *gin.Context) {
	index, _ := strconv.Atoi(c.Query("index"))
	typStr := c.Query(fmt.Sprintf("constraint_type_%d", index))
	c.Header("Content-Type", "text/html")
	templates.ConstraintFields(index, domain.ConstraintType(typStr), nil).Render(c.Request.Context(), c.Writer)
}

func parseConstraints(c *gin.Context) ([]domain.Constraint, error) {
	var constraints []domain.Constraint
	for i := 0; ; i++ {
		typStr := c.PostForm(fmt.Sprintf("constraint_type_%d", i))
		if typStr == "" {
			break
		}
		cType := domain.ConstraintType(typStr)
		required := c.PostForm(fmt.Sprintf("constraint_required_%d", i)) == "on"
		weight, _ := strconv.ParseFloat(c.PostForm(fmt.Sprintf("constraint_weight_%d", i)), 64)

		constraint := domain.Constraint{
			Type:     cType,
			Required: required,
			Weight:   weight,
		}

		switch cType {
		case domain.ConstraintWindSpeed, domain.ConstraintTideHeight:
			acceptableMin, _ := strconv.ParseFloat(c.PostForm(fmt.Sprintf("constraint_acceptable_min_%d", i)), 64)
			idealMin, _ := strconv.ParseFloat(c.PostForm(fmt.Sprintf("constraint_ideal_min_%d", i)), 64)
			idealMax, _ := strconv.ParseFloat(c.PostForm(fmt.Sprintf("constraint_ideal_max_%d", i)), 64)
			acceptableMax, _ := strconv.ParseFloat(c.PostForm(fmt.Sprintf("constraint_acceptable_max_%d", i)), 64)
			if cType == domain.ConstraintWindSpeed {
				acceptableMin = ktsToMS(acceptableMin)
				idealMin = ktsToMS(idealMin)
				idealMax = ktsToMS(idealMax)
				acceptableMax = ktsToMS(acceptableMax)
			}
			constraint.AcceptableMin = acceptableMin
			constraint.IdealMin = idealMin
			constraint.IdealMax = idealMax
			constraint.AcceptableMax = acceptableMax
		case domain.ConstraintWindDir:
			points := c.PostFormArray(fmt.Sprintf("constraint_directions_%d", i))
			dirs := make([]float64, len(points))
			for j, p := range points {
				dirs[j] = compassPointToDeg(p)
			}
			constraint.Preferred = dirs
			constraint.Tolerance, _ = strconv.ParseFloat(c.PostForm(fmt.Sprintf("constraint_tolerance_%d", i)), 64)
		case domain.ConstraintWeather:
			constraint.Acceptable = c.PostFormArray(fmt.Sprintf("constraint_acceptable_weather_%d", i))
			constraint.Ideal = c.PostFormArray(fmt.Sprintf("constraint_ideal_weather_%d", i))
		}

		constraints = append(constraints, constraint)
	}
	return constraints, nil
}

func ktsToMS(kts float64) float64 { return kts * 0.514444 }

var compassDegrees = map[string]float64{
	"N": 0, "NE": 45, "E": 90, "SE": 135,
	"S": 180, "SW": 225, "W": 270, "NW": 315,
}

func compassPointToDeg(p string) float64 {
	return compassDegrees[p]
}

func parseInt32(s string) int32 {
	v, _ := strconv.Atoi(s)
	return int32(v)
}
