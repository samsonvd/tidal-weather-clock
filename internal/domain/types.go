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
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Name        string       `json:"name"`
	DurationHrs int          `json:"duration_hrs"`
	WindowStart int          `json:"window_start"`
	WindowEnd   int          `json:"window_end"`
	Constraints []Constraint `json:"constraints"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type HourlyData struct {
	ID          uuid.UUID   `json:"id"`
	LocationID  uuid.UUID   `json:"location_id"`
	Time        time.Time   `json:"time"`
	WindSpeedMS float64     `json:"wind_speed_ms"`
	WindDirDeg  float64     `json:"wind_dir_deg"`
	Weather     WeatherCode `json:"weather_code"`
	TideHeightM float64     `json:"tide_height_m"`
	FetchedAt   time.Time   `json:"fetched_at"`
}

type Location struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Lat           float64   `json:"lat"`
	Lon           float64   `json:"lon"`
	TideStationID string    `json:"tide_station_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
