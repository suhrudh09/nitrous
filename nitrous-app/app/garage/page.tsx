'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import Nav from '@/components/Nav'
import { useCanTune, useUser } from '@/hooks/usePermission'
import styles from './garage.module.css'

// ── Types ─────────────────────────────────────────────────────────────────────

interface VehicleSpec {
  make: string
  model: string
  year: number
  trim: string
  engine: string
  displacement: number
  cylinders: number
  hp: number
  torque: number
  topSpeed: number
  weight: number
  zeroToSixty: number
  drivetrain: string
  fuelType: string
  seats: number
}

interface TunedStats {
  hp: number
  torque: number
  topSpeed: number
  zeroToSixty: number
  weight: number
  config: string
}

interface TuneResponse {
  base: VehicleSpec
  tuned: TunedStats
  delta: { hp: number; torque: number; topSpeed: number; zeroToSixty: number; weight: number }
  config: TuningConfig
}

interface TuningConfig {
  label: string
  hpMult: number
  torqueMult: number
  topSpeedMult: number
  zeroMult: number
  weightMult: number
}

interface CarEntry {
  make: string
  model: string
  year: number
  category: string
  icon: string
  accentColor: string
}

interface GarageMake {
  id: string
  name: string
}

interface GarageModel {
  id: string
  name: string
  make: string
}

interface SavedVehicleConfig {
  id: string
  make: string
  model: string
  year: number
  engine: string
}

// ── Constants ─────────────────────────────────────────────────────────────────

const API_BASE = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api').replace(/\/$/, '')
const GARAGE_API_BASE = API_BASE.endsWith('/api') ? API_BASE : `${API_BASE}/api`

const VEHICLE_ICONS = ['🏎️', '⚡', '🛞', '🚗', '🔥', '🔧']
const VEHICLE_COLORS = ['#ff2a2a', '#00e5ff', '#facc15', '#60a5fa', '#22c55e', '#fb923c']

const TUNING_KEYS = ['stock', 'street', 'track', 'race', 'drift'] as const
type TuningKey = typeof TUNING_KEYS[number]

const TUNING_LABELS: Record<TuningKey, string> = {
  stock: 'STOCK',
  street: 'STREET',
  track: 'TRACK',
  race: 'RACE SPEC',
  drift: 'DRIFT',
}

const LOCAL_TUNING: Record<TuningKey, TuningConfig> = {
  stock: { label: 'Stock', hpMult: 1.0, torqueMult: 1.0, topSpeedMult: 1.0, zeroMult: 1.0, weightMult: 1.0 },
  street: { label: 'Street', hpMult: 1.08, torqueMult: 1.06, topSpeedMult: 1.04, zeroMult: 0.95, weightMult: 0.97 },
  track: { label: 'Track', hpMult: 1.18, torqueMult: 1.12, topSpeedMult: 1.1, zeroMult: 0.86, weightMult: 0.9 },
  race: { label: 'Race Spec', hpMult: 1.35, torqueMult: 1.25, topSpeedMult: 1.18, zeroMult: 0.76, weightMult: 0.82 },
  drift: { label: 'Drift', hpMult: 1.2, torqueMult: 1.3, topSpeedMult: 0.96, zeroMult: 0.92, weightMult: 0.94 },
}

// ── API helpers ───────────────────────────────────────────────────────────────

async function fetchVehicle(make: string, model: string, year: number): Promise<{ spec: VehicleSpec | null; error: string | null }> {
  try {
    const url = `${GARAGE_API_BASE}/garage/vehicle?make=${encodeURIComponent(make)}&model=${encodeURIComponent(model)}&year=${year}`
    const res = await fetch(url)
    const data = await res.json().catch(() => ({}))

    if (!res.ok) {
      return { spec: null, error: (data as { error?: string }).error ?? `HTTP ${res.status}` }
    }

    return { spec: (data as { vehicle?: VehicleSpec }).vehicle ?? null, error: null }
  } catch {
    return { spec: null, error: 'Network error — is the Go backend running?' }
  }
}

