package fetcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/samson/tidal-weather-clock/internal/db"
)

const fetchDays = 7

type Service struct {
	db *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{db: queries}
}

func (s *Service) FetchAllLocations(ctx context.Context) error {
	locations, err := s.db.GetAllLocations(ctx)
	if err != nil {
		return fmt.Errorf("get locations: %w", err)
	}
	if len(locations) == 0 {
		log.Println("no locations found, nothing to fetch")
		return nil
	}
	for _, loc := range locations {
		log.Printf("fetching data for %s", loc.Name)
		if err := s.FetchForLocation(ctx, loc); err != nil {
			return fmt.Errorf("fetch for location %s: %w", loc.Name, err)
		}
	}
	return nil
}

func (s *Service) FetchForLocation(ctx context.Context, loc db.Location) error {
	if err := s.fetchTides(ctx, loc); err != nil {
		return fmt.Errorf("tides: %w", err)
	}
	if err := s.fetchWeather(ctx, loc); err != nil {
		return fmt.Errorf("weather: %w", err)
	}
	return nil
}

func (s *Service) fetchTides(ctx context.Context, loc db.Location) error {
	log.Printf("  fetching tides from EasyTide station %s", loc.TideStationID)
	tides, err := FetchTides(ctx, loc.TideStationID)
	if err != nil {
		return err
	}
	for _, t := range tides {
		if err := s.db.UpsertTideHeight(ctx, db.UpsertTideHeightParams{
			LocationID:  loc.ID,
			Time:        t.Time,
			TideHeightM: t.TideHeightM,
		}); err != nil {
			return fmt.Errorf("upsert tide at %s: %w", t.Time, err)
		}
	}
	log.Printf("  stored %d tide records", len(tides))
	return nil
}

func (s *Service) fetchWeather(ctx context.Context, loc db.Location) error {
	log.Printf("  fetching weather from OpenMeteo")
	weather, err := FetchWeather(ctx, loc.Lat, loc.Lon, fetchDays)
	if err != nil {
		return err
	}
	for _, w := range weather {
		if err := s.db.UpsertWeatherData(ctx, db.UpsertWeatherDataParams{
			LocationID:         loc.ID,
			Time:               w.Time,
			WindSpeedMs:        w.WindSpeedMS,
			WindDirDeg:         w.WindDirDeg,
			WeatherCode:        string(w.Weather),
			TemperatureCelsius: w.Temperature,
		}); err != nil {
			return fmt.Errorf("upsert weather at %s: %w", w.Time, err)
		}
	}
	log.Printf("  stored %d weather records", len(weather))
	return nil
}

// WindowStart returns midnight UTC today.
func WindowStart() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
