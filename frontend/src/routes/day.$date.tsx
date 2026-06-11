import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { api } from '@/lib/api'
import {
  msToKnots, msToBeaufort, degreesToCompass, weatherLabels,
  formatTime, formatDateLabel, addDays, localDateStr, scorePercent,
} from '@/lib/format'
import type { Constraint, ConstraintResult, HourlyData, ScoredWindow, WeatherCode } from '@/types/api'

export const Route = createFileRoute('/day/$date')({
  component: DayPage,
})

function DayPage() {
  const { date } = Route.useParams()
  const today = localDateStr()
  const prev = addDays(date, -1)
  const next = addDays(date, 1)

  const { data, isLoading } = useQuery({
    queryKey: ['day', date],
    queryFn: () => api.day(date),
  })

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b sticky top-0 z-10 bg-background/95 backdrop-blur">
        <div className="max-w-4xl mx-auto px-4 h-14 flex items-center gap-4">
          <Link to="/day/$date" params={{ date: today }} className="font-semibold text-sm shrink-0">
            Tidal Clock
          </Link>

          <div className="flex items-center gap-1 flex-1 justify-center">
            <Link to="/day/$date" params={{ date: prev }} className="px-2 py-1 text-sm hover:bg-muted rounded-md">←</Link>
            <span className="text-sm font-medium w-36 text-center">{formatDateLabel(date)}</span>
            <Link to="/day/$date" params={{ date: next }} className="px-2 py-1 text-sm hover:bg-muted rounded-md">→</Link>
            {date !== today && (
              <Link to="/day/$date" params={{ date: today }} className="text-xs text-muted-foreground hover:text-foreground ml-1 px-2 py-1 hover:bg-muted rounded-md">
                Today
              </Link>
            )}
          </div>

          <div className="flex items-center gap-3 shrink-0">
            {data?.location && (
              <span className="text-sm text-muted-foreground hidden md:block">{data.location.name}</span>
            )}
            <Link to="/settings" className="text-sm text-muted-foreground hover:text-foreground">Settings</Link>
            <button
              onClick={async () => { await api.logout(); window.location.href = '/login' }}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Sign out
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-6 space-y-8">
        {isLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map(i => <Skeleton key={i} className="h-16 w-full" />)}
          </div>
        ) : !data?.location ? (
          <div className="text-center py-16 text-muted-foreground">
            <p className="mb-2">No location set.</p>
            <Link to="/settings" className="text-sm underline hover:text-foreground">
              Go to Settings to pick a tide station
            </Link>
          </div>
        ) : data.hours.length === 0 ? (
          <div className="text-center py-16 text-muted-foreground">
            No data available for this date.
          </div>
        ) : (
          <>
            <WindowsSection windows={data.windows} />
            <HoursTable hours={data.hours} />
          </>
        )}
      </main>
    </div>
  )
}

function WindowsSection({ windows }: { windows: ScoredWindow[] }) {
  if (windows.length === 0) {
    return (
      <div className="text-muted-foreground text-sm">
        No activities configured. <Link to="/settings" className="underline hover:text-foreground">Add one in Settings.</Link>
      </div>
    )
  }

  const byActivity: Record<string, ScoredWindow[]> = {}
  for (const w of windows) {
    if (!byActivity[w.activity.id]) byActivity[w.activity.id] = []
    byActivity[w.activity.id].push(w)
  }

  return (
    <div className="space-y-6">
      {Object.values(byActivity).map(group => (
        <ActivityGroup key={group[0].activity.id} windows={group} />
      ))}
    </div>
  )
}

function ActivityGroup({ windows }: { windows: ScoredWindow[] }) {
  const activity = windows[0].activity
  const sorted = [...windows].sort((a, b) => {
    if (a.excluded !== b.excluded) return a.excluded ? 1 : -1
    return b.score - a.score
  })

  return (
    <div>
      <div className="flex items-baseline gap-2 mb-2">
        <h2 className="font-semibold">{activity.name}</h2>
        <span className="text-sm text-muted-foreground">
          {activity.duration_hrs}h · {String(activity.window_start).padStart(2, '0')}:00–{String(activity.window_end).padStart(2, '0')}:00
        </span>
      </div>
      <div className="space-y-2">
        {sorted.map((w, i) => <WindowCard key={i} window={w} />)}
      </div>
    </div>
  )
}