async function postTune(make: string, model: string, year: number, tuning: TuningKey, teamId?: string): Promise<TuneResponse | null> {
  try {
    const token = localStorage.getItem('nitrous_token')
    const headers: Record<string, string> = { 'Content-Type': 'application/json' }
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    
    const body: Record<string, string | number> = { make, model, year, tuning }
    if (teamId) {
      body.teamId = teamId
    }
    
    const res = await fetch(`${GARAGE_API_BASE}/garage/tune`, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    })
    if (!res.ok) {
      const error = await res.json().catch(() => ({}))
      throw new Error(error.error || `Tuning failed (${res.status})`)
    }
    return await res.json()
  } catch (err) {
    console.error('Tune error:', err)
    throw err
  }
}

async function fetchGarageMakes(): Promise<GarageMake[]> {
  try {
    const res = await fetch(`${GARAGE_API_BASE}/garage/makes`)
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray((data as { makes?: GarageMake[] }).makes) ? (data as { makes: GarageMake[] }).makes : []
  } catch {
    return []
  }
}

async function fetchGarageModels(make: string): Promise<GarageModel[]> {
  try {
    const res = await fetch(`${GARAGE_API_BASE}/garage/models?make=${encodeURIComponent(make)}`)
    if (!res.ok) return []
    const data = await res.json()
    return Array.isArray((data as { models?: GarageModel[] }).models) ? (data as { models: GarageModel[] }).models : []
  } catch {
    return []
  }
}

