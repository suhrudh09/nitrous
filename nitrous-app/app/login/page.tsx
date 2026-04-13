'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Nav from '@/components/Nav'
import styles from './login.module.css'

type Mode = 'signin' | 'signup'

export default function LoginPage() {
  const router = useRouter()
  const [mode, setMode] = useState<Mode>('signin')
  const [showPass, setShowPass] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  
  // Unified form state - keeps the Go backend DTO in mind
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    name: ''
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
    setFormData({ email: '', password: '', name: '' })
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
      const endpoint = `/api/auth/${mode === 'signin' ? 'login' : 'register'}`
      
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData),
      })

      const data = await res.json()

      if (!res.ok) {
        throw new Error(data.error || 'Authentication failed')
      }

      // Store auth state
      localStorage.setItem('nitrous_token', data.token)
      if (data.user) {
        localStorage.setItem('nitrous_user', JSON.stringify(data.user))
      }

      // Next.js standard for navigation
      router.push('/dashboard')
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