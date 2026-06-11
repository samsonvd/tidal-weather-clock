import { useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api } from '@/lib/api'
import { msToKnots, knotsToMs, degreesToCompass, weatherLabels, localDateStr } from '@/lib/format'
import type { Activity, Constraint, ConstraintType, WeatherCode } from '@/types/api'

export const Route = createFileRoute('/settings')({
  component: SettingsPage,
})

const WEATHER_CODES: WeatherCode[] = [
  'clear', 'partly_cloudy', 'overcast', 'fog', 'drizzle', 'rain', 'heavy_rain', 'snow', 'thunderstorm',
]

const COMPASS_POINTS = [
  { label: 'N', deg: 0 }, { label: 'NE', deg: 45 }, { label: 'E', deg: 90 }, { label: 'SE', deg: 135 },
  { label: 'S', deg: 180 }, { label: 'SW', deg: 225 }, { label: 'W', deg: 270 }, { label: 'NW', deg: 315 },
]

function emptyConstraint(type: ConstraintType): Constraint {
  switch (type) {
    case 'wind_speed': return { type, required: false, weight: 1, ideal_min: 0, ideal_max: 20, acceptable_min: 0, acceptable_max: 30 }
    case 'wind_dir': return { type, required: false, weight: 1, preferred: [], tolerance: 45 }
    case 'weather': return { type, required: false, weight: 1, ideal: ['clear', 'partly_cloudy'], acceptable: ['overcast'] }
    case 'tide_height': return { type, required: false, weight: 1, ideal_min: 0, ideal_max: 3, acceptable_min: 0, acceptable_max: 5 }
  }
}

function emptyActivity(): Omit<Activity, 'id' | 'user_id' | 'created_at' | 'updated_at'> {
  return { name: '', duration_hrs: 2, window_start: 6, window_end: 20, constraints: [] }
}

