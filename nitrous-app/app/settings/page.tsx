'use client'
import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'
import Link from 'next/link'
import Nav from '@/components/Nav'
import RoleBadge from '@/components/RoleBadge'
import { createPaymentIntent, confirmPayment, getCurrentUser, updateCurrentUserPlan, updateCurrentUserRole } from '@/lib/api'
import styles from './settings.module.css'

type Tab = 'profile' | 'notifications' | 'security' | 'preferences'

interface UserProfile {
  name: string
  email: string
  initials: string
  joinedDate: string
  plan: 'FREE' | 'VIP' | 'PLATINUM'
  role: 'viewer' | 'participant' | 'manager' | 'sponsor' | 'admin'
}

type MembershipPlan = 'FREE' | 'VIP' | 'PLATINUM'
type UserRole = 'viewer' | 'participant' | 'manager' | 'sponsor'

interface PlanCheckout {
  plan: Exclude<MembershipPlan, 'FREE'>
  label: string
  priceLabel: string
  amountCents: number
}

function formatCardNumber(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 16)
  return digits.replace(/(\d{4})(?=\d)/g, '$1 ')
}

function formatExpiry(value: string) {
  const digits = value.replace(/\D/g, '').slice(0, 4)
  if (digits.length >= 2) return digits.slice(0, 2) + '/' + digits.slice(2)
  return digits
}

const tabs: { id: Tab; label: string; icon: string }[] = [
  { id: 'profile', label: 'Profile', icon: '👤' },
  { id: 'notifications', label: 'Notifications', icon: '🔔' },
  { id: 'security', label: 'Security', icon: '🔒' },
  { id: 'preferences', label: 'Preferences', icon: '⚙️' },
]

