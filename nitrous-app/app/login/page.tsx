'use client'

import { useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Nav from '@/components/Nav'
import { login, register, saveCart } from '@/lib/api'
import type { CartItem } from '@/types'
import styles from './login.module.css'

type Mode = 'signin' | 'signup'
const CART_STORAGE_KEY = 'nitrous_cart_v1'

export default function LoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [mode, setMode] = useState<Mode>(() => (searchParams.get('mode') === 'signup' ? 'signup' : 'signin'))
  const [showPass, setShowPass] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  
  // Unified form state - keeps the Go backend DTO in mind
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    name: '',
    role: 'viewer' as 'viewer' | 'participant' | 'manager' | 'sponsor'
  })

  // Standardized input handler using the 'name' attribute
  const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  // Handle switching between Sign In and Sign Up
  const toggleMode = (newMode: Mode) => {
    setMode(newMode)
    setError('')
    // Clear sensitive fields when switching modes
    setFormData({ email: '', password: '', name: '', role: 'viewer' })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')

    // Veteran tip: Basic client-side validation saves server resources
    if (mode === 'signup' && formData.password.length < 8) {
      setLoading(false)
      return setError('Password must be at least 8 characters')
    }

    try {
      const data =
        mode === 'signin'
          ? await login(formData.email, formData.password)
          : await register(formData.email, formData.password, formData.name, formData.role)

      // Store auth state
      localStorage.setItem('nitrous_token', data.token)
      if (data.user) {
        localStorage.setItem('nitrous_user', JSON.stringify(data.user))
      }

      // Preserve intended role selection for post-payment activation.
      if (mode === 'signup') {
        localStorage.setItem('nitrous_signup_selected_role', formData.role)
      }

      // One-way sync only: guest local cart -> authenticated server cart on login/signup.
      try {
        const rawGuestCart = localStorage.getItem(CART_STORAGE_KEY)
        if (rawGuestCart) {
          const parsed = JSON.parse(rawGuestCart)
          if (Array.isArray(parsed) && parsed.length > 0) {
            const items: CartItem[] = parsed
              .filter((entry) => entry?.item?.id && typeof entry?.quantity === 'number')
              .map((entry) => ({
                merchId: entry.item.id,
                name: entry.item.name,
                icon: entry.item.icon,
                price: entry.item.price,
                category: entry.item.category,
                quantity: Math.max(1, Math.floor(entry.quantity)),
                size: entry.size,
              }))

            if (items.length > 0) {
              await saveCart(items, data.token)
            }
          }
          // Prevent authenticated cart data from leaking back into guest sessions.
          localStorage.removeItem(CART_STORAGE_KEY)
        }
      } catch {
        // Ignore cart sync errors; auth success should still continue.
      }

      // Next.js standard for navigation
      if (mode === 'signup') {
        if (formData.role === 'viewer') {
          router.push('/')
        } else if (formData.role === 'participant' || formData.role === 'manager') {
          router.push(`/settings?upgrade=VIP&targetRole=${encodeURIComponent(formData.role)}`)
        } else if (formData.role === 'sponsor') {
          router.push('/settings?upgrade=PLATINUM&targetRole=sponsor')
        } else {
          router.push('/')
        }
      } else {
        router.push('/')
      }
      router.refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  // Password strength logic (simple version)
  const strength = formData.password.length === 0 ? 0 : 
                   formData.password.length < 5 ? 1 : 
                   formData.password.length < 9 ? 2 : 3

  return (
    <>
      <Nav />
      <main className={styles.page}>
        <div className={styles.bgGrid} />
        <div className={styles.bgGlow1} />
        <div className={styles.bgGlow2} />

        <div className={styles.container}>
          {/* Left panel — Branding */}
          <div className={styles.leftPanel}>
            <div className={styles.leftContent}>
              <div className={styles.leftTag}>
                <div className={styles.leftTagDot} />
                <span>SECURE ACCESS PORTAL</span>
              </div>
              <h2 className={styles.leftTitle}>
                FUEL<br />
                <span className={styles.leftTitleAccent}>YOUR</span><br />
                SPEED
              </h2>
              <p className={styles.leftSub}>
                One login. Every race on the planet. VIP passes and exclusive journeys.
              </p>
            </div>
          </div>

          {/* Right panel — Form */}
          <div className={styles.rightPanel}>
            <div className={styles.formBox}>
              <div className={styles.modeToggle}>
                <button
                  type="button"
                  className={`${styles.modeBtn} ${mode === 'signin' ? styles.modeBtnActive : ''}`}
                  onClick={() => toggleMode('signin')}
                >
                  SIGN IN
                </button>
                <button
                  type="button"
                  className={`${styles.modeBtn} ${mode === 'signup' ? styles.modeBtnActive : ''}`}
                  onClick={() => toggleMode('signup')}
                >
                  CREATE ACCOUNT
                </button>
              </div>

              <div className={styles.formHeader}>
                <div className={styles.formTitle}>
                  {mode === 'signin' ? 'Welcome back' : 'Join Nitrous'}
                </div>
              </div>

              <form onSubmit={handleSubmit} className={styles.form}>
                {mode === 'signup' && (
                  <div className={styles.fieldGroup}>
                    <label className={styles.fieldLabel}>FULL NAME</label>
                    <div className={styles.fieldWrap}>
                      <span className={styles.fieldIcon}>👤</span>
                      <input
                        name="name"
                        className={styles.fieldInput}
                        type="text"
                        placeholder="Your full name"
                        value={formData.name}
                        onChange={handleInput}
                        required
                      />
                    </div>
                  </div>
                )}

                {mode === 'signup' && (
                  <div className={styles.fieldGroup}>
                    <label className={styles.fieldLabel}>SELECT YOUR ROLE</label>
                    <div className={styles.roleGrid}>
                      {[
                        { value: 'viewer', label: 'Viewer', icon: '👁', desc: 'Watch live streams and events' },
                        { value: 'participant', label: 'Participant', icon: '🏎', desc: 'Join events and track progress' },
                        { value: 'manager', label: 'Manager', icon: '📊', desc: 'Manage teams and events' },
                        { value: 'sponsor', label: 'Sponsor', icon: '💎', desc: 'Fund and support events' }
                      ].map((option) => (
                        <button
                          key={option.value}
                          type="button"
                          className={`${styles.roleOption} ${formData.role === option.value ? styles.roleOptionActive : ''}`}
                          onClick={() => setFormData(prev => ({ ...prev, role: option.value as typeof prev.role }))}
                        >
                          <span className={styles.roleIcon}>{option.icon}</span>
                          <span className={styles.roleLabel}>{option.label}</span>
                          <span className={styles.roleDesc}>{option.desc}</span>
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <div className={styles.fieldGroup}>
                  <label className={styles.fieldLabel}>EMAIL ADDRESS</label>
                  <div className={styles.fieldWrap}>
                    <span className={styles.fieldIcon}>✉</span>
                    <input
                      name="email"
                      className={styles.fieldInput}
                      type="email"
                      placeholder="your@email.com"
                      value={formData.email}
                      onChange={handleInput}
                      required
                      autoComplete="email"
                    />
                  </div>
                </div>

                <div className={styles.fieldGroup}>
                  <div className={styles.fieldLabelRow}>
                    <label className={styles.fieldLabel}>PASSWORD</label>
                  </div>
                  <div className={styles.fieldWrap}>
                    <span className={styles.fieldIcon}>🔒</span>
                    <input
                      name="password"
                      className={styles.fieldInput}
                      type={showPass ? 'text' : 'password'}
                      placeholder="Your password"
                      value={formData.password}
                      onChange={handleInput}
                      required
                      autoComplete={mode === 'signin' ? 'current-password' : 'new-password'}
                    />
                    <button
                      type="button"
                      className={styles.showPassBtn}
                      onClick={() => setShowPass(!showPass)}
                      tabIndex={-1}
                    >
                      {showPass ? '🙈' : '👁'}
                    </button>
                  </div>
                  
                  {mode === 'signup' && (
                    <div className={styles.passwordStrength}>
                      {[0, 1, 2, 3].map((i) => (
                        <div
                          key={i}
                          className={styles.strengthBar}
                          style={{
                            background: strength > i 
                              ? (strength < 2 ? 'var(--red)' : strength < 3 ? '#facc15' : '#4ade80') 
                              : 'rgba(255,255,255,0.08)'
                          }}
                        />
                      ))}
                    </div>
                  )}
                </div>

                {error && <div className={styles.errorMsg}><span>⚠</span> {error}</div>}

                <button type="submit" className={styles.submitBtn} disabled={loading}>
                  {loading ? 'AUTHENTICATING...' : (mode === 'signin' ? '▶ IGNITE ACCESS' : '▶ CREATE ACCOUNT')}
                </button>
              </form>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}