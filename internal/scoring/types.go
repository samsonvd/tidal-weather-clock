package scoring

import (
	"time"

	"github.com/samson/tidal-weather-clock/internal/domain"
)

type ConstraintResult struct {
	Constraint domain.Constraint `json:"constraint"`
	HourScores []float64         `json:"hour_scores"`
	Score      float64           `json:"score"`
	Passed     bool              `json:"passed"`
}

type ScoredWindow struct {
	Start             time.Time          `json:"start"`
	End               time.Time          `json:"end"`
	Score             float64            `json:"score"`
	Excluded          bool               `json:"excluded"`
	Activity          domain.Activity    `json:"activity"`
	Hours             []domain.HourlyData `json:"hours"`
	ConstraintResults []ConstraintResult  `json:"constraint_results"`
}
