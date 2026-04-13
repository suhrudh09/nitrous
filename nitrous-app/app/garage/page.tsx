'use client'
import { useState, useEffect, useRef } from 'react'
import Nav from '@/components/Nav'
import styles from './garage.module.css'

// ── Static car database ───────────────────────────────────────────────────────
const CAR_DATABASE = [
  {
    id: 'f40',
    make: 'Ferrari',
    model: 'F40',
    year: 1992,
    category: 'SUPERCAR',
    color: 'red',
    icon: '🏎️',
    base: {
      hp: 478, torque: 426, topSpeed: 201, zeroToSixty: 3.8,
      weight: 2425, engine: 'Twin-Turbo V8 2.9L', drivetrain: 'RWD', displacement: 2936,
    },
  },
  {
    id: 'gt3rs',
    make: 'Porsche',
    model: '911 GT3 RS',
    year: 2024,
    category: 'TRACK',
    color: 'orange',
    icon: '🔶',
    base: {
      hp: 518, torque: 343, topSpeed: 184, zeroToSixty: 3.0,
      weight: 3042, engine: 'Flat-6 4.0L NA', drivetrain: 'RWD', displacement: 3996,
    },
  },
  {
    id: 'r35',
    make: 'Nissan',
    model: 'GT-R R35',
    year: 2023,
    category: 'SPORTS',
    color: 'cyan',
    icon: '⚡',
    base: {
      hp: 565, torque: 467, topSpeed: 196, zeroToSixty: 2.9,
      weight: 3960, engine: 'Twin-Turbo V6 3.8L', drivetrain: 'AWD', displacement: 3799,
    },
  },
  {
    id: 'huracan',
    make: 'Lamborghini',
    model: 'Huracán EVO',
    year: 2023,
    category: 'SUPERCAR',
    color: 'gold',
    icon: '🐂',
    base: {
      hp: 630, torque: 443, topSpeed: 202, zeroToSixty: 2.9,
      weight: 3135, engine: 'NA V10 5.2L', drivetrain: 'AWD', displacement: 5204,
    },
  },
  {
    id: 'supra',
    make: 'Toyota',
    model: 'GR Supra',
    year: 2024,
    category: 'SPORTS',
    color: 'blue',
    icon: '🌀',
    base: {
      hp: 382, torque: 368, topSpeed: 155, zeroToSixty: 3.9,
      weight: 3400, engine: 'Turbo I6 3.0L', drivetrain: 'RWD', displacement: 2998,
    },
  },
  {
    id: 'evora',
    make: 'Lotus',
    model: 'Evora GT',
    year: 2022,
    category: 'TRACK',
    color: 'purple',
    icon: '🍃',
    base: {
      hp: 416, torque: 317, topSpeed: 188, zeroToSixty: 3.8,
      weight: 3043, engine: 'S/C V6 3.5L', drivetrain: 'RWD', displacement: 3456,
    },
  },
]

const TUNING_CONFIGS: Record<string, { label: string; short: string; hpMult: number; weightMult: number; topSpeedMult: number; zeroMult: number; torqueMult: number }> = {
  stock:  { label: 'STOCK',      short: 'STOCK',  hpMult: 1.00, weightMult: 1.00, topSpeedMult: 1.00, zeroMult: 1.00, torqueMult: 1.00 },
  street: { label: 'STREET',     short: 'STREET', hpMult: 1.08, weightMult: 0.97, topSpeedMult: 1.04, zeroMult: 0.95, torqueMult: 1.06 },
  track:  { label: 'TRACK',      short: 'TRACK',  hpMult: 1.18, weightMult: 0.90, topSpeedMult: 1.10, zeroMult: 0.86, torqueMult: 1.12 },
  race:   { label: 'RACE SPEC',  short: 'RACE',   hpMult: 1.35, weightMult: 0.82, topSpeedMult: 1.18, zeroMult: 0.76, torqueMult: 1.25 },
  drift:  { label: 'DRIFT',      short: 'DRIFT',  hpMult: 1.20, weightMult: 0.94, topSpeedMult: 0.96, zeroMult: 0.92, torqueMult: 1.30 },
}

const colorMap: Record<string, string> = {
  red:    '#ff2a2a',
  cyan:   '#00e5ff',
  orange: '#fb923c',
  blue:   '#60a5fa',
  purple: '#a78bfa',
  gold:   '#facc15',
}