function WindowCard({ window: w }: { window: ScoredWindow }) {
  const [expanded, setExpanded] = useState(false)
  const pct = scorePercent(w.score)
  const barColor = w.excluded ? 'bg-muted-foreground/30' : pct >= 70 ? 'bg-green-500' : pct >= 40 ? 'bg-yellow-500' : 'bg-red-500'

  return (
    <div className={`border rounded-lg overflow-hidden ${w.excluded ? 'opacity-60' : ''}`}>
      <button
        className="w-full flex items-center gap-4 px-4 py-3 hover:bg-muted/40 text-left"
        onClick={() => setExpanded(!expanded)}
      >
        <span className="text-sm font-mono shrink-0 w-28">
          {formatTime(w.start)}–{formatTime(w.end)}
        </span>
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <div className="h-1.5 rounded-full bg-muted flex-1 max-w-40">
            <div className={`h-1.5 rounded-full ${barColor}`} style={{ width: `${pct}%` }} />
          </div>
          <span className="text-sm tabular-nums text-muted-foreground">{pct}%</span>
        </div>
        {w.excluded && <Badge variant="outline" className="text-xs shrink-0">Excluded</Badge>}
        <span className="text-muted-foreground text-xs ml-auto">{expanded ? '▲' : '▼'}</span>
      </button>

      {expanded && (
        <div className="border-t bg-muted/20 px-4 py-3 space-y-2">
          {w.constraint_results.map((cr, i) => (
            <ConstraintRow key={i} result={cr} hours={w.hours} />
          ))}
          {w.constraint_results.length === 0 && (
            <p className="text-sm text-muted-foreground">No constraints defined.</p>
          )}
        </div>
      )}
    </div>
  )
}

function ConstraintRow({ result, hours }: { result: ConstraintResult; hours: HourlyData[] }) {
  const c = result.constraint
  const pct = scorePercent(result.score)
  const label = c.type.replace('_', ' ')

  return (
    <div className="grid grid-cols-[7rem_1fr_1fr_3rem] gap-x-4 items-start text-sm">
      <div className="flex items-center gap-1.5">
        <span className="text-muted-foreground capitalize">{label}</span>
        {c.required && <Badge variant="secondary" className="text-xs px-1 py-0">req</Badge>}
      </div>
      <div className="text-muted-foreground">{describeRequested(c)}</div>
      <div>{describeActual(c, hours)}</div>
      <div className={`text-right tabular-nums ${result.passed ? 'text-green-600' : 'text-red-500'}`}>
        {result.passed ? `${pct}%` : '✗'}
      </div>
    </div>
  )
}

function describeRequested(c: Constraint): string {
  switch (c.type) {
    case 'wind_speed':
      return `${msToKnots(c.ideal_min ?? 0)}–${msToKnots(c.ideal_max ?? 0)} kts ideal`
    case 'wind_dir': {
      const dirs = (c.preferred ?? []).map(d => degreesToCompass(d)).join(', ')
      return `${dirs} ±${c.tolerance ?? 0}°`
    }
    case 'weather': {
      const list = (c.ideal ?? c.acceptable ?? []).map(w => weatherLabels[w as WeatherCode] ?? w)
      return list.join(', ') || '—'
    }
    case 'tide_height':
      return `${(c.ideal_min ?? 0).toFixed(1)}–${(c.ideal_max ?? 0).toFixed(1)}m ideal`
    default:
      return '—'
  }
}

function describeActual(c: Constraint, hours: HourlyData[]): string {
  if (!hours.length) return '—'
  switch (c.type) {
    case 'wind_speed': {
      const avg = hours.reduce((s, h) => s + h.wind_speed_ms, 0) / hours.length
      return `${msToKnots(avg)} kts avg`
    }
    case 'wind_dir': {
      const avg = hours.reduce((s, h) => s + h.wind_dir_deg, 0) / hours.length
      return degreesToCompass(avg)
    }
    case 'weather': {
      const counts: Record<string, number> = {}
      hours.forEach(h => { counts[h.weather_code] = (counts[h.weather_code] || 0) + 1 })
      const top = Object.entries(counts).sort((a, b) => b[1] - a[1])[0]
      return top ? (weatherLabels[top[0] as WeatherCode] ?? top[0]) : '—'
    }
    case 'tide_height': {
      const avg = hours.reduce((s, h) => s + h.tide_height_m, 0) / hours.length
      return `${avg.toFixed(1)}m avg`
    }
    default:
      return '—'
  }
}

function HoursTable({ hours }: { hours: HourlyData[] }) {
  return (
    <div>
      <h3 className="text-sm font-medium text-muted-foreground mb-2">Hourly Conditions</h3>
      <div className="border rounded-lg overflow-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50 text-muted-foreground">
              <th className="text-left px-3 py-2 font-medium">Time</th>
              <th className="text-left px-3 py-2 font-medium">Wind</th>
              <th className="text-left px-3 py-2 font-medium">Dir</th>
              <th className="text-left px-3 py-2 font-medium">Weather</th>
              <th className="text-left px-3 py-2 font-medium">Tide</th>
            </tr>
          </thead>
          <tbody>
            {hours.map((h, i) => (
              <tr key={i} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                <td className="px-3 py-2 font-mono">{formatTime(h.time)}</td>
                <td className="px-3 py-2 tabular-nums">
                  {msToKnots(h.wind_speed_ms)} kts{' '}
                  <span className="text-muted-foreground">F{msToBeaufort(h.wind_speed_ms)}</span>
                </td>
                <td className="px-3 py-2">{degreesToCompass(h.wind_dir_deg)}</td>
                <td className="px-3 py-2">{weatherLabels[h.weather_code]}</td>
                <td className="px-3 py-2 tabular-nums">{h.tide_height_m.toFixed(1)}m</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
