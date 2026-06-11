import type { Activity, DayData, Location, StationGroup, User } from '@/types/api'

const BASE = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '')

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    ...options,
  })
  if (res.status === 401) {
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  me: () => request<User>('/api/v1/me'),

  day: (date: string) => request<DayData>(`/api/v1/day/${date}`),

  activities: {
    list: () => request<Activity[]>('/api/v1/activities'),
    create: (data: Omit<Activity, 'id' | 'user_id' | 'created_at' | 'updated_at'>) =>
      request<Activity>('/api/v1/activities', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: Omit<Activity, 'id' | 'user_id' | 'created_at' | 'updated_at'>) =>
      request<Activity>(`/api/v1/activities/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<void>(`/api/v1/activities/${id}`, { method: 'DELETE' }),
  },

  location: {
    get: () => request<Location | null>('/api/v1/locations'),
    set: (stationId: string) =>
      request<Location>('/api/v1/locations', { method: 'PUT', body: JSON.stringify({ station_id: stationId }) }),
    stations: () => request<StationGroup[]>('/api/v1/stations'),
  },

  logout: () => request<void>('/auth/logout', { method: 'POST' }),

  requestLink: (email: string) =>
    request<{ message: string }>('/login', { method: 'POST', body: JSON.stringify({ email }) }),
}
