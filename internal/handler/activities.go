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

func (h *ActivityHandler) List(c *gin.Context) {
	user := auth.GetUser(c)
	rows, err := h.db.ListActivities(c.Request.Context(), user.ID)
	if err != nil {
		internalError(c)
		return
	}
	activities, err := dbActivitiesToDomain(rows)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, activities)
}

func (h *ActivityHandler) Get(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "invalid id")
		return
	}
	row, err := h.db.GetActivity(c.Request.Context(), db.GetActivityParams{ID: id, UserID: user.ID})
	if err != nil {
		apiError(c, http.StatusNotFound, "not found")
		return
	}
	activity, err := dbActivityToDomain(row)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, activity)
}

type activityInput struct {
	Name        string              `json:"name"`
	DurationHrs int32               `json:"duration_hrs"`
	WindowStart int32               `json:"window_start"`
	WindowEnd   int32               `json:"window_end"`
	Constraints []domain.Constraint `json:"constraints"`
}

func (h *ActivityHandler) Create(c *gin.Context) {
	user := auth.GetUser(c)
	var input activityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, "invalid input")
		return
	}
	constraints, _ := json.Marshal(input.Constraints)
	row, err := h.db.CreateActivity(c.Request.Context(), db.CreateActivityParams{
		UserID:      user.ID,
		Name:        input.Name,
		DurationHrs: input.DurationHrs,
		WindowStart: input.WindowStart,
		WindowEnd:   input.WindowEnd,
		Constraints: constraints,
	})
	if err != nil {
		internalError(c)
		return
	}
	activity, err := dbActivityToDomain(row)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusCreated, activity)
}

func (h *ActivityHandler) Update(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "invalid id")
		return
	}
	var input activityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apiError(c, http.StatusBadRequest, "invalid input")
		return
	}
	constraints, _ := json.Marshal(input.Constraints)
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
		internalError(c)
		return
	}
	activity, err := dbActivityToDomain(row)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, activity)
}

func (h *ActivityHandler) Delete(c *gin.Context) {
	user := auth.GetUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apiError(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.db.DeleteActivity(c.Request.Context(), db.DeleteActivityParams{ID: id, UserID: user.ID}); err != nil {
		internalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}
