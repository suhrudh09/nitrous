'use client'

import styles from './RoleBadge.module.css'

type UserRole = 'viewer' | 'participant' | 'manager' | 'sponsor' | 'admin'

interface RoleBadgeProps {
  role: UserRole
  size?: 'sm' | 'md' | 'lg'
}

const roleConfig: Record<UserRole, { label: string; className: string }> = {
  viewer: { label: 'VIEWER', className: styles.viewer },
  participant: { label: 'PARTICIPANT', className: styles.participant },
  manager: { label: 'MANAGER', className: styles.manager },
  sponsor: { label: 'SPONSOR', className: styles.sponsor },
  admin: { label: 'ADMIN', className: styles.admin },
}

export default function RoleBadge({ role, size = 'md' }: RoleBadgeProps) {
  const config = roleConfig[role] || roleConfig.viewer
  
  return (
    <span className={`${styles.badge} ${config.className} ${styles[size]}`}>
      {config.label}
    </span>
  )
}