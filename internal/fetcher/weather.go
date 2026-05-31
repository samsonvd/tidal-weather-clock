package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/samson/tidal-weather-clock/internal/domain"
)

type HourlyWeather struct {
	Time        time.Time
	WindSpeedMS float64
	WindDirDeg  float64
	Weather     domain.WeatherCode
}

type openMeteoResponse struct {
	Hourly struct {
		Time             []string  `json:"time"`
		WindSpeedMS      []float64 `json:"wind_speed_10m"`
		WindDirDeg       []float64 `json:"wind_direction_10m"`
		WeatherCode      []int     `json:"weather_code"`
	} `json:"hourly"`
}

func FetchWeather(ctx context.Context, lat, lon float64, days int) ([]HourlyWeather, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=wind_speed_10m,wind_direction_10m,weather_code&wind_speed_unit=ms&timezone=UTC&forecast_days=%d",
		lat, lon, days,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open-meteo returned %d", resp.StatusCode)
	}

	var body openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	n := len(body.Hourly.Time)
	results := make([]HourlyWeather, 0, n)
	for i := 0; i < n; i++ {
		t, err := time.Parse("2006-01-02T15:04", body.Hourly.Time[i])
		if err != nil {
			return nil, fmt.Errorf("parse time %q: %w", body.Hourly.Time[i], err)
		}
		results = append(results, HourlyWeather{
			Time:        t.UTC(),
			WindSpeedMS: body.Hourly.WindSpeedMS[i],
			WindDirDeg:  body.Hourly.WindDirDeg[i],
			Weather:     wmoToWeatherCode(body.Hourly.WeatherCode[i]),
		})
	}
	return results, nil
}

func wmoToWeatherCode(wmo int) domain.WeatherCode {
	switch {
	case wmo == 0:
		return domain.WeatherClear
	case wmo <= 2:
		return domain.WeatherPartlyCloudy
	case wmo == 3:
		return domain.WeatherOvercast
	case wmo == 45 || wmo == 48:
		return domain.WeatherFog
	case wmo >= 51 && wmo <= 57:
		return domain.WeatherDrizzle
	case wmo >= 61 && wmo <= 63:
		return domain.WeatherRain
	case wmo == 65 || wmo == 82:
		return domain.WeatherHeavyRain
	case wmo >= 66 && wmo <= 67:
		return domain.WeatherRain
	case wmo >= 71 && wmo <= 77:
		return domain.WeatherSnow
	case wmo >= 80 && wmo <= 81:
		return domain.WeatherRain
	case wmo >= 85 && wmo <= 86:
		return domain.WeatherSnow
	case wmo >= 95:
		return domain.WeatherThunderstorm
	default:
		return domain.WeatherOvercast
	}
}