function StatBar({ label, value, max, color }: { label: string; value: number; max: number; color: string }) {
  const pct = Math.min((value / max) * 100, 100)
  return (
    <div className={styles.statRow}>
      <span className={styles.statLabel}>{label}</span>
      <div className={styles.statBarWrap}>
        <div
          className={styles.statBarFill}
          style={{ width: `${pct}%`, background: color, boxShadow: `0 0 10px ${color}88` }}
        />
      </div>
      <span className={styles.statVal}>{Math.round(value)}</span>
    </div>
  )
}

export default function GaragePage() {
  const [selectedCar, setSelectedCar] = useState(CAR_DATABASE[0])
  const [tuning, setTuning] = useState('stock')
  const [scanLine, setScanLine] = useState(0)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const cfg = TUNING_CONFIGS[tuning]
  const stats = {
    hp:          Math.round(selectedCar.base.hp * cfg.hpMult),
    torque:      Math.round(selectedCar.base.torque * cfg.torqueMult),
    topSpeed:    Math.round(selectedCar.base.topSpeed * cfg.topSpeedMult),
    zeroToSixty: +(selectedCar.base.zeroToSixty * cfg.zeroMult).toFixed(1),
    weight:      Math.round(selectedCar.base.weight * cfg.weightMult),
  }

  const accent = colorMap[selectedCar.color] ?? '#00e5ff'

  const handleCarSelect = (car: typeof CAR_DATABASE[0]) => {
    setSelectedCar(car)
  }

  // Scan line animation
  useEffect(() => {
    let alive = true
    const interval = setInterval(() => {
      if (alive) setScanLine(p => (p + 2) % 100)
    }, 30)
    return () => {
      alive = false
      clearInterval(interval)
    }
  }, [])

  // Draw Tron grid on canvas
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let alive = true

    const w = canvas.offsetWidth
    const h = canvas.offsetHeight
    if (!w || !h) return

    canvas.width  = w
    canvas.height = h

    if (!alive) return
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    const gridSize   = 40
    const accentHex  = accent.startsWith('var') ? '#00e5ff' : accent

    // Flat grid
    ctx.strokeStyle = `${accentHex}20`
    ctx.lineWidth   = 0.5
    for (let x = 0; x <= canvas.width; x += gridSize) {
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, canvas.height); ctx.stroke()
    }
    for (let y = 0; y <= canvas.height; y += gridSize) {
      ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(canvas.width, y); ctx.stroke()
    }

    // Horizon glow — stronger
    const gradient = ctx.createRadialGradient(
      canvas.width / 2, canvas.height * 0.65, 0,
      canvas.width / 2, canvas.height * 0.65, canvas.width * 0.55
    )
    gradient.addColorStop(0, `${accentHex}28`)
    gradient.addColorStop(1, 'transparent')
    ctx.fillStyle = gradient
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    // Perspective lines — stronger opacity
    ctx.strokeStyle = `${accentHex}45`
    ctx.lineWidth   = 1
    const horizon   = canvas.height * 0.52
    const vp        = { x: canvas.width / 2, y: horizon }
    for (let i = -10; i <= 10; i++) {
      const x = canvas.width / 2 + i * 75
      ctx.beginPath(); ctx.moveTo(vp.x, vp.y); ctx.lineTo(x, canvas.height); ctx.stroke()
    }
    for (let i = 0; i <= 8; i++) {
      const t      = i / 8
      const y      = horizon + (canvas.height - horizon) * t
      const spread = t * canvas.width * 0.9
      ctx.beginPath(); ctx.moveTo(vp.x - spread / 2, y); ctx.lineTo(vp.x + spread / 2, y); ctx.stroke()
    }

    return () => { alive = false }
  }, [selectedCar, accent])

  return (
    <>
      <Nav />
      <main className={styles.page}>

        {/* ── Sub-header strip (replaces the overlapping pageTitle) ── */}
        <div className={styles.subHeader}>
          <div className={styles.subHeaderLeft}>
            <span className={styles.breadcrumb}>/ GARAGE</span>
            <span className={styles.subHeaderCar} style={{ color: accent }}>
              {selectedCar.make.toUpperCase()} {selectedCar.model.toUpperCase()}
            </span>
          </div>
          <div className={styles.headerStats}>
            <div className={styles.hStat}>
              <span className={styles.hStatN}>{CAR_DATABASE.length}</span>
              <span className={styles.hStatL}>VEHICLES</span>
            </div>
            <div className={styles.hStatDiv} />
            <div className={styles.hStat}>
              <span className={styles.hStatN}>{Object.keys(TUNING_CONFIGS).length}</span>
              <span className={styles.hStatL}>CONFIGS</span>
            </div>
            <div className={styles.hStatDiv} />
            <div className={styles.hStat}>
              <span className={styles.hStatN} style={{ color: accent }}>{stats.hp}</span>
              <span className={styles.hStatL}>ACTIVE HP</span>
            </div>
          </div>
        </div>

        <div className={styles.garageGrid}>

          {/* ── LEFT: Car selector ── */}
          <div className={styles.selectorPanel}>
            <div className={styles.panelLabel}>SELECT VEHICLE</div>
            {CAR_DATABASE.map((car) => {
              const a       = colorMap[car.color] ?? '#00e5ff'
              const isActive = car.id === selectedCar.id
              return (
                <div
                  key={car.id}
                  className={`${styles.carSlot} ${isActive ? styles.carSlotActive : ''}`}
                  style={isActive ? { borderColor: a, background: `${a}12` } : {}}
                  onClick={() => handleCarSelect(car)}
                >
                  {isActive && (
                    <div className={styles.carSlotAccent} style={{ background: a, boxShadow: `0 0 12px ${a}` }} />
                  )}
                  <span className={styles.carSlotIcon}>{car.icon}</span>
                  <div className={styles.carSlotInfo}>
                    <div className={styles.carSlotMake}>{car.make}</div>
                    <div className={styles.carSlotModel}>{car.model}</div>
                    <div className={styles.carSlotYear}>{car.year}</div>
                  </div>
                  <div
                    className={styles.carSlotBadge}
                    style={{ color: a, borderColor: `${a}55`, background: `${a}10` }}
                  >
                    {car.category}
                  </div>
                </div>
              )
            })}
          </div>

          {/* ── CENTER: Viewport + HUD ── */}
          <div className={styles.viewportPanel}>
            <div className={styles.viewport}>
              <canvas ref={canvasRef} className={styles.gridCanvas} />

              {/* Scan line */}
              <div className={styles.scanLine} style={{ top: `${scanLine}%` }} />

              {/* HUD corner brackets */}
              {(['TL','TR','BL','BR'] as const).map(pos => (
                <div key={pos} className={`${styles.hudCorner} ${styles[`hudCorner${pos}`]}`}>
                  <svg viewBox="0 0 24 24" fill="none">
                    <path
                      d={pos.startsWith('T') ? 'M1 23V5L5 1H23' : 'M1 1V19L5 23H23'}
                      stroke={accent} strokeWidth="2"
                    />
                  </svg>
                </div>
              ))}

              {/* Car display */}
              <div className={styles.carDisplay}>
                <div className={styles.carGlowRing} style={{ boxShadow: `0 0 80px ${accent}55, 0 0 160px ${accent}22` }} />
                <div className={styles.carEmoji} style={{ filter: `drop-shadow(0 0 28px ${accent})` }}>
                  {selectedCar.icon}
                </div>
                <div className={styles.carReflection} style={{ background: `radial-gradient(ellipse 60% 25% at 50% 100%, ${accent}30, transparent)` }} />
              </div>

              {/* HUD overlays */}
              <div className={styles.hudTopLeft} style={{ color: accent }}>
                <div className={styles.hudLine}><span className={styles.hudKey}>MODEL</span><span>{selectedCar.make} {selectedCar.model}</span></div>
                <div className={styles.hudLine}><span className={styles.hudKey}>YEAR</span><span>{selectedCar.year}</span></div>
                <div className={styles.hudLine}><span className={styles.hudKey}>ENGINE</span><span>{selectedCar.base.engine}</span></div>
              </div>
              <div className={styles.hudTopRight} style={{ color: accent }}>
                <div className={styles.hudLine}><span className={styles.hudKey}>DRIVE</span><span>{selectedCar.base.drivetrain}</span></div>
                <div className={styles.hudLine}><span className={styles.hudKey}>CC</span><span>{selectedCar.base.displacement}</span></div>
                <div className={styles.hudLine}><span className={styles.hudKey}>CONFIG</span><span>{TUNING_CONFIGS[tuning].label}</span></div>
              </div>

              {/* Bottom status */}
              <div className={styles.hudBottom}>
                <div className={styles.hudStatus}>
                  <span className={styles.statusDot} style={{ background: accent, boxShadow: `0 0 8px ${accent}` }} />
                  <span style={{ color: accent }}>SYSTEM ONLINE — {TUNING_CONFIGS[tuning].label}</span>
                </div>
              </div>
            </div>

            {/* Instrument cluster strip */}
            <div className={styles.instrumentStrip} style={{ borderColor: `${accent}30` }}>
              <div className={styles.instrItem}>
                <span className={styles.instrVal} style={{ color: accent }}>{stats.topSpeed}</span>
                <span className={styles.instrLabel}>TOP SPEED MPH</span>
              </div>
              <div className={styles.instrDivider} style={{ background: `${accent}30` }} />
              <div className={styles.instrItem}>
                <span className={styles.instrVal} style={{ color: accent }}>{stats.hp}</span>
                <span className={styles.instrLabel}>HORSEPOWER</span>
              </div>
              <div className={styles.instrDivider} style={{ background: `${accent}30` }} />
              <div className={styles.instrItem}>
                <span className={styles.instrVal} style={{ color: accent }}>{stats.zeroToSixty}s</span>
                <span className={styles.instrLabel}>0 – 60 MPH</span>
              </div>
              <div className={styles.instrDivider} style={{ background: `${accent}30` }} />
              <div className={styles.instrItem}>
                <span className={styles.instrVal} style={{ color: accent }}>{stats.weight}</span>
                <span className={styles.instrLabel}>WEIGHT LBS</span>
              </div>
            </div>

            {/* Tuning config selector */}
            <div className={styles.tuningBar}>
              <span className={styles.tuningLabel}>CONFIGURATION</span>
              <div className={styles.tuningBtns}>
                {Object.entries(TUNING_CONFIGS).map(([key, c]) => (
                  <button
                    key={key}
                    className={`${styles.tuningBtn} ${tuning === key ? styles.tuningBtnActive : ''}`}
                    style={tuning === key ? { borderColor: accent, color: accent, background: `${accent}18` } : {}}
                    onClick={() => setTuning(key)}
                  >
                    {c.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* ── RIGHT: Stats panel ── */}
          <div className={styles.statsPanel}>
            <div className={styles.panelLabel}>PERFORMANCE DATA</div>

            <div className={styles.statGroup}>
              <div className={styles.statGroupLabel}>POWER</div>
              <StatBar label="HP"      value={stats.hp}     max={800} color={accent} />
              <StatBar label="TORQUE"  value={stats.torque} max={600} color={accent} />
            </div>

            <div className={styles.statGroup}>
              <div className={styles.statGroupLabel}>DYNAMICS</div>
              <StatBar label="TOP SPD" value={stats.topSpeed} max={250}  color={accent} />
              <StatBar label="WEIGHT"  value={stats.weight}   max={5000} color={accent} />
            </div>

            <div className={styles.statGroup}>
              <div className={styles.statGroupLabel}>ACCELERATION</div>
              <div className={styles.bigStatCard} style={{ borderColor: `${accent}30`, background: `${accent}08` }}>
                <span className={styles.bigStatNum} style={{ color: accent }}>{stats.zeroToSixty}</span>
                <span className={styles.bigStatUnit}>sec 0 → 60</span>
              </div>
            </div>

            {/* Config delta */}
            {tuning !== 'stock' && (
              <div className={styles.deltaPanel}>
                <div className={styles.deltaPanelLabel}>DELTA VS STOCK</div>
                <div className={styles.deltaRow}>
                  <span>POWER</span>
                  <span style={{ color: '#4ade80' }}>+{stats.hp - selectedCar.base.hp} hp</span>
                </div>
                <div className={styles.deltaRow}>
                  <span>WEIGHT</span>
                  <span style={{ color: stats.weight < selectedCar.base.weight ? '#4ade80' : '#ff2a2a' }}>
                    {stats.weight < selectedCar.base.weight ? '−' : '+'}{Math.abs(stats.weight - selectedCar.base.weight)} lbs
                  </span>
                </div>
                <div className={styles.deltaRow}>
                  <span>0 – 60</span>
                  <span style={{ color: '#4ade80' }}>
                    −{(selectedCar.base.zeroToSixty - stats.zeroToSixty).toFixed(1)}s
                  </span>
                </div>
              </div>
            )}

            <button
              className={styles.addBtn}
              style={{ borderColor: accent, color: accent, background: `${accent}10` }}
            >
              + ADD TO MY GARAGE
            </button>
          </div>

        </div>
      </main>
    </>
  )
}