function SettingsPage() {
  const today = localDateStr()
  const qc = useQueryClient()

  const { data: location } = useQuery({ queryKey: ['location'], queryFn: api.location.get })
  const { data: stations } = useQuery({ queryKey: ['stations'], queryFn: api.location.stations })
  const { data: activities } = useQuery({ queryKey: ['activities'], queryFn: api.activities.list })

  const setLocation = useMutation({
    mutationFn: (id: string) => api.location.set(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['location'] }),
  })

  const createActivity = useMutation({
    mutationFn: api.activities.create,
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['activities'] }); setEditingNew(false) },
  })

  const updateActivity = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Omit<Activity, 'id' | 'user_id' | 'created_at' | 'updated_at'> }) =>
      api.activities.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['activities'] }); setEditingId(null) },
  })

  const deleteActivity = useMutation({
    mutationFn: (id: string) => api.activities.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['activities'] }),
  })

  const [stationPicker, setStationPicker] = useState(false)
  const [editingNew, setEditingNew] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [newForm, setNewForm] = useState(emptyActivity)

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b sticky top-0 z-10 bg-background/95 backdrop-blur">
        <div className="max-w-2xl mx-auto px-4 h-14 flex items-center gap-4">
          <Link to="/day/$date" params={{ date: today }} className="font-semibold text-sm">
            ← Tidal Clock
          </Link>
          <span className="text-sm font-medium">Settings</span>
          <div className="ml-auto">
            <button
              onClick={async () => { await api.logout(); window.location.replace('/#/login') }}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Sign out
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-6 space-y-8">
        {/* Location */}
        <section>
          <h2 className="font-semibold mb-3">Location</h2>
          {location ? (
            <div className="flex items-center gap-3">
              <span className="text-sm">{location.name}</span>
              <Button variant="outline" size="sm" onClick={() => setStationPicker(!stationPicker)}>
                Change
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <span className="text-sm text-muted-foreground">No location set</span>
              <Button variant="outline" size="sm" onClick={() => setStationPicker(!stationPicker)}>
                Pick station
              </Button>
            </div>
          )}

          {stationPicker && (
            <div className="mt-4 border rounded-lg p-4 space-y-4 max-h-96 overflow-y-auto">
              {stations?.map(group => (
                <div key={group.country}>
                  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">{group.country}</p>
                  <div className="space-y-1">
                    {group.stations.map(s => (
                      <button
                        key={s.id}
                        className={`w-full text-left px-3 py-2 text-sm rounded-md hover:bg-muted transition-colors ${location?.tide_station_id === s.id ? 'bg-muted font-medium' : ''}`}
                        onClick={() => { setLocation.mutate(s.id); setStationPicker(false) }}
                      >
                        {s.name}
                      </button>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <Separator />

        {/* Activities */}
        <section>
          <div className="flex items-center justify-between mb-3">
            <h2 className="font-semibold">Activities</h2>
            {!editingNew && (
              <Button size="sm" onClick={() => { setNewForm(emptyActivity()); setEditingNew(true) }}>
                + New
              </Button>
            )}
          </div>

          <div className="space-y-3">
            {editingNew && (
              <div className="border rounded-lg p-4">
                <ActivityForm
                  value={newForm}
                  onChange={setNewForm}
                  onSave={() => createActivity.mutate(newForm)}
                  onCancel={() => setEditingNew(false)}
                  saving={createActivity.isPending}
                />
              </div>
            )}

            {activities?.map(activity => (
              <div key={activity.id} className="border rounded-lg">
                {editingId === activity.id ? (
                  <div className="p-4">
                    <ActivityFormEditing
                      activity={activity}
                      onSave={(data) => updateActivity.mutate({ id: activity.id, data })}
                      onCancel={() => setEditingId(null)}
                      saving={updateActivity.isPending}
                    />
                  </div>
                ) : (
                  <div className="flex items-center gap-3 px-4 py-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-0.5">
                        <span className="font-medium text-sm">{activity.name}</span>
                        <span className="text-xs text-muted-foreground">
                          {activity.duration_hrs}h · {String(activity.window_start).padStart(2,'0')}:00–{String(activity.window_end).padStart(2,'0')}:00
                        </span>
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {activity.constraints.map((c, i) => (
                          <Badge key={i} variant={c.required ? 'default' : 'secondary'} className="text-xs">
                            {c.type.replace('_', ' ')}
                          </Badge>
                        ))}
                      </div>
                    </div>
                    <div className="flex gap-2 shrink-0">
                      <Button variant="ghost" size="sm" onClick={() => setEditingId(activity.id)}>Edit</Button>
                      <Button
                        variant="ghost" size="sm"
                        className="text-destructive hover:text-destructive"
                        onClick={() => { if (confirm(`Delete "${activity.name}"?`)) deleteActivity.mutate(activity.id) }}
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            ))}

            {!activities?.length && !editingNew && (
              <p className="text-sm text-muted-foreground">No activities yet.</p>
            )}
          </div>
        </section>
      </main>
    </div>
  )
}

type ActivityFormData = Omit<Activity, 'id' | 'user_id' | 'created_at' | 'updated_at'>

function ActivityFormEditing({ activity, onSave, onCancel, saving }: {
  activity: Activity
  onSave: (data: ActivityFormData) => void
  onCancel: () => void
  saving: boolean
}) {
  const [form, setForm] = useState<ActivityFormData>({
    name: activity.name,
    duration_hrs: activity.duration_hrs,
    window_start: activity.window_start,
    window_end: activity.window_end,
    constraints: activity.constraints,
  })
  return <ActivityForm value={form} onChange={setForm} onSave={() => onSave(form)} onCancel={onCancel} saving={saving} />
}

function ActivityForm({ value, onChange, onSave, onCancel, saving }: {
  value: ActivityFormData
  onChange: (v: ActivityFormData) => void
  onSave: () => void
  onCancel: () => void
  saving: boolean
}) {
  function updateField<K extends keyof ActivityFormData>(k: K, v: ActivityFormData[K]) {
    onChange({ ...value, [k]: v })
  }

  function addConstraint(type: ConstraintType) {
    onChange({ ...value, constraints: [...value.constraints, emptyConstraint(type)] })
  }

  function updateConstraint(i: number, c: Constraint) {
    const cs = [...value.constraints]
    cs[i] = c
    onChange({ ...value, constraints: cs })
  }

  function removeConstraint(i: number) {
    onChange({ ...value, constraints: value.constraints.filter((_, j) => j !== i) })
  }

  const hours = Array.from({ length: 24 }, (_, i) => i)

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        <div className="col-span-2 space-y-1.5">
          <Label>Name</Label>
          <Input value={value.name} onChange={e => updateField('name', e.target.value)} placeholder="e.g. Kitesurfing" />
        </div>
        <div className="space-y-1.5">
          <Label>Duration (hours)</Label>
          <Input type="number" min={1} max={12} value={value.duration_hrs}
            onChange={e => updateField('duration_hrs', +e.target.value)} />
        </div>
        <div className="space-y-1.5">
          <Label>Window</Label>
          <div className="flex items-center gap-2">
            <Select value={String(value.window_start)} onValueChange={v => updateField('window_start', +v)}>
              <SelectTrigger className="flex-1">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {hours.map(h => <SelectItem key={h} value={String(h)}>{String(h).padStart(2,'0')}:00</SelectItem>)}
              </SelectContent>
            </Select>
            <span className="text-muted-foreground text-sm">–</span>
            <Select value={String(value.window_end)} onValueChange={v => updateField('window_end', +v)}>
              <SelectTrigger className="flex-1">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {hours.map(h => <SelectItem key={h} value={String(h)}>{String(h).padStart(2,'0')}:00</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {/* Constraints */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <Label>Constraints</Label>
          <Select onValueChange={(v) => addConstraint(v as ConstraintType)}>
            <SelectTrigger className="w-36 h-7 text-xs">
              <SelectValue placeholder="+ Add" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="wind_speed">Wind Speed</SelectItem>
              <SelectItem value="wind_dir">Wind Direction</SelectItem>
              <SelectItem value="weather">Weather</SelectItem>
              <SelectItem value="tide_height">Tide Height</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-3">
          {value.constraints.map((c, i) => (
            <ConstraintEditor key={i} constraint={c} onChange={nc => updateConstraint(i, nc)} onRemove={() => removeConstraint(i)} />
          ))}
          {value.constraints.length === 0 && (
            <p className="text-xs text-muted-foreground">No constraints. Add one above.</p>
          )}
        </div>
      </div>

      <div className="flex gap-2 justify-end">
        <Button variant="outline" size="sm" onClick={onCancel}>Cancel</Button>
        <Button size="sm" onClick={onSave} disabled={saving || !value.name.trim()}>
          {saving ? 'Saving…' : 'Save'}
        </Button>
      </div>
    </div>
  )
}

function ConstraintEditor({ constraint: c, onChange, onRemove }: {
  constraint: Constraint
  onChange: (c: Constraint) => void
  onRemove: () => void
}) {
  const label = c.type.replace('_', ' ')

  return (
    <div className="border rounded-md p-3 space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium capitalize">{label}</span>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-1.5 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={c.required}
              onChange={e => onChange({ ...c, required: e.target.checked })}
              className="rounded"
            />
            Required
          </label>
          <button onClick={onRemove} className="text-muted-foreground hover:text-destructive text-xs">Remove</button>
        </div>
      </div>

      {(c.type === 'wind_speed' || c.type === 'tide_height') && (
        <WindSpeedOrTideFields c={c} onChange={onChange} />
      )}

      {c.type === 'wind_dir' && (
        <WindDirFields c={c} onChange={onChange} />
      )}

      {c.type === 'weather' && (
        <WeatherFields c={c} onChange={onChange} />
      )}
    </div>
  )
}

function WindSpeedOrTideFields({ c, onChange }: { c: Constraint; onChange: (c: Constraint) => void }) {
  const isWind = c.type === 'wind_speed'
  const unit = isWind ? 'kts' : 'm'

  function toDisplay(ms: number | undefined) {
    return ms !== undefined ? (isWind ? msToKnots(ms) : ms) : 0
  }
  function fromDisplay(val: number) {
    return isWind ? knotsToMs(val) : val
  }

  return (
    <div className="grid grid-cols-2 gap-2 text-sm">
      <div className="space-y-1">
        <Label className="text-xs">Ideal min ({unit})</Label>
        <Input type="number" min={0} step={isWind ? 1 : 0.1}
          value={toDisplay(c.ideal_min)}
          onChange={e => onChange({ ...c, ideal_min: fromDisplay(+e.target.value) })} />
      </div>
      <div className="space-y-1">
        <Label className="text-xs">Ideal max ({unit})</Label>
        <Input type="number" min={0} step={isWind ? 1 : 0.1}
          value={toDisplay(c.ideal_max)}
          onChange={e => onChange({ ...c, ideal_max: fromDisplay(+e.target.value) })} />
      </div>
      <div className="space-y-1">
        <Label className="text-xs">Acceptable min ({unit})</Label>
        <Input type="number" min={0} step={isWind ? 1 : 0.1}
          value={toDisplay(c.acceptable_min)}
          onChange={e => onChange({ ...c, acceptable_min: fromDisplay(+e.target.value) })} />
      </div>
      <div className="space-y-1">
        <Label className="text-xs">Acceptable max ({unit})</Label>
        <Input type="number" min={0} step={isWind ? 1 : 0.1}
          value={toDisplay(c.acceptable_max)}
          onChange={e => onChange({ ...c, acceptable_max: fromDisplay(+e.target.value) })} />
      </div>
    </div>
  )
}

function WindDirFields({ c, onChange }: { c: Constraint; onChange: (c: Constraint) => void }) {
  function toggleDir(deg: number) {
    const preferred = c.preferred ?? []
    const next = preferred.includes(deg) ? preferred.filter(d => d !== deg) : [...preferred, deg]
    onChange({ ...c, preferred: next })
  }

  return (
    <div className="space-y-2 text-sm">
      <div>
        <Label className="text-xs mb-1.5 block">Preferred directions</Label>
        <div className="flex flex-wrap gap-1.5">
          {COMPASS_POINTS.map(({ label, deg }) => (
            <button
              key={deg}
              onClick={() => toggleDir(deg)}
              className={`px-2 py-1 rounded text-xs border transition-colors ${
                (c.preferred ?? []).includes(deg)
                  ? 'bg-primary text-primary-foreground border-primary'
                  : 'hover:bg-muted border-border'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>
      <div className="space-y-1">
        <Label className="text-xs">Tolerance (°)</Label>
        <Input type="number" min={0} max={180} value={c.tolerance ?? 45}
          onChange={e => onChange({ ...c, tolerance: +e.target.value })}
          className="w-24" />
      </div>
    </div>
  )
}

function WeatherFields({ c, onChange }: { c: Constraint; onChange: (c: Constraint) => void }) {
  function toggle(field: 'ideal' | 'acceptable', code: WeatherCode) {
    const current = (c[field] ?? []) as WeatherCode[]
    const next = current.includes(code) ? current.filter(w => w !== code) : [...current, code]
    onChange({ ...c, [field]: next })
  }

  return (
    <div className="space-y-3 text-sm">
      {(['ideal', 'acceptable'] as const).map(field => (
        <div key={field}>
          <Label className="text-xs capitalize mb-1.5 block">{field} conditions</Label>
          <div className="flex flex-wrap gap-1.5">
            {WEATHER_CODES.map(code => (
              <button
                key={code}
                onClick={() => toggle(field, code)}
                className={`px-2 py-1 rounded text-xs border transition-colors ${
                  ((c[field] ?? []) as string[]).includes(code)
                    ? 'bg-primary text-primary-foreground border-primary'
                    : 'hover:bg-muted border-border'
                }`}
              >
                {weatherLabels[code]}
              </button>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
