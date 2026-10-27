import { ref } from 'vue'

export const updateVisible = ref(false)
export const updateManualRequest = ref(0)
export function openSoftwareUpdate() {
  updateVisible.value = true
  updateManualRequest.value++
}
export interface UpdatePolicy {
  day: string
  attempts: number
  checked: boolean
  prompted: boolean
  lastAttempt: number
  mutedUntil: number
}
export function localDay(now: Date) {
  return `${now.getFullYear()}-${now.getMonth() + 1}-${now.getDate()}`
}
export function normalizePolicy(value: Partial<UpdatePolicy>, now = new Date()): UpdatePolicy {
  const sameDay = value.day === localDay(now)
  return {
    day: localDay(now),
    attempts: sameDay && Number.isFinite(value.attempts) ? Math.max(0, Number(value.attempts)) : 0,
    checked: sameDay && value.checked === true,
    prompted: sameDay && value.prompted === true,
    lastAttempt: sameDay && Number.isFinite(value.lastAttempt) ? Number(value.lastAttempt) : 0,
    mutedUntil: Number.isFinite(value.mutedUntil) ? Number(value.mutedUntil) : 0,
  }
}
export function canAutoCheck(p: UpdatePolicy, now = Date.now()) {
  return (
    !p.checked &&
    !p.prompted &&
    p.attempts < 5 &&
    now >= p.mutedUntil &&
    now - p.lastAttempt >= 60_000
  )
}
const key = 'cm:update:policy:v1' // localStorage is isolated by server origin.
let memory: Partial<UpdatePolicy> = {}
let storageAvailable = true
export function readUpdatePolicy() {
  try {
    if (storageAvailable) memory = JSON.parse(localStorage.getItem(key) || '{}') || {}
  } catch {
    storageAvailable = false
  }
  return normalizePolicy(memory)
}
export function saveUpdatePolicy(value: UpdatePolicy) {
  memory = value
  try {
    if (storageAvailable) localStorage.setItem(key, JSON.stringify(value))
  } catch {
    storageAvailable = false
  }
}
export function setUpdatesMuted(muted: boolean) {
  const p = readUpdatePolicy()
  p.mutedUntil = muted ? Date.now() + 7 * 24 * 60 * 60 * 1000 : 0
  saveUpdatePolicy(p)
}
export interface UpdateInfo {
  currentVersion: string
  version: string
  platform: string
  hasUpdate: boolean
  changelog: string[]
}
export interface UpdateState {
  phase:
    | 'idle'
    | 'downloading'
    | 'ready'
    | 'waiting'
    | 'backingUp'
    | 'installing'
    | 'installed'
    | 'failed'
  version: string
  downloaded: number
  total: number
  error?: string
  job?: string
}
export class UpdateRejectedError extends Error {}
export async function updateRequest<T>(path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  for (const name of ['token', 'adminToken']) {
    const value = sessionStorage.getItem(name)
    if (value) headers[name] = value
  }
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), 12_000)
  try {
    const response = await fetch(`/api/update/${path}`, {
      method: body === undefined ? 'GET' : 'POST',
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    })
    const data = await response.json()
    if (!response.ok || !data.status)
      throw new UpdateRejectedError(data.msg || '更新请求失败，请稍后重试')
    return data.data as T
  } finally {
    clearTimeout(timeout)
  }
}
export async function withUpdateLock(work: () => Promise<void>) {
  if (navigator.locks) {
    await navigator.locks.request('cm:automatic-update', { ifAvailable: true }, async (lock) => {
      if (lock) await work()
    })
    return
  }
  const leaseKey = 'cm:update:lease'
  const token = `${Date.now()}-${Math.random()}`
  try {
    const old = JSON.parse(localStorage.getItem(leaseKey) || '{}')
    if (old.until > Date.now()) return
    localStorage.setItem(leaseKey, JSON.stringify({ token, until: Date.now() + 20_000 }))
    await new Promise((resolve) => setTimeout(resolve, 100))
    if (JSON.parse(localStorage.getItem(leaseKey) || '{}').token !== token) return
  } catch {
    /* LAN HTTP or disabled storage: session fallback */
  }
  try {
    await work()
  } finally {
    try {
      if (JSON.parse(localStorage.getItem(leaseKey) || '{}').token === token)
        localStorage.removeItem(leaseKey)
    } catch {
      /* no storage */
    }
  }
}
