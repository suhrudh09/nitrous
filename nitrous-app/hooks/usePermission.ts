import { useEffect, useState } from 'react'
import { getCurrentUser } from '@/lib/api'

export type UserRole = 'viewer' | 'participant' | 'manager' | 'sponsor' | 'admin'

interface UserInfo {
  id: string
  email: string
  name: string
  role: UserRole
  createdAt: string
}

// Cache for user info
let userCache: UserInfo | null = null
let cacheToken: string | null = null

export function useUser() {
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchUser() {
      const token = localStorage.getItem('nitrous_token')
      const storedUser = localStorage.getItem('nitrous_user')
      
      // Use cached user if token hasn't changed
      if (token === cacheToken && userCache) {
        setUser(userCache)
        setLoading(false)
        return
      }

      if (!token) {
        setLoading(false)
        return
      }

      try {
        const userData = await getCurrentUser(token)
        const userInfo: UserInfo = {
          id: userData.id,
          email: userData.email,
          name: userData.name,
          role: (userData.role as UserRole) || 'viewer',
          createdAt: userData.createdAt,
        }
        
        // Update cache
        userCache = userInfo
        cacheToken = token
        
        // Also store in localStorage for persistence
        localStorage.setItem('nitrous_user', JSON.stringify(userInfo))
        
        setUser(userInfo)
      } catch (err) {
        // Token might be invalid
        localStorage.removeItem('nitrous_token')
        localStorage.removeItem('nitrous_user')
        userCache = null
        cacheToken = null
        setError('Session expired')
      } finally {
        setLoading(false)
      }
    }

    fetchUser()
  }, [])

  const clearUser = () => {
    userCache = null
    cacheToken = null
    localStorage.removeItem('nitrous_token')
    localStorage.removeItem('nitrous_user')
    setUser(null)
  }

  return { user, loading, error, clearUser }
}

// Check if user has required role
export function hasRole(userRole: UserRole | undefined, allowedRoles: UserRole[]): boolean {
  if (!userRole) return false
  return allowedRoles.includes(userRole)
}

// Check if user can manage a specific team
export function useCanManageTeam(teamId: string) {
  const { user, loading } = useUser()
  
  if (loading || !user) return false
  
  // Admin can manage any team
  if (user.role === 'admin') return true
  
  // Manager can manage their teams (would need team ownership check)
  if (user.role === 'manager') return true
  
  return false
}

// Check if user is admin
export function useIsAdmin() {
  const { user, loading } = useUser()
  if (loading || !user) return false
  return user.role === 'admin'
}

// Check if user can tune (manager or admin)
export function useCanTune() {
  const { user, loading } = useUser()
  if (loading || !user) return false
  return user.role === 'manager' || user.role === 'admin'
}

// Check if user can book journeys for themselves
export function useCanBookJourneys() {
  const { user, loading } = useUser()
  if (loading || !user) return false
  // Participants cannot self-register
  return user.role !== 'participant'
}

// Check if user can register others for journeys
export function useCanRegisterOthers() {
  const { user, loading } = useUser()
  if (loading || !user) return false
  return user.role === 'manager' || user.role === 'admin'
}