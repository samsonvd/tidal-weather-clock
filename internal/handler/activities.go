package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samson/tidal-weather-clock/internal/auth"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
)

type ActivityHandler struct {
	db *db.Queries
}

func NewActivityHandler(queries *db.Queries) *ActivityHandler {
	return &ActivityHandler{db: queries}
}

type activityInput struct {
	Name        string             `json:"name" form:"name"`
	DurationHrs int32              `json:"duration_hrs" form:"duration_hrs"`
	WindowStart int32              `json:"window_start" form:"window_start"`
	WindowEnd   int32              `json:"window_end" form:"window_end"`
	Constraints []domain.Constraint `json:"constraints"`
}

func (h *ActivityHandler) List(c *gin.Context) {
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
	c.JSON(http.StatusOK, activities)
}

func (h *ActivityHandler) Create(c *gin.Context) {
	user := auth.GetUser(c)
	var input activityInput
	if err := c.ShouldBind(&input); err != nil {
		c.String(http.StatusBadRequest, "invalid input")
		return
	}

	constraints, err := json.Marshal(input.Constraints)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	row, err := h.db.CreateActivity(c.Request.Context(), db.CreateActivityParams{
		UserID:      user.ID,
		Name:        input.Name,
		DurationHrs: input.DurationHrs,
		WindowStart: input.WindowStart,
		WindowEnd:   input.WindowEnd,
		Constraints: constraints,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	activity, err := dbActivityToDomain(row)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusCreated, activity)
}

func (h *ActivityHandler) Update(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "invalid id")
		return
	}

	var input activityInput
	if err := c.ShouldBind(&input); err != nil {
		c.String(http.StatusBadRequest, "invalid input")
		return
	}

	constraints, err := json.Marshal(input.Constraints)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	row, err := h.db.UpdateActivity(c.Request.Context(), db.UpdateActivityParams{
		ID:          id,
		UserID:      user.ID,
		Name:        input.Name,
		DurationHrs: input.DurationHrs,
		WindowStart: input.WindowStart,
		WindowEnd:   input.WindowEnd,
		Constraints: constraints,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}

	activity, err := dbActivityToDomain(row)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	c.JSON(http.StatusOK, activity)
}

func (h *ActivityHandler) Delete(c *gin.Context) {
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

func dbActivityToDomain(row db.Activity) (domain.Activity, error) {
	var constraints []domain.Constraint
	if err := json.Unmarshal(row.Constraints, &constraints); err != nil {
		return domain.Activity{}, err
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
