-- name: UpsertTideHeight :exec
INSERT INTO hourly_data (location_id, time, wind_speed_ms, wind_dir_deg, weather_code, tide_height_m)
VALUES ($1, $2, 0, 0, '', $3)
ON CONFLICT (location_id, time) DO NOTHING;

-- name: UpsertWeatherData :exec
INSERT INTO hourly_data (location_id, time, wind_speed_ms, wind_dir_deg, weather_code, tide_height_m, temperature_celsius)
VALUES ($1, $2, $3, $4, $5, 0, $6)
ON CONFLICT (location_id, time) DO UPDATE SET
    wind_speed_ms       = EXCLUDED.wind_speed_ms,
    wind_dir_deg        = EXCLUDED.wind_dir_deg,
    weather_code        = EXCLUDED.weather_code,
    temperature_celsius = EXCLUDED.temperature_celsius,
    fetched_at          = now();

-- name: GetHourlyDataForLocation :many
SELECT * FROM hourly_data
WHERE location_id = $1
  AND time >= $2
  AND time < $3
ORDER BY time;
