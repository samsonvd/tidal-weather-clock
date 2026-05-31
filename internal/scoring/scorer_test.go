package scoring

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samson/tidal-weather-clock/internal/domain"
)

func TestTrapezoid(t *testing.T) {
	tests := []struct {
		val, acceptMin, idealMin, idealMax, acceptMax float64
		want                                          float64
	}{
		{5, 0, 10, 20, 30, 0.5},  // below ideal, in acceptable
		{15, 0, 10, 20, 30, 1.0}, // inside ideal
		{25, 0, 10, 20, 30, 0.5}, // above ideal, in acceptable
		{-1, 0, 10, 20, 30, 0.0}, // below acceptable
		{35, 0, 10, 20, 30, 0.0}, // above acceptable
		{10, 0, 10, 20, 30, 1.0}, // at ideal boundary
		{20, 0, 10, 20, 30, 1.0}, // at ideal boundary
	}
	for _, tt := range tests {
		got := trapezoid(tt.val, tt.acceptMin, tt.idealMin, tt.idealMax, tt.acceptMax)
		if got != tt.want {
			t.Errorf("trapezoid(%v, %v, %v, %v, %v) = %v, want %v",
				tt.val, tt.acceptMin, tt.idealMin, tt.idealMax, tt.acceptMax, got, tt.want)
		}
	}
}

func TestScoreDirection(t *testing.T) {
	tests := []struct {
		dir       float64
		preferred []float64
		tolerance float64
		want      float64
	}{
		{180, []float64{180}, 22.5, 1.0},           // exact match
		{190, []float64{180}, 22.5, 1.0},           // within tolerance
		{210, []float64{180}, 22.5, 0.0},           // outside tolerance
		{5, []float64{350}, 20, 1.0},               // wraparound: 5 is 15deg from 350
		{355, []float64{10}, 20, 1.0},              // wraparound other direction
		{90, []float64{180, 270}, 22.5, 0.0},       // no match
		{270, []float64{180, 270}, 22.5, 1.0},      // matches second preferred
	}
	for _, tt := range tests {
		got := scoreDirection(tt.dir, tt.preferred, tt.tolerance)
		if got != tt.want {
			t.Errorf("scoreDirection(%v, %v, %v) = %v, want %v",
				tt.dir, tt.preferred, tt.tolerance, got, tt.want)
		}
	}
}

func TestFindWindows(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	hours := make([]domain.HourlyData, 24)
	for i := range hours {
		hours[i] = domain.HourlyData{Time: base.Add(time.Duration(i) * time.Hour)}
	}

	activity := domain.Activity{
		DurationHrs: 2,
		WindowStart: 13,
		WindowEnd:   17,
	}

	windows := findWindows(activity, hours)

	// Expected windows: [13,14], [14,15], [15,16], [16,17]
	if len(windows) != 4 {
		t.Fatalf("expected 4 windows, got %d", len(windows))
	}
	if windows[0][0].Time.Hour() != 13 {
		t.Errorf("first window should start at 13:00, got %d", windows[0][0].Time.Hour())
	}
	if windows[3][1].Time.Hour() != 17 {
		t.Errorf("last window should end at 17:00, got %d", windows[3][1].Time.Hour())
	}
}

func TestFindWindowsGap(t *testing.T) {
	base := time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	// 13:00, 14:00, gap (missing 15:00), 16:00, 17:00
	hours := []domain.HourlyData{
		{Time: base},
		{Time: base.Add(time.Hour)},
		{Time: base.Add(3 * time.Hour)},
		{Time: base.Add(4 * time.Hour)},
	}
	activity := domain.Activity{DurationHrs: 2, WindowStart: 13, WindowEnd: 17}
	windows := findWindows(activity, hours)
	// [13,14] valid; [14,16] gap — invalid; [16,17] valid
	if len(windows) != 2 {
		t.Fatalf("expected 2 windows with gap, got %d", len(windows))
	}
}

func TestScoreDayRequiredConstraint(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	hours := make([]domain.HourlyData, 24)
	for i := range hours {
		hours[i] = domain.HourlyData{
			Time:        base.Add(time.Duration(i) * time.Hour),
			WindSpeedMS: 15,
			Weather:     domain.WeatherRain, // always raining
		}
	}

	activity := domain.Activity{
		ID:          uuid.New(),
		Name:        "Surfing",
		DurationHrs: 2,
		WindowStart: 9,
		WindowEnd:   17,
		Constraints: []domain.Constraint{
			{
				Type:       domain.ConstraintWeather,
				Required:   true,
				Acceptable: []string{"clear", "partly_cloudy"},
				Ideal:      []string{"clear"},
			},
		},
	}

	results := ScoreDay([]domain.Activity{activity}, hours)
	for _, w := range results {
		if !w.Excluded {
			t.Errorf("expected all windows excluded due to rain, got score %v", w.Score)
		}
	}
}
