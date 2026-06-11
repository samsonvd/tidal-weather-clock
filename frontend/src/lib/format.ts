import type { WeatherCode } from '@/types/api'

export function msToKnots(ms: number): number {
  return Math.round(ms * 1.94384)
}

export function knotsToMs(kts: number): number {
  return kts / 1.94384
}

export function msToBeaufort(ms: number): number {
  const thresholds = [0.3, 1.6, 3.4, 5.5, 8.0, 10.8, 13.9, 17.2, 20.8, 24.5, 28.5, 32.7]
  const b = thresholds.findIndex(t => ms < t)
  return b === -1 ? 12 : b
}

const COMPASS = ['N','NNE','NE','ENE','E','ESE','SE','SSE','S','SSW','SW','WSW','W','WNW','NW','NNW']

export function degreesToCompass(deg: number): string {
  return COMPASS[Math.round(((deg % 360) + 360) % 360 / 22.5) % 16]
}

export const weatherLabels: Record<WeatherCode, string> = {
  clear: 'Clear',
  partly_cloudy: 'Partly Cloudy',
  overcast: 'Overcast',
  fog: 'Fog',
  drizzle: 'Drizzle',
  rain: 'Rain',
  heavy_rain: 'Heavy Rain',
  snow: 'Snow',
  thunderstorm: 'Thunderstorm',
}

export function localDateStr(): string {
  return new Date().toLocaleDateString('en-CA')
}

export function addDays(dateStr: string, n: number): string {
  const d = new Date(dateStr + 'T12:00:00')
  d.setDate(d.getDate() + n)
  return d.toLocaleDateString('en-CA')
}

export function formatTime(iso: string | Date): string {
  return new Date(iso).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
}

export function formatDateLabel(dateStr: string): string {
  return new Date(dateStr + 'T12:00:00').toLocaleDateString('en-GB', {
    weekday: 'short', day: 'numeric', month: 'short',
  })
}

export function scorePercent(score: number): number {
  return Math.round(score * 100)
}
