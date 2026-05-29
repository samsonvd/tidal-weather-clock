package domain

import (
	"time"

	"github.com/google/uuid"
)

type WindUnit string

const (
	UnitMS       WindUnit = "m/s"
	UnitKnots    WindUnit = "kts"
	UnitBeaufort WindUnit = "beaufort"
)

type WeatherCode string

const (
	WeatherClear        WeatherCode = "clear"
	WeatherPartlyCloudy WeatherCode = "partly_cloudy"
	WeatherOvercast     WeatherCode = "overcast"
	WeatherFog          WeatherCode = "fog"
	WeatherDrizzle      WeatherCode = "drizzle"
	WeatherRain         WeatherCode = "rain"
	WeatherHeavyRain    WeatherCode = "heavy_rain"
	WeatherSnow         WeatherCode = "snow"
	WeatherThunderstorm WeatherCode = "thunderstorm"
)

type ConstraintType string

const (
	ConstraintWindSpeed  ConstraintType = "wind_speed"
	ConstraintWindDir    ConstraintType = "wind_dir"
	ConstraintWeather    ConstraintType = "weather"
	ConstraintTideHeight ConstraintType = "tide_height"
)

type Constraint struct {
	Type     ConstraintType `json:"type"`
	Required bool           `json:"required"`
	Weight   float64        `json:"weight"`

	// wind_speed, tide_height
	IdealMin      float64 `json:"ideal_min,omitempty"`
	IdealMax      float64 `json:"ideal_max,omitempty"`
	AcceptableMin float64 `json:"acceptable_min,omitempty"`
	AcceptableMax float64 `json:"acceptable_max,omitempty"`

	// wind_dir
	Preferred []float64 `json:"preferred,omitempty"`
	Tolerance float64   `json:"tolerance,omitempty"`

	// weather
	Acceptable []string `json:"acceptable,omitempty"`
	Ideal      []string `json:"ideal,omitempty"`
}

type Activity struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	DurationHrs int
	WindowStart int
	WindowEnd   int
	Constraints []Constraint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type HourlyData struct {
	ID           uuid.UUID
	LocationID   uuid.UUID
	Time         time.Time
	WindSpeedMS  float64
	WindDirDeg   float64
	Weather      WeatherCode
	TideHeightM  float64
	FetchedAt    time.Time
}

type Location struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Lat       float64
	Lon       float64
	CreatedAt time.Time
}

type User struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
}
