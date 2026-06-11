export interface User {
  id: string
  email: string
}

export type WeatherCode =
  | 'clear'
  | 'partly_cloudy'
  | 'overcast'
  | 'fog'
  | 'drizzle'
  | 'rain'
  | 'heavy_rain'
  | 'snow'
  | 'thunderstorm'

export interface HourlyData {
  time: string
  wind_speed_ms: number
  wind_dir_deg: number
  weather_code: WeatherCode
  tide_height_m: number
}

export type ConstraintType = 'wind_speed' | 'wind_dir' | 'weather' | 'tide_height'

export interface Constraint {
  type: ConstraintType
  required: boolean
  weight: number
  ideal_min?: number
  ideal_max?: number
  acceptable_min?: number
  acceptable_max?: number
  preferred?: number[]
  tolerance?: number
  acceptable?: string[]
  ideal?: string[]
}

export interface Activity {
  id: string
  user_id: string
  name: string
  duration_hrs: number
  window_start: number
  window_end: number
  constraints: Constraint[]
  created_at: string
  updated_at: string
}

export interface ConstraintResult {
  constraint: Constraint
  hour_scores: number[]
  score: number
  passed: boolean
}

export interface ScoredWindow {
  start: string
  end: string
  score: number
  excluded: boolean
  activity: Activity
  hours: HourlyData[]
  constraint_results: ConstraintResult[]
}

export interface Location {
  id: string
  name: string
  lat: number
  lon: number
  tide_station_id: string
}

export interface DayData {
  date: string
  location: Location | null
  hours: HourlyData[]
  windows: ScoredWindow[]
}

export interface Station {
  id: string
  name: string
  country: string
  lat: number
  lon: number
}

export interface StationGroup {
  country: string
  stations: Station[]
}