export default function SettingsPage() {
  const searchParams = useSearchParams()
  const [tab, setTab] = useState<Tab>('profile')
  const [user, setUser] = useState<UserProfile | null>(null)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [saved, setSaved] = useState(false)
  const [checkout, setCheckout] = useState<PlanCheckout | null>(null)
  const [payStatus, setPayStatus] = useState<'idle' | 'processing' | 'done' | 'error'>('idle')
  const [payError, setPayError] = useState('')
  const [cardNumber, setCardNumber] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvc, setCvc] = useState('')
  const [cardName, setCardName] = useState('')
  const [pendingRoleUpgrade, setPendingRoleUpgrade] = useState<UserRole | null>(null)
  const [autoCheckoutHandled, setAutoCheckoutHandled] = useState(false)

  // Notification toggles
  const [notifs, setNotifs] = useState({
    liveAlerts: true,
    eventReminders: true,
    journeyUpdates: true,
    orderUpdates: true,
    teamNews: false,
    promoEmails: false,
  })

  // Preferences
  const [prefs, setPrefs] = useState({
    autoplay: true,
    hd: true,
    timezone: 'auto',
    theme: 'dark',
    units: 'metric',
  })

  useEffect(() => {
    try {
      const token = localStorage.getItem('nitrous_token')
      const stored = localStorage.getItem('nitrous_user')
      if (stored && token) {
        const parsed = JSON.parse(stored)
        const names = (parsed.name || 'User').trim().split(' ')
        const initials = names.length >= 2
          ? `${names[0][0]}${names[names.length - 1][0]}`.toUpperCase()
          : names[0].slice(0, 2).toUpperCase()
        const profile: UserProfile = {
          name: parsed.name || 'User',
          email: parsed.email || '',
          initials,
          joinedDate: parsed.joinedDate || 'Jan 2025',
          plan: (parsed.plan || 'FREE').toUpperCase() === 'PRO' ? 'VIP' : (parsed.plan || 'FREE'),
          role: parsed.role || 'viewer',
        }
        setUser(profile)
        setName(profile.name)
        setEmail(profile.email)

        getCurrentUser(token)
          .then((remote) => {
            const remoteNames = (remote.name || 'User').trim().split(' ')
            const remoteInitials = remoteNames.length >= 2
              ? `${remoteNames[0][0]}${remoteNames[remoteNames.length - 1][0]}`.toUpperCase()
              : remoteNames[0].slice(0, 2).toUpperCase()
            const remoteProfile: UserProfile = {
              name: remote.name,
              email: remote.email,
              initials: remoteInitials,
              joinedDate: profile.joinedDate,
              plan: remote.plan,
              role: remote.role,
            }
            setUser(remoteProfile)
            setName(remoteProfile.name)
            setEmail(remoteProfile.email)
            localStorage.setItem('nitrous_user', JSON.stringify({ ...parsed, ...remote, plan: remote.plan }))
          })
          .catch(() => {
            // keep local snapshot if /auth/me fails
          })
      } else {
        globalThis.location.href = '/login'
      }
    } catch {
      globalThis.location.href = '/login'
    }
  }, [])

  function handleSaveProfile(e: React.FormEvent) {
    e.preventDefault()
    const names = name.trim().split(' ')
    const initials = names.length >= 2
      ? `${names[0][0]}${names[names.length - 1][0]}`.toUpperCase()
      : names[0].slice(0, 2).toUpperCase()
    const updated = { ...user, name, email, initials } as UserProfile
    setUser(updated)
    try {
      const stored = JSON.parse(localStorage.getItem('nitrous_user') || '{}')
      localStorage.setItem('nitrous_user', JSON.stringify({ ...stored, name, email }))
    } catch {}
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  function handleSignOut() {
    localStorage.removeItem('nitrous_token')
    localStorage.removeItem('nitrous_user')
    globalThis.location.href = '/'
  }

  const planColors: Record<string, string> = {
    FREE: 'var(--muted)',
    VIP: 'var(--cyan)',
    PLATINUM: '#facc15',
  }

  const planRank: Record<MembershipPlan, number> = {
    FREE: 0,
    VIP: 1,
    PLATINUM: 2,
  }

  const plans: Array<{
    id: MembershipPlan
    label: string
    price: string
    amountCents: number
    perks: string[]
  }> = [
    { id: 'FREE', label: 'Free', price: '$0', amountCents: 0, perks: ['Standard streams', 'Public events', 'Basic merch access'] },
    { id: 'VIP', label: 'VIP', price: '$12/mo', amountCents: 1200, perks: ['HD + 4K streams', 'Priority event access', '10% merch discount', 'Journey early access'] },
    { id: 'PLATINUM', label: 'Platinum', price: '$29/mo', amountCents: 2900, perks: ['All VIP features', '15% merch discount', 'Exclusive journeys', 'Pit lane access passes', 'Dedicated support'] },
  ]

  const visiblePlans = user
    ? plans.filter(plan => planRank[plan.id] >= planRank[user.plan])
    : plans

  const openPlanCheckout = (plan: { id: MembershipPlan; label: string; price: string; amountCents: number }) => {
    if (plan.id === 'FREE' || !user) return
    setCheckout({
      plan: plan.id,
      label: plan.label,
      priceLabel: plan.price,
      amountCents: plan.amountCents,
    })
    setPayStatus('idle')
    setPayError('')
    setCardNumber('')
    setExpiry('')
    setCvc('')
    setCardName('')
  }

  useEffect(() => {
    if (!user || autoCheckoutHandled) return

    const rawTargetRole = (searchParams.get('targetRole') || localStorage.getItem('nitrous_signup_selected_role') || '').toLowerCase()
    const requestedPlan = (searchParams.get('upgrade') || '').toUpperCase()
    const targetRole: UserRole | null =
      rawTargetRole === 'participant' || rawTargetRole === 'manager' || rawTargetRole === 'sponsor'
        ? (rawTargetRole as UserRole)
        : null

    if (!targetRole || (requestedPlan !== 'VIP' && requestedPlan !== 'PLATINUM')) {
      setAutoCheckoutHandled(true)
      return
    }

    const chosen = plans.find(p => p.id === requestedPlan)
    if (!chosen) {
      setAutoCheckoutHandled(true)
      return
    }

    // If user already has enough plan, only perform role update.
    if (planRank[user.plan] >= planRank[chosen.id]) {
      const token = localStorage.getItem('nitrous_token')
      if (!token) {
        setAutoCheckoutHandled(true)
        return
      }
      updateCurrentUserRole(targetRole, token)
        .then((updatedRoleUser) => {
          setUser(prev => prev ? ({ ...prev, role: updatedRoleUser.role }) : prev)
          try {
            const stored = JSON.parse(localStorage.getItem('nitrous_user') || '{}')
            localStorage.setItem('nitrous_user', JSON.stringify({ ...stored, role: updatedRoleUser.role }))
            localStorage.removeItem('nitrous_signup_selected_role')
          } catch {
            // ignore
          }
        })
        .finally(() => setAutoCheckoutHandled(true))
      return
    }

    setPendingRoleUpgrade(targetRole)
    openPlanCheckout(chosen)
    setAutoCheckoutHandled(true)
  }, [user, autoCheckoutHandled, searchParams])

  const handlePlanPayment = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!checkout || !user) return

    const token = localStorage.getItem('nitrous_token')
    if (!token) {
      globalThis.location.href = '/login'
      return
    }

    setPayStatus('processing')
    setPayError('')

    try {
      const intent = await createPaymentIntent(checkout.amountCents, 'membership', checkout.plan, token)
      await confirmPayment(intent.paymentId, token)
      const updatedPlanUser = await updateCurrentUserPlan(checkout.plan, token)
      let finalUser = updatedPlanUser

      if (pendingRoleUpgrade) {
        finalUser = await updateCurrentUserRole(pendingRoleUpgrade, token)
      }

      const names = (finalUser.name || 'User').trim().split(' ')
      const initials = names.length >= 2
        ? `${names[0][0]}${names[names.length - 1][0]}`.toUpperCase()
        : names[0].slice(0, 2).toUpperCase()

      setUser(prev => prev ? ({
        ...prev,
        name: finalUser.name,
        email: finalUser.email,
        initials,
        role: finalUser.role,
        plan: finalUser.plan,
      }) : prev)

      try {
        const stored = JSON.parse(localStorage.getItem('nitrous_user') || '{}')
        localStorage.setItem('nitrous_user', JSON.stringify({
          ...stored,
          name: finalUser.name,
          email: finalUser.email,
          role: finalUser.role,
          plan: finalUser.plan,
        }))
        localStorage.removeItem('nitrous_signup_selected_role')
      } catch {
        // ignore local storage sync errors
      }

      setPendingRoleUpgrade(null)
      setPayStatus('done')
      setTimeout(() => setCheckout(null), 1600)
    } catch (err: unknown) {
      setPayStatus('error')
      setPayError(err instanceof Error ? err.message : 'Payment failed')
    }
  }

  if (!user) return null

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Header */}
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ SETTINGS</div>
            <h1 className={styles.pageTitle}>ACCOUNT SETTINGS</h1>
          </div>
        </div>

        <div className={styles.layout}>
          {/* Sidebar */}
          <aside className={styles.sidebar}>
            {/* Profile card */}
            <div className={styles.profileCard}>
              <div className={styles.avatarRingWrap}>
                <div className={styles.avatar}>
                  <span className={styles.avatarInitials}>{user.initials}</span>
                </div>
                <div className={styles.avatarGlow} />
              </div>
              <div className={styles.profileName}>{user.name}</div>
              <div className={styles.profileEmail}>{user.email}</div>
              <RoleBadge role={user.role} size="sm" />
              <div
                className={styles.planBadge}
                style={{ color: planColors[user.plan], borderColor: `${planColors[user.plan]}66` }}
              >
                {user.plan} MEMBER
              </div>
            </div>

            {/* Nav */}
            <nav className={styles.sideNav}>
              {tabs.map(t => (
                <button
                  key={t.id}
                  className={`${styles.sideNavItem} ${tab === t.id ? styles.sideNavItemActive : ''}`}
                  onClick={() => setTab(t.id)}
                >
                  <span className={styles.sideNavIcon}>{t.icon}</span>
                  <span>{t.label}</span>
                  <span className={styles.sideNavArrow}>→</span>
                </button>
              ))}
            </nav>

            {/* Quick links */}
            <div className={styles.sideQuick}>
              <Link href="/garage" className={styles.quickLink}>🚗 My Garage</Link>
              <Link href="/passes" className={styles.quickLink}>🎫 My Passes</Link>
              <Link href="/reminders" className={styles.quickLink}>⏰ My Reminders</Link>
              <Link href="/journeys" className={styles.quickLink}>🌍 My Journeys</Link>
            </div>

            <button className={styles.signOutBtn} onClick={handleSignOut}>
              <span>⎋</span> Sign Out
            </button>
          </aside>

          {/* Content */}
          <div className={styles.content}>

            {/* Profile tab */}
            {tab === 'profile' && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <div className={styles.sectionTitle}>Profile Information</div>
                  <div className={styles.sectionSub}>Update your personal details</div>
                </div>

                <form onSubmit={handleSaveProfile} className={styles.form}>
                  <div className={styles.formRow}>
                    <div className={styles.fieldGroup}>
                      <label className={styles.fieldLabel}>FULL NAME</label>
                      <input
                        className={styles.fieldInput}
                        type="text"
                        value={name}
                        onChange={e => setName(e.target.value)}
                      />
                    </div>
                    <div className={styles.fieldGroup}>
                      <label className={styles.fieldLabel}>EMAIL ADDRESS</label>
                      <input
                        className={styles.fieldInput}
                        type="email"
                        value={email}
                        onChange={e => setEmail(e.target.value)}
                      />
                    </div>
                  </div>

                  <div className={styles.fieldGroup}>
                    <label className={styles.fieldLabel}>MEMBER SINCE</label>
                    <input className={styles.fieldInput} type="text" value={user.joinedDate} disabled />
                  </div>

                  <div className={styles.formActions}>
                    <button type="submit" className={`${styles.saveBtn} ${saved ? styles.saveBtnDone : ''}`}>
                      {saved ? '✓ SAVED' : 'SAVE CHANGES'}
                    </button>
                  </div>
                </form>

                {/* Plan section */}
                <div className={styles.planSection}>
                  <div className={styles.sectionTitle} style={{ fontSize: '15px', marginBottom: '16px' }}>Membership Plan</div>
                  <div className={styles.planCards}>
                    {visiblePlans.map(plan => (
                      <div
                        key={plan.id}
                        className={`${styles.planCard} ${user.plan === plan.id ? styles.planCardActive : ''}`}
                        style={user.plan === plan.id ? { borderColor: planColors[plan.id] } : {}}
                      >
                        {user.plan === plan.id && (
                          <div className={styles.planCurrent} style={{ color: planColors[plan.id] }}>CURRENT</div>
                        )}
                        <div className={styles.planName} style={{ color: user.plan === plan.id ? planColors[plan.id] : 'var(--text-bright)' }}>
                          {plan.label}
                        </div>
                        <div className={styles.planPrice}>{plan.price}</div>
                        <div className={styles.planPerks}>
                          {plan.perks.map((p, i) => (
                            <div key={i} className={styles.planPerk}>
                              <span style={{ color: planColors[plan.id] }}>✓</span> {p}
                            </div>
                          ))}
                        </div>
                        {user.plan !== plan.id && (
                          <button
                            className={styles.upgradePlanBtn}
                            style={{ borderColor: planColors[plan.id], color: planColors[plan.id] }}
                            onClick={() => openPlanCheckout(plan)}
                          >
                            Upgrade →
                          </button>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Notifications tab */}
            {tab === 'notifications' && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <div className={styles.sectionTitle}>Notifications</div>
                  <div className={styles.sectionSub}>Choose what you want to be notified about</div>
                </div>
                <div className={styles.toggleList}>
                  {(Object.entries(notifs) as [keyof typeof notifs, boolean][]).map(([key, val]) => {
                    const labels: Record<string, { label: string; desc: string }> = {
                      liveAlerts: { label: 'Live Alerts', desc: 'Get notified when events go live' },
                      eventReminders: { label: 'Event Reminders', desc: 'Reminders 1 hour before events' },
                      journeyUpdates: { label: 'Journey Updates', desc: 'Updates on your booked journeys' },
                      orderUpdates: { label: 'Order Updates', desc: 'Shipping and delivery notifications' },
                      teamNews: { label: 'Team News', desc: 'Latest from teams you follow' },
                      promoEmails: { label: 'Promotional Emails', desc: 'Special offers and member deals' },
                    }
                    return (
                      <div key={key} className={styles.toggleRow}>
                        <div className={styles.toggleInfo}>
                          <div className={styles.toggleLabel}>{labels[key].label}</div>
                          <div className={styles.toggleDesc}>{labels[key].desc}</div>
                        </div>
                        <button
                          className={`${styles.toggle} ${val ? styles.toggleOn : ''}`}
                          onClick={() => setNotifs(p => ({ ...p, [key]: !val }))}
                          aria-label={`Toggle ${key}`}
                        >
                          <div className={styles.toggleThumb} />
                        </button>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Security tab */}
            {tab === 'security' && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <div className={styles.sectionTitle}>Security</div>
                  <div className={styles.sectionSub}>Manage your password and account access</div>
                </div>

                <div className={styles.securityCard}>
                  <div className={styles.secCardTitle}>Change Password</div>
                  <div className={styles.form}>
                    <div className={styles.fieldGroup}>
                      <label className={styles.fieldLabel}>CURRENT PASSWORD</label>
                      <input className={styles.fieldInput} type="password" placeholder="••••••••" />
                    </div>
                    <div className={styles.formRow}>
                      <div className={styles.fieldGroup}>
                        <label className={styles.fieldLabel}>NEW PASSWORD</label>
                        <input className={styles.fieldInput} type="password" placeholder="••••••••" />
                      </div>
                      <div className={styles.fieldGroup}>
                        <label className={styles.fieldLabel}>CONFIRM NEW PASSWORD</label>
                        <input className={styles.fieldInput} type="password" placeholder="••••••••" />
                      </div>
                    </div>
                    <button className={styles.saveBtn} type="button">UPDATE PASSWORD</button>
                  </div>
                </div>

                <div className={styles.securityCard}>
                  <div className={styles.secCardTitle}>Active Sessions</div>
                  {[
                    { device: 'Chrome on macOS', location: 'New York, US', time: 'Current session', current: true },
                    { device: 'Safari on iPhone', location: 'New York, US', time: '2 hours ago', current: false },
                  ].map((session, i) => (
                    <div key={i} className={styles.sessionRow}>
                      <div className={styles.sessionInfo}>
                        <div className={styles.sessionDevice}>{session.device}</div>
                        <div className={styles.sessionMeta}>{session.location} · {session.time}</div>
                      </div>
                      {session.current ? (
                        <div className={styles.sessionCurrent}>CURRENT</div>
                      ) : (
                        <button className={styles.revokeBtn}>Revoke</button>
                      )}
                    </div>
                  ))}
                </div>

                <div className={`${styles.securityCard} ${styles.dangerCard}`}>
                  <div className={styles.secCardTitle} style={{ color: 'var(--red)' }}>Danger Zone</div>
                  <p className={styles.dangerDesc}>Permanently delete your account and all associated data. This action cannot be undone.</p>
                  <button className={styles.deleteAccountBtn}>Delete Account</button>
                </div>
              </div>
            )}

            {/* Preferences tab */}
            {tab === 'preferences' && (
              <div className={styles.section}>
                <div className={styles.sectionHeader}>
                  <div className={styles.sectionTitle}>Preferences</div>
                  <div className={styles.sectionSub}>Customize your Nitrous experience</div>
                </div>

                <div className={styles.prefGrid}>
                  {/* Streaming */}
                  <div className={styles.prefCard}>
                    <div className={styles.prefCardTitle}>STREAMING</div>
                    <div className={styles.toggleRow}>
                      <div className={styles.toggleInfo}>
                        <div className={styles.toggleLabel}>Auto-play Streams</div>
                        <div className={styles.toggleDesc}>Start streams automatically when opening a channel</div>
                      </div>
                      <button
                        className={`${styles.toggle} ${prefs.autoplay ? styles.toggleOn : ''}`}
                        onClick={() => setPrefs(p => ({ ...p, autoplay: !p.autoplay }))}
                      >
                        <div className={styles.toggleThumb} />
                      </button>
                    </div>
                    <div className={styles.toggleRow}>
                      <div className={styles.toggleInfo}>
                        <div className={styles.toggleLabel}>Default to HD</div>
                        <div className={styles.toggleDesc}>Use highest available quality by default</div>
                      </div>
                      <button
                        className={`${styles.toggle} ${prefs.hd ? styles.toggleOn : ''}`}
                        onClick={() => setPrefs(p => ({ ...p, hd: !p.hd }))}
                      >
                        <div className={styles.toggleThumb} />
                      </button>
                    </div>
                  </div>

                  {/* Regional */}
                  <div className={styles.prefCard}>
                    <div className={styles.prefCardTitle}>REGIONAL</div>
                    <div className={styles.fieldGroup}>
                      <label className={styles.fieldLabel}>UNITS</label>
                      <div className={styles.segmentControl}>
                        {['metric', 'imperial'].map(u => (
                          <button
                            key={u}
                            className={`${styles.segmentBtn} ${prefs.units === u ? styles.segmentBtnActive : ''}`}
                            onClick={() => setPrefs(p => ({ ...p, units: u }))}
                          >
                            {u.toUpperCase()}
                          </button>
                        ))}
                      </div>
                    </div>
                    <div className={styles.fieldGroup}>
                      <label className={styles.fieldLabel}>TIMEZONE</label>
                      <select
                        className={styles.fieldSelect}
                        value={prefs.timezone}
                        onChange={e => setPrefs(p => ({ ...p, timezone: e.target.value }))}
                      >
                        <option value="auto">Auto-detect</option>
                        <option value="utc">UTC</option>
                        <option value="et">Eastern Time</option>
                        <option value="pt">Pacific Time</option>
                        <option value="cet">Central European Time</option>
                        <option value="jst">Japan Standard Time</option>
                      </select>
                    </div>
                  </div>
                </div>

                <div className={styles.formActions} style={{ marginTop: '24px' }}>
                  <button className={styles.saveBtn}>SAVE PREFERENCES</button>
                </div>
              </div>
            )}

          </div>
        </div>

        {checkout && (
          <div className={styles.payModalOverlay} onClick={() => payStatus !== 'processing' && setCheckout(null)}>
            <div className={styles.payModal} onClick={e => e.stopPropagation()}>
              {payStatus === 'done' ? (
                <div className={styles.payDone}>
                  <div className={styles.payDoneIcon}>✓</div>
                  <div className={styles.payDoneTitle}>Plan Updated</div>
                  <div className={styles.payDoneSub}>{checkout.label} membership is now active.</div>
                </div>
              ) : (
                <>
                  <div className={styles.payModalHeader}>
                    <div className={styles.payModalTitle}>UPGRADE MEMBERSHIP</div>
                    <button className={styles.payModalClose} onClick={() => setCheckout(null)} disabled={payStatus === 'processing'}>✕</button>
                  </div>

                  <div className={styles.payOrderSummary}>
                    <div className={styles.payOrderRow}>
                      <span>{checkout.label} Plan</span>
                      <span>{checkout.priceLabel}</span>
                    </div>
                    <div className={styles.payOrderTotal}>
                      <span>CHARGE TODAY</span>
                      <span>{checkout.priceLabel}</span>
                    </div>
                  </div>

                  <form className={styles.payForm} onSubmit={handlePlanPayment}>
                    <div className={styles.payFormField}>
                      <label className={styles.payLabel}>CARDHOLDER NAME</label>
                      <input className={styles.payInput} value={cardName} onChange={e => setCardName(e.target.value)} placeholder="Full Name" required />
                    </div>
                    <div className={styles.payFormField}>
                      <label className={styles.payLabel}>CARD NUMBER</label>
                      <input className={styles.payInput} value={cardNumber} onChange={e => setCardNumber(formatCardNumber(e.target.value))} maxLength={19} placeholder="1234 5678 9012 3456" required />
                    </div>
                    <div className={styles.payFormRow}>
                      <div className={styles.payFormField}>
                        <label className={styles.payLabel}>EXPIRY</label>
                        <input className={styles.payInput} value={expiry} onChange={e => setExpiry(formatExpiry(e.target.value))} maxLength={5} placeholder="MM/YY" required />
                      </div>
                      <div className={styles.payFormField}>
                        <label className={styles.payLabel}>CVC</label>
                        <input className={styles.payInput} value={cvc} onChange={e => setCvc(e.target.value.replace(/\D/g, '').slice(0, 4))} maxLength={4} placeholder="123" required />
                      </div>
                    </div>
                    {payError && <div className={styles.payError}>{payError}</div>}
                    <button className={styles.paySubmitBtn} type="submit" disabled={payStatus === 'processing'}>
                      {payStatus === 'processing' ? 'PROCESSING...' : `PAY ${checkout.priceLabel}`}
                    </button>
                  </form>
                </>
              )}
            </div>
          </div>
        )}
      </main>
    </>
  )
}