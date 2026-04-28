'use client'
import { useState, useEffect } from 'react'
import Nav from '@/components/Nav'
import RoleBadge from '@/components/RoleBadge'
import { useIsAdmin, useUser, UserRole } from '@/hooks/usePermission'
import { triggerSync, updateUserRole, getTeams } from '@/lib/api'
import type { Team, User } from '@/types'
import styles from './admin.module.css'

export default function AdminPage() {
  const isAdmin = useIsAdmin()
  const { user } = useUser()
  
  const [syncing, setSyncing] = useState(false)
  const [syncResults, setSyncResults] = useState<Record<string, { success: boolean; error?: string }> | null>(null)
  const [syncError, setSyncError] = useState<string | null>(null)
  
  const [users, setUsers] = useState<Array<{ id: string; name: string; email: string; role: UserRole }>>([])
  const [selectedUser, setSelectedUser] = useState<string>('')
  const [newRole, setNewRole] = useState<UserRole>('viewer')
  const [updatingRole, setUpdatingRole] = useState(false)
  const [roleUpdateSuccess, setRoleUpdateSuccess] = useState<string | null>(null)
  
  const [teams, setTeams] = useState<Team[]>([])

  useEffect(() => {
    // Load mock users for demo
    setUsers([
      { id: 'user-1', name: 'John Doe', email: 'john@example.com', role: 'viewer' },
      { id: 'user-2', name: 'Jane Smith', email: 'jane@example.com', role: 'participant' },
      { id: 'user-3', name: 'Bob Wilson', email: 'bob@example.com', role: 'manager' },
      { id: 'user-4', name: 'Alice Brown', email: 'alice@example.com', role: 'sponsor' },
    ])
    
    // Load teams
    getTeams()
      .then(setTeams)
      .catch(() => {})
  }, [])

  async function handleSync() {
    const token = localStorage.getItem('nitrous_token')
    if (!token || !isAdmin) return
    
    setSyncing(true)
    setSyncError(null)
    setSyncResults(null)
    
    try {
      const result = await triggerSync(token)
      setSyncResults(result.results)
    } catch (err) {
      setSyncError(err instanceof Error ? err.message : 'Sync failed')
    } finally {
      setSyncing(false)
    }
  }

  async function handleRoleUpdate() {
    const token = localStorage.getItem('nitrous_token')
    if (!token || !selectedUser || !isAdmin) return
    
    setUpdatingRole(true)
    setRoleUpdateSuccess(null)
    
    try {
      await updateUserRole(selectedUser, newRole, token)
      setUsers(prev => prev.map(u => u.id === selectedUser ? { ...u, role: newRole } : u))
      setRoleUpdateSuccess(`Role updated to ${newRole}`)
      setTimeout(() => setRoleUpdateSuccess(null), 3000)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Role update failed')
    } finally {
      setUpdatingRole(false)
    }
  }

  // Show access denied for non-admins
  if (!isAdmin && user) {
    return (
      <>
        <Nav />
        <main className={styles.page}>
          <div className={styles.accessDenied}>
            <span className={styles.accessIcon}>🔒</span>
            <h1>Access Denied</h1>
            <p>You need Admin role to access this page.</p>
            <p>Your current role: <RoleBadge role={user.role} /></p>
          </div>
        </main>
      </>
    )
  }

  return (
    <>
      <Nav />
      <main className={styles.page}>
        {/* Header */}
        <div className={styles.pageHeader}>
          <div>
            <div className={styles.headerTag}>/ ADMIN</div>
            <h1 className={styles.pageTitle}>ADMIN PANEL</h1>
            <p className={styles.pageSubtitle}>System management and user administration</p>
          </div>
        </div>

        <div className={styles.grid}>
          {/* Sync Panel */}
          <div className={styles.panel}>
            <div className={styles.panelHeader}>
              <span className={styles.panelIcon}>🔄</span>
              <h2 className={styles.panelTitle}>DATA SYNC</h2>
            </div>
            <div className={styles.panelContent}>
              <p className={styles.panelDesc}>
                Trigger data synchronization from external providers (Jolpica, SportsDB, OpenF1)
              </p>
              <button 
                className={styles.actionBtn}
                onClick={handleSync}
                disabled={syncing || !isAdmin}
              >
                {syncing ? 'SYNCING...' : 'TRIGGER SYNC'}
              </button>
              
              {syncError && (
                <div className={styles.errorMsg}>
                  <span>⚠</span> {syncError}
                </div>
              )}
              
              {syncResults && (
                <div className={styles.resultsList}>
                  {Object.entries(syncResults).map(([provider, result]) => (
                    <div key={provider} className={`${styles.resultItem} ${result.success ? styles.success : styles.error}`}>
                      <span className={styles.resultIcon}>{result.success ? '✓' : '✗'}</span>
                      <span className={styles.resultProvider}>{provider}</span>
                      <span className={styles.resultStatus}>
                        {result.success ? 'Success' : result.error || 'Failed'}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* User Role Management */}
          <div className={styles.panel}>
            <div className={styles.panelHeader}>
              <span className={styles.panelIcon}>👥</span>
              <h2 className={styles.panelTitle}>USER ROLE MANAGEMENT</h2>
            </div>
            <div className={styles.panelContent}>
              <p className={styles.panelDesc}>
                Change user roles to control access permissions
              </p>
              
              <div className={styles.formGroup}>
                <label className={styles.fieldLabel}>SELECT USER</label>
                <select 
                  className={styles.select}
                  value={selectedUser}
                  onChange={(e) => setSelectedUser(e.target.value)}
                  disabled={!isAdmin}
                >
                  <option value="">Choose a user...</option>
                  {users.map(u => (
                    <option key={u.id} value={u.id}>{u.name} ({u.email})</option>
                  ))}
                </select>
              </div>
              
              <div className={styles.formGroup}>
                <label className={styles.fieldLabel}>NEW ROLE</label>
                <select 
                  className={styles.select}
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as UserRole)}
                  disabled={!isAdmin}
                >
                  <option value="viewer">Viewer</option>
                  <option value="participant">Participant</option>
                  <option value="manager">Manager</option>
                  <option value="sponsor">Sponsor</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
              
              <button 
                className={styles.actionBtn}
                onClick={handleRoleUpdate}
                disabled={updatingRole || !selectedUser || !isAdmin}
              >
                {updatingRole ? 'UPDATING...' : 'UPDATE ROLE'}
              </button>
              
              {roleUpdateSuccess && (
                <div className={styles.successMsg}>
                  <span>✓</span> {roleUpdateSuccess}
                </div>
              )}
              
              {/* User list */}
              <div className={styles.userList}>
                <div className={styles.userListHeader}>CURRENT USERS</div>
                {users.map(u => (
                  <div key={u.id} className={styles.userItem}>
                    <span className={styles.userName}>{u.name}</span>
                    <span className={styles.userEmail}>{u.email}</span>
                    <RoleBadge role={u.role} size="sm" />
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Teams Overview */}
          <div className={styles.panel}>
            <div className={styles.panelHeader}>
              <span className={styles.panelIcon}>🏎️</span>
              <h2 className={styles.panelTitle}>TEAMS OVERVIEW</h2>
            </div>
            <div className={styles.panelContent}>
              <p className={styles.panelDesc}>
                View all teams in the system
              </p>
              
              <div className={styles.teamList}>
                {teams.length === 0 ? (
                  <div className={styles.emptyState}>No teams found</div>
                ) : (
                  teams.slice(0, 5).map(team => (
                    <div key={team.id} className={styles.teamItem}>
                      <span className={styles.teamName}>{team.name}</span>
                      <span className={styles.teamCountry}>{team.country || 'N/A'}</span>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}