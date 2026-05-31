package scoring

import (
	"time"

	"github.com/samson/tidal-weather-clock/internal/domain"
)

type ConstraintResult struct {
	Constraint domain.Constraint
	HourScores []float64 // one score per hour in the window
	Score      float64   // min of HourScores
	Passed     bool      // false if required and any hour scored 0
}

type ScoredWindow struct {
	Start             time.Time
	End               time.Time // exclusive: start of the hour after the last hour
	Score             float64   // 0.0–1.0; 0 if excluded
	Excluded          bool      // true if any required constraint failed
	Activity          domain.Activity
	ConstraintResults []ConstraintResult
}
