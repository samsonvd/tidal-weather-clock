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
		{180, []float64{180}, 22.5, 1.0},      // exact match
		{190, []float64{180}, 22.5, 1.0},      // within tolerance
		{210, []float64{180}, 22.5, 0.0},      // outside tolerance
		{5, []float64{350}, 20, 1.0},          // wraparound: 5 is 15deg from 350
		{355, []float64{10}, 20, 1.0},         // wraparound other direction
		{90, []float64{180, 270}, 22.5, 0.0},  // no match
		{270, []float64{180, 270}, 22.5, 1.0}, // matches second preferred
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

func TestPreferredConstraintUsesAverage(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// 2-hour window: hour 1 scores 1.0, hour 2 scores 0.0 on the preferred constraint.
	hours := []domain.HourlyData{
		{Time: base.Add(10 * time.Hour), WindSpeedMS: 5},  // inside ideal
		{Time: base.Add(11 * time.Hour), WindSpeedMS: 50}, // way outside acceptable
	}
	activity := domain.Activity{
		ID:          uuid.New(),
		Name:        "Test",
		DurationHrs: 2,
		WindowStart: 10,
		WindowEnd:   11,
		Constraints: []domain.Constraint{
			{
				Type:          domain.ConstraintWindSpeed,
				Required:      false,
				Weight:        1.0,
				AcceptableMin: 0,
				IdealMin:      3,
				IdealMax:      8,
				AcceptableMax: 15,
			},
		},
	}

	results := ScoreDay([]domain.Activity{activity}, hours)
	if len(results) == 0 {
		t.Fatal("expected at least one window")
	}
	w := results[0]
	if w.Excluded {
		t.Fatal("window should not be excluded — constraint is preferred")
	}
	// Average of 1.0 and 0.0 = 0.5; score should not be 0.
	if w.Score == 0 {
		t.Errorf("preferred constraint with one bad hour should not produce score 0, got %v", w.Score)
	}
	if w.Score != 0.5 {
		t.Errorf("expected score 0.5 (average of 1.0 and 0.0), got %v", w.Score)
	}
}

func TestScoreWindowHourScores(t *testing.T) {
	tests := []struct {
		name           string
		hourScores     []float64
		c              domain.Constraint
		expectedScore  float64
		expectedPassed bool
	}{
		{name: "average over optional", hourScores: []float64{1.0, 0.0}, c: domain.Constraint{Required: false}, expectedScore: 0.5, expectedPassed: true},
		{name: "average over optional where all 0", hourScores: []float64{0.0, 0.0}, c: domain.Constraint{Required: false}, expectedScore: 0.0, expectedPassed: true},
		{name: "minimum value over required", hourScores: []float64{1.0, 0.0}, c: domain.Constraint{Required: true}, expectedScore: 0.0, expectedPassed: false},
		{name: "optional empty scores", hourScores: []float64{}, c: domain.Constraint{Required: false}, expectedScore: 0, expectedPassed: true},
		{name: "required empty scores", hourScores: []float64{}, c: domain.Constraint{Required: true}, expectedScore: 0, expectedPassed: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			result := scoreWindowHourScores(test.hourScores, test.c)
			if result.Score != test.expectedScore {
				tt.Errorf("expected score %v, got %v", test.expectedScore, result.Score)
			}
			if result.Passed != test.expectedPassed {
				tt.Errorf("expected passed %v, got %v", test.expectedPassed, result.Passed)
			}
		})
	}
}