async function fetchGarageYearRange(make: string, model: string): Promise<{ minYear: number; maxYear: number } | null> {
  try {
    const res = await fetch(`${GARAGE_API_BASE}/garage/years?make=${encodeURIComponent(make)}&model=${encodeURIComponent(model)}`)
    if (!res.ok) return null
    const data = await res.json()
    const minYear = Number((data as { minYear?: string | number }).minYear)
    const maxYear = Number((data as { maxYear?: string | number }).maxYear)
    if (!Number.isFinite(minYear) || !Number.isFinite(maxYear)) return null
    return { minYear, maxYear }
  } catch {
    return null
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function localTune(spec: VehicleSpec, key: TuningKey) {
  const cfg = LOCAL_TUNING[key]
  return {
    hp: Math.round(spec.hp * cfg.hpMult),
    torque: Math.round(spec.torque * cfg.torqueMult),
    topSpeed: Math.round(spec.topSpeed * cfg.topSpeedMult),
    zeroToSixty: +(spec.zeroToSixty * cfg.zeroMult).toFixed(1),
    weight: Math.round(spec.weight * cfg.weightMult),
  }
}

function hashCode(input: string) {
  let hash = 0
  for (let i = 0; i < input.length; i++) {
    hash = (hash << 5) - hash + input.charCodeAt(i)
    hash |= 0
  }
  return Math.abs(hash)
}

function buildCarEntry(make: string, model: string, year: number): CarEntry {
  const id = hashCode(`${make}|${model}`)
  return {
    make,
    model,
    year,
    category: 'NHTSA',
    icon: VEHICLE_ICONS[id % VEHICLE_ICONS.length],
    accentColor: VEHICLE_COLORS[id % VEHICLE_COLORS.length],
  }
}

// ── Sub-components ────────────────────────────────────────────────────────────

function StatBar({ label, value, max, accent }: { label: string; value: number; max: number; accent: string }) {
  const pct = Math.min((value / max) * 100, 100)
  return (
    <div className={styles.statRow}>
      <span className={styles.statKey}>{label}</span>
      <div className={styles.barWrap}>
        <div
          className={styles.barFill}
          style={{ width: `${pct}%`, background: accent, boxShadow: `0 0 8px ${accent}55` }}
        />
      </div>
      <span className={styles.statNum}>{value.toLocaleString()}</span>
    </div>
  )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function GaragePage() {
  const [makes, setMakes] = useState<GarageMake[]>([])
  const [models, setModels] = useState<GarageModel[]>([])
  const [selectedMake, setSelectedMake] = useState('')
  const [selectedModel, setSelectedModel] = useState('')
  const [selectedYear, setSelectedYear] = useState<number>(new Date().getFullYear())
  const [minYear, setMinYear] = useState(1980)
  const [maxYear, setMaxYear] = useState(new Date().getFullYear())

  const [loadingMakes, setLoadingMakes] = useState(false)
  const [loadingModels, setLoadingModels] = useState(false)
  const [loadingYears, setLoadingYears] = useState(false)
  const [spec, setSpec] = useState<VehicleSpec | null>(null)
  const [specError, setSpecError] = useState<string | null>(null)
  const [savedConfigs, setSavedConfigs] = useState<SavedVehicleConfig[]>([])
  const [tuning, setTuning] = useState<TuningKey>('stock')
  const [tuneResult, setTuneResult] = useState<TuneResponse | null>(null)
  const [loadingSpec, setLoadingSpec] = useState(false)
  const [selectedTeam, setSelectedTeam] = useState<string>('')
  const [tuneError, setTuneError] = useState<string | null>(null)

  // Permission hooks
  const canTune = useCanTune()
  const { user } = useUser()

  const canvasRef = useRef<HTMLCanvasElement>(null)
  const scanRef = useRef<number>(0)
  const rafRef = useRef<number>(0)
  const pendingSavedRef = useRef<SavedVehicleConfig | null>(null)

  const selected = buildCarEntry(selectedMake, selectedModel, selectedYear)
  const accent = selected.accentColor

  useEffect(() => {
    let cancelled = false
    setLoadingMakes(true)
    fetchGarageMakes().then(ms => {
      if (cancelled) return
      setMakes(ms)
      if (ms.length > 0) setSelectedMake(ms[0].name)
      setLoadingMakes(false)
    })

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!selectedMake) return
    let cancelled = false
    setLoadingModels(true)
    setModels([])
    setSelectedModel('')
    setSpec(null)
    setSpecError(null)
    setTuneResult(null)

    fetchGarageModels(selectedMake).then(ms => {
      if (cancelled) return
      setModels(ms)
      if (ms.length > 0) {
        const pending = pendingSavedRef.current
        if (pending && pending.make === selectedMake) {
          const matched = ms.find(model => model.name === pending.model)
          setSelectedModel(matched ? matched.name : ms[0].name)
        } else {
          setSelectedModel(ms[0].name)
        }
      }
      setLoadingModels(false)
    })

    return () => {
      cancelled = true
    }
  }, [selectedMake])

  useEffect(() => {
    if (!selectedMake || !selectedModel) return
    let cancelled = false
    setLoadingYears(true)
    fetchGarageYearRange(selectedMake, selectedModel).then(range => {
      if (cancelled) return
      if (range) {
        setMinYear(range.minYear)
        setMaxYear(range.maxYear)
        const pending = pendingSavedRef.current
        if (pending && pending.make === selectedMake && pending.model === selectedModel) {
          setSelectedYear(Math.max(range.minYear, Math.min(range.maxYear, pending.year)))
        } else {
          setSelectedYear(range.maxYear)
        }
      }
      setLoadingYears(false)
    })

    return () => {
      cancelled = true
    }
  }, [selectedMake, selectedModel])

  useEffect(() => {
    if (!selectedMake || !selectedModel || !selectedYear) return
    let cancelled = false
    setSpec(null)
    setSpecError(null)
    setTuneResult(null)
    setLoadingSpec(true)

    fetchVehicle(selectedMake, selectedModel, selectedYear).then(({ spec: vehicleSpec, error }) => {
      if (cancelled) return
      setSpec(vehicleSpec)
      setSpecError(error)
      const pending = pendingSavedRef.current
      if (pending && pending.make === selectedMake && pending.model === selectedModel && pending.year === selectedYear) {
        pendingSavedRef.current = null
      }
      setLoadingSpec(false)
    })

    return () => {
      cancelled = true
    }
  }, [selectedMake, selectedModel, selectedYear])

  useEffect(() => {
    if (!spec || tuning === 'stock') {
      setTuneResult(null)
      setTuneError(null)
      return
    }

    let cancelled = false

    postTune(selectedMake, selectedModel, selectedYear, tuning, selectedTeam || undefined)
      .then(result => {
        if (!cancelled) {
          setTuneResult(result)
          setTuneError(null)
        }
      })
      .catch(err => {
        if (!cancelled) {
          setTuneResult(null)
          setTuneError(err instanceof Error ? err.message : 'Tuning failed')
        }
      })

    return () => {
      cancelled = true
    }
  }, [spec, tuning, selectedMake, selectedModel, selectedYear, selectedTeam])

  const drawGrid = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const width = canvas.offsetWidth
    const height = canvas.offsetHeight
    if (!width || !height) return
    canvas.width = width
    canvas.height = height

    ctx.clearRect(0, 0, width, height)
    const gridSize = 38

    ctx.strokeStyle = `${accent}18`
    ctx.lineWidth = 0.5
    for (let x = 0; x <= width; x += gridSize) {
      ctx.beginPath()
      ctx.moveTo(x, 0)
      ctx.lineTo(x, height)
      ctx.stroke()
    }
    for (let y = 0; y <= height; y += gridSize) {
      ctx.beginPath()
      ctx.moveTo(0, y)
      ctx.lineTo(width, y)
      ctx.stroke()
    }

    const horizon = height * 0.52
    const vanishingPoint = width / 2

    ctx.strokeStyle = `${accent}30`
    ctx.lineWidth = 0.8
    for (let i = -14; i <= 14; i++) {
      const x = vanishingPoint + i * 58
      ctx.beginPath()
      ctx.moveTo(vanishingPoint, horizon)
      ctx.lineTo(x, height)
      ctx.stroke()
    }
    for (let i = 0; i <= 10; i++) {
      const t = i / 10
      const y = horizon + (height - horizon) * t
      const span = t * width * 0.9
      ctx.beginPath()
      ctx.moveTo(vanishingPoint - span / 2, y)
      ctx.lineTo(vanishingPoint + span / 2, y)
      ctx.stroke()
    }

    const gradient = ctx.createRadialGradient(vanishingPoint, height * 0.6, 0, vanishingPoint, height * 0.6, width * 0.5)
    gradient.addColorStop(0, `${accent}1a`)
    gradient.addColorStop(1, 'transparent')
    ctx.fillStyle = gradient
    ctx.fillRect(0, 0, width, height)
  }, [accent])

  useEffect(() => {
    drawGrid()
  }, [drawGrid])

  useEffect(() => {
    const resizeObserver = new ResizeObserver(drawGrid)
    if (canvasRef.current) resizeObserver.observe(canvasRef.current.parentElement!)
    return () => resizeObserver.disconnect()
  }, [drawGrid])

  useEffect(() => {
    const tick = () => {
      scanRef.current = (scanRef.current + 0.35) % 100
      const el = document.getElementById('scan-line')
      if (el) el.style.top = `${scanRef.current}%`
      rafRef.current = requestAnimationFrame(tick)
    }
    rafRef.current = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(rafRef.current)
  }, [])

  const displaySpec = spec ?? {
    hp: 0,
    torque: 0,
    topSpeed: 0,
    zeroToSixty: 0,
    weight: 0,
    engine: '—',
    drivetrain: '—',
    displacement: 0,
  }

  const tuned = spec ? (tuneResult ? tuneResult.tuned : localTune(spec, tuning)) : null
  const activeHP = tuned?.hp ?? displaySpec.hp
  const activeTorq = tuned?.torque ?? displaySpec.torque
  const activeSpd = tuned?.topSpeed ?? displaySpec.topSpeed
  const activeZero = tuned?.zeroToSixty ?? displaySpec.zeroToSixty
  const activeWt = tuned?.weight ?? displaySpec.weight
  const delta = tuneResult?.delta ?? null

  const saveCurrentConfig = () => {
    if (!selectedMake || !selectedModel || !selectedYear) return
    const engineValue = spec?.engine || 'N/A'

    const config: SavedVehicleConfig = {
      id: `${selectedMake}|${selectedModel}|${selectedYear}|${engineValue}`,
      make: selectedMake,
      model: selectedModel,
      year: selectedYear,
      engine: engineValue,
    }

    setSavedConfigs(prev => {
      if (prev.some(saved => saved.id === config.id)) return prev
      return [config, ...prev].slice(0, 10)
    })
  }

  const loadSavedConfig = (saved: SavedVehicleConfig) => {
    pendingSavedRef.current = saved
    setSelectedMake(saved.make)
    setSelectedModel(saved.model)
    setSelectedYear(saved.year)
    setTuning('stock')
  }

  const statusText = loadingSpec
    ? 'FETCHING VEHICLE DATA…'
    : specError
      ? `DATA ERROR — ${specError}`
      : tuneError
        ? `TUNING ERROR — ${tuneError}`
        : `SYSTEM ONLINE — ${TUNING_LABELS[tuning]}`

  // Show access denied message for non-managers
  const showAccessDenied = !canTune && user && user.role !== 'admin'

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Role-based access banner */}
        {showAccessDenied && (
          <div className={styles.accessBanner}>
            <span className={styles.accessIcon}>🔒</span>
            <span>TUNING UNAVAILABLE — You need Manager or Admin role to access vehicle tuning</span>
          </div>
        )}

        <div className={styles.subHeader}>
          <div className={styles.subHeaderLeft}>
            <span className={styles.breadcrumb}>/ GARAGE</span>
            <span className={styles.subHeaderCar} style={{ color: accent }}>
              {selectedMake ? selectedMake.toUpperCase() : '—'} {selectedModel ? selectedModel.toUpperCase() : ''}
            </span>
            {(loadingSpec || loadingMakes || loadingModels || loadingYears) && (
              <span className={styles.loadingPip} style={{ background: accent }} />
            )}
          </div>
          <div className={styles.headerStats}>
            <div className={styles.hStat}>
              <span className={styles.hStatN}>{models.length}</span>
              <span className={styles.hStatL}>VEHICLES</span>
            </div>
            <div className={styles.hStatDiv} />
            <div className={styles.hStat}>
              <span className={styles.hStatN}>{TUNING_KEYS.length}</span>
              <span className={styles.hStatL}>CONFIGS</span>
            </div>
            <div className={styles.hStatDiv} />
            <div className={styles.hStat}>
              <span className={styles.hStatN} style={{ color: accent }}>{activeHP || '—'}</span>
              <span className={styles.hStatL}>ACTIVE HP</span>
            </div>
          </div>
        </div>

        <div className={styles.garageGrid}>
          <div className={styles.selectorPanel}>
            <div className={styles.panelLabel}>SELECT VEHICLE</div>

            <div className={styles.hStatL}>MAKE</div>
            <div className={styles.searchWrap}>
              <select
                className={styles.searchInput}
                value={selectedMake}
                onChange={e => {
                  setSelectedMake(e.target.value)
                  setTuning('stock')
                }}
              >
                {makes.map(make => (
                  <option key={make.id} value={make.name}>{make.name}</option>
                ))}
              </select>
            </div>

            <div className={styles.hStatL}>MODEL</div>
            <div className={styles.searchWrap}>
              <select
                className={styles.searchInput}
                value={selectedModel}
                onChange={e => {
                  setSelectedModel(e.target.value)
                  setTuning('stock')
                }}
              >
                {models.map(model => (
                  <option key={model.id} value={model.name}>{model.name}</option>
                ))}
              </select>
            </div>

            <div className={styles.hStatL}>YEAR</div>
            <div className={styles.searchWrap}>
              <input
                className={styles.searchInput}
                type="number"
                min={minYear}
                max={maxYear}
                value={selectedYear}
                onChange={e => {
                  const next = Number(e.target.value)
                  if (Number.isFinite(next)) {
                    setSelectedYear(Math.max(minYear, Math.min(maxYear, next)))
                    setTuning('stock')
                  }
                }}
              />
            </div>

            <button
              className={styles.tuningBtn}
              style={{ width: '100%', marginTop: 8, borderColor: accent, color: accent, background: `${accent}15` }}
              onClick={saveCurrentConfig}
              disabled={!selectedMake || !selectedModel || !selectedYear}
            >
              SAVE CONFIGURATION
            </button>

            {savedConfigs.length > 0 && (
              <>
                <div className={styles.panelLabel} style={{ marginTop: 14 }}>SAVED CONFIGS</div>
                <div className={styles.carList}>
                  {savedConfigs.map(saved => {
                    const savedCar = buildCarEntry(saved.make, saved.model, saved.year)
                    const isActive = selectedMake === saved.make && selectedModel === saved.model && selectedYear === saved.year

                    return (
                      <div
                        key={saved.id}
                        className={`${styles.carSlot} ${isActive ? styles.carSlotActive : ''}`}
                        style={isActive ? { borderColor: `${savedCar.accentColor}60`, background: `${savedCar.accentColor}0d` } : {}}
                        onClick={() => loadSavedConfig(saved)}
                      >
                        {isActive && (
                          <div
                            className={styles.carSlotAccentBar}
                            style={{ background: savedCar.accentColor, boxShadow: `0 0 10px ${savedCar.accentColor}` }}
                          />
                        )}
                        <span className={styles.carSlotIcon}>{savedCar.icon}</span>
                        <div className={styles.carSlotInfo}>
                          <div className={styles.carSlotMake}>{saved.make}</div>
                          <div className={styles.carSlotModel}>{saved.model}</div>
                          <div className={styles.carSlotYear}>{saved.year} • {saved.engine}</div>
                        </div>
                        <div
                          className={styles.catBadge}
                          style={{ color: savedCar.accentColor, borderColor: `${savedCar.accentColor}44`, background: `${savedCar.accentColor}12` }}
                        >
                          SAVED
                        </div>
                      </div>
                    )
                  })}
                </div>
              </>
            )}
          </div>

          <div className={styles.viewportPanel}>
            <div className={styles.viewport}>
              <canvas ref={canvasRef} className={styles.gridCanvas} />
              <div id="scan-line" className={styles.scanLine} style={{ background: `linear-gradient(90deg,transparent,${accent}50,transparent)` }} />

              {(['TL', 'TR', 'BL', 'BR'] as const).map(pos => (
                <div key={pos} className={`${styles.hudCorner} ${styles[`hudCorner${pos}` as keyof typeof styles]}`}>
                  <svg viewBox="0 0 22 22" fill="none" width="22" height="22">
                    <path
                      d={pos.startsWith('T') ? 'M1 21V4L4 1H21' : 'M1 1V18L4 21H21'}
                      stroke={accent}
                      strokeWidth="1.5"
                    />
                  </svg>
                </div>
              ))}

              <div className={styles.hudTL} style={{ color: accent }}>
                <div className={styles.hudLine}><span className={styles.hudK}>MODEL</span><span>{selected.make} {selected.model}</span></div>
                <div className={styles.hudLine}><span className={styles.hudK}>YEAR</span><span>{selectedYear}</span></div>
                <div className={styles.hudLine}><span className={styles.hudK}>ENGINE</span><span>{displaySpec.engine}</span></div>
              </div>
              <div className={styles.hudTR} style={{ color: accent }}>
                <div className={styles.hudLine}><span>{displaySpec.drivetrain}</span><span className={styles.hudK}>DRIVE</span></div>
                <div className={styles.hudLine}><span>{displaySpec.displacement ? `${displaySpec.displacement}cc` : '—'}</span><span className={styles.hudK}>CC</span></div>
                <div className={styles.hudLine}><span>{TUNING_LABELS[tuning]}</span><span className={styles.hudK}>CONFIG</span></div>
              </div>

              <div className={styles.carDisplay}>
                <div className={styles.glowRing} style={{ boxShadow: `0 0 80px ${accent}44, 0 0 160px ${accent}18` }} />
                <div className={styles.carEmoji} style={{ filter: `drop-shadow(0 0 30px ${accent})` }}>
                  {selected.icon}
                </div>
                <div
                  className={styles.carReflection}
                  style={{ background: `radial-gradient(ellipse 60% 20% at 50% 100%, ${accent}28, transparent)` }}
                />
              </div>

              <div className={styles.hudBottom}>
                <span
                  className={styles.statusDot}
                  style={{ background: specError ? '#ff4444' : accent, boxShadow: `0 0 6px ${specError ? '#ff4444' : accent}` }}
                />
                <span style={{ color: specError ? '#ff4444' : accent }}>{statusText}</span>
              </div>
            </div>

            <div className={styles.instruments} style={{ borderColor: `${accent}25` }}>
              {[
                { val: activeSpd, label: 'TOP SPEED MPH' },
                { val: activeHP, label: 'HORSEPOWER' },
                { val: activeZero ? `${activeZero}s` : '—', label: '0 – 60 MPH' },
                { val: activeWt ? activeWt.toLocaleString() : '—', label: 'WEIGHT LBS' },
              ].map(({ val, label }) => (
                <div key={label} className={styles.instrItem}>
                  <span className={styles.instrVal} style={{ color: accent }}>{val || '—'}</span>
                  <span className={styles.instrLabel}>{label}</span>
                </div>
              ))}
            </div>

            <div className={styles.tuningBar}>
              <span className={styles.tuningLabel}>CONFIGURATION</span>
              <div className={styles.tuningBtns}>
                {TUNING_KEYS.map(key => (
                  <button
                    key={key}
                    className={`${styles.tuningBtn} ${tuning === key ? styles.tuningBtnActive : ''}`}
                    style={tuning === key ? { borderColor: accent, color: accent, background: `${accent}15` } : {}}
                    onClick={() => setTuning(key)}
                    disabled={!canTune && !showAccessDenied}
                  >
                    {TUNING_LABELS[key]}
                  </button>
                ))}
              </div>
              {canTune && (
                <div className={styles.teamSelector}>
                  <span className={styles.teamSelectorLabel}>TEAM</span>
                  <select 
                    value={selectedTeam} 
                    onChange={(e) => setSelectedTeam(e.target.value)}
                  >
                    <option value="">None</option>
                    <option value="team-1">Team Alpha</option>
                    <option value="team-2">Team Beta</option>
                  </select>
                </div>
              )}
            </div>
          </div>

          <div className={styles.statsPanel}>
            <div className={styles.panelLabel}>PERFORMANCE DATA</div>

            <div className={styles.statGroup}>
              <div className={styles.statGroupLabel}>POWER</div>
              <StatBar label="HP" value={activeHP} max={800} accent={accent} />
              <StatBar label="TORQUE" value={activeTorq} max={600} accent={accent} />
            </div>

            <div className={styles.statGroup}>
              <div className={styles.statGroupLabel}>DYNAMICS</div>
              <StatBar label="TOP SPD" value={activeSpd} max={250} accent={accent} />
              <StatBar label="WEIGHT" value={activeWt} max={5000} accent={accent} />
            </div>

            <div className={styles.statGroup}>
              <div className={styles.statGroupLabel}>ACCELERATION</div>
              <div className={styles.bigStatCard} style={{ borderColor: `${accent}30`, background: `${accent}08` }}>
                <span className={styles.bigStatNum} style={{ color: accent }}>{activeZero || '—'}</span>
                <span className={styles.bigStatUnit}>sec 0 → 60</span>
              </div>
            </div>

            {tuning !== 'stock' && spec && (
              <div className={styles.deltaPanel}>
                <div className={styles.deltaPanelLabel}>
                  DELTA VS STOCK
                  {!tuneResult && <span className={styles.deltaEstimate}> (est.)</span>}
                </div>
                <div className={styles.deltaRow}>
                  <span>POWER</span>
                  <span className={styles.pos}>+{activeHP - spec.hp} hp</span>
                </div>
                <div className={styles.deltaRow}>
                  <span>WEIGHT</span>
                  <span className={activeWt < spec.weight ? styles.pos : styles.neg}>
                    {activeWt < spec.weight ? '−' : '+'}{Math.abs(activeWt - spec.weight).toLocaleString()} lbs
                  </span>
                </div>
                <div className={styles.deltaRow}>
                  <span>0–60</span>
                  <span className={styles.pos}>−{(spec.zeroToSixty - activeZero).toFixed(1)}s</span>
                </div>
                {delta && (
                  <div className={styles.deltaRow}>
                    <span>TOP SPEED</span>
                    <span className={styles.pos}>+{delta.topSpeed} mph</span>
                  </div>
                )}
              </div>
            )}

            <div className={styles.apiNote}>
              Data via <a href="https://vpic.nhtsa.dot.gov/api/" target="_blank" rel="noreferrer">NHTSA vPIC API</a>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}
