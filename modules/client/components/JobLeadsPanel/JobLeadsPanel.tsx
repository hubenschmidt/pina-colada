"use client"

import { useEffect, useState } from 'react'
import { X, ExternalLink, CheckCircle, Trash2, XCircle, Filter } from 'lucide-react'
import {
  fetchLeads,
  fetchLeadStatuses,
  updateJobLeadStatus,
  markLeadAsApplied,
  markLeadAsDoNotApply,
  deleteJob,
  type LeadStatus,
  type JobWithLeadStatus
} from '../../api/jobs'
import styles from './JobLeadsPanel.module.css'

interface JobLeadsPanelProps {
  isOpen: boolean
  onClose: () => void
  onLeadsChange?: () => void  // Callback to notify parent of changes
}

const JobLeadsPanel = ({ isOpen, onClose, onLeadsChange }: JobLeadsPanelProps) => {
  const [leads, setLeads] = useState<JobWithLeadStatus[]>([])
  const [leadStatuses, setLeadStatuses] = useState<LeadStatus[]>([])
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>(['Qualifying'])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showToast, setShowToast] = useState(false)
  const [toastMessage, setToastMessage] = useState('')

  // Load lead statuses on mount
  useEffect(() => {
    loadLeadStatuses()
  }, [])

  // Load leads when filters change or panel opens
  useEffect(() => {
    if (isOpen) {
      loadLeads()
    }
  }, [isOpen, selectedStatuses])

  const loadLeadStatuses = async () => {
    try {
      const statuses = await fetchLeadStatuses()
      setLeadStatuses(statuses)
    } catch (err) {
      console.error('Error loading lead statuses:', err)
    }
  }

  const loadLeads = async () => {
    setLoading(true)
    setError(null)
    try {
      const statusNames = selectedStatuses as ('Qualifying' | 'Cold' | 'Warm' | 'Hot')[]
      const data = await fetchLeads(statusNames.length > 0 ? statusNames : undefined)
      setLeads(data)
    } catch (err) {
      console.error('Error loading leads:', err)
      setError('Failed to load job leads')
    } finally {
      setLoading(false)
    }
  }

  const showSuccessToast = (message: string) => {
    setToastMessage(message)
    setShowToast(true)
    setTimeout(() => setShowToast(false), 3000)
  }

  const handleStatusFilterChange = (statusName: string) => {
    setSelectedStatuses(prev => {
      if (prev.includes(statusName)) {
        return prev.filter(s => s !== statusName)
      } else {
        return [...prev, statusName]
      }
    })
  }

  const handleSelectAllStatuses = () => {
    if (selectedStatuses.length === leadStatuses.length) {
      setSelectedStatuses([])
    } else {
      setSelectedStatuses(leadStatuses.map(s => s.name))
    }
  }

  const handleLeadStatusChange = async (jobId: string, newStatusId: string) => {
    try {
      await updateJobLeadStatus(jobId, newStatusId)
      await loadLeads()
      onLeadsChange?.()
      showSuccessToast('Lead status updated')
    } catch (err) {
      console.error('Error updating lead status:', err)
      alert('Failed to update lead status')
    }
  }

  const handleMarkAsApplied = async (jobId: string, company: string) => {
    try {
      await markLeadAsApplied(jobId)
      await loadLeads()
      onLeadsChange?.()
      showSuccessToast(`Marked ${company} as applied`)
    } catch (err) {
      console.error('Error marking as applied:', err)
      alert('Failed to mark as applied')
    }
  }

  const handleMarkAsDoNotApply = async (jobId: string, company: string) => {
    try {
      await markLeadAsDoNotApply(jobId)
      await loadLeads()
      onLeadsChange?.()
      showSuccessToast(`Marked ${company} as do not apply`)
    } catch (err) {
      console.error('Error marking as do not apply:', err)
      alert('Failed to mark as do not apply')
    }
  }

  const handleRemove = async (jobId: string, company: string) => {
    try {
      await deleteJob(jobId)
      await loadLeads()
      onLeadsChange?.()
      showSuccessToast(`Removed ${company} from leads`)
    } catch (err) {
      console.error('Error removing lead:', err)
      alert('Failed to remove lead')
    }
  }

  const getStatusBadgeClass = (statusName: string): string => {
    switch (statusName) {
      case 'Qualifying':
        return styles.badgeQualifying
      case 'Cold':
        return styles.badgeCold
      case 'Warm':
        return styles.badgeWarm
      case 'Hot':
        return styles.badgeHot
      default:
        return styles.badgeQualifying
    }
  }

  if (!isOpen) return null

  return (
    <>
      {/* Backdrop */}
      <div className={styles.backdrop} onClick={onClose} />

      {/* Panel */}
      <div className={styles.panel}>
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <h2 className={styles.title}>Job Leads</h2>
            <span className={styles.count}>({leads.length})</span>
          </div>
          <button
            onClick={onClose}
            className={styles.closeButton}
            aria-label="Close panel"
          >
            <X size={24} />
          </button>
        </div>

        {/* Filter */}
        <div className={styles.filterSection}>
          <div className={styles.filterHeader}>
            <Filter size={16} />
            <span>Filter by Status</span>
          </div>
          <div className={styles.filterOptions}>
            <label className={styles.filterOption}>
              <input
                type="checkbox"
                checked={selectedStatuses.length === leadStatuses.length}
                onChange={handleSelectAllStatuses}
              />
              <span>All</span>
            </label>
            {leadStatuses.map(status => (
              <label key={status.id} className={styles.filterOption}>
                <input
                  type="checkbox"
                  checked={selectedStatuses.includes(status.name)}
                  onChange={() => handleStatusFilterChange(status.name)}
                />
                <span className={getStatusBadgeClass(status.name)}>
                  {status.name}
                </span>
              </label>
            ))}
          </div>
        </div>

        {/* Content */}
        <div className={styles.content}>
          {loading && (
            <div className={styles.loading}>Loading leads...</div>
          )}

          {error && (
            <div className={styles.error}>{error}</div>
          )}

          {!loading && !error && leads.length === 0 && (
            <div className={styles.empty}>
              No job leads found. Try adjusting your filters.
            </div>
          )}

          {!loading && !error && leads.length > 0 && (
            <div className={styles.leadsList}>
              {leads.map(lead => (
                <div key={lead.id} className={styles.leadCard}>
                  <div className={styles.leadHeader}>
                    <div>
                      <h3 className={styles.leadCompany}>{lead.company}</h3>
                      <p className={styles.leadTitle}>{lead.job_title}</p>
                    </div>
                    {lead.lead_status && (
                      <select
                        value={lead.lead_status_id || ''}
                        onChange={(e) => handleLeadStatusChange(lead.id, e.target.value)}
                        className={`${styles.statusDropdown} ${getStatusBadgeClass(lead.lead_status.name)}`}
                      >
                        {leadStatuses.map(status => (
                          <option key={status.id} value={status.id}>
                            {status.name}
                          </option>
                        ))}
                      </select>
                    )}
                  </div>

                  {lead.job_url && (
                    <a
                      href={lead.job_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={styles.leadUrl}
                    >
                      <ExternalLink size={14} />
                      <span>View posting</span>
                    </a>
                  )}

                  <div className={styles.leadActions}>
                    <button
                      onClick={() => handleMarkAsApplied(lead.id, lead.company)}
                      className={`${styles.actionButton} ${styles.appliedButton}`}
                      title="Mark as applied"
                    >
                      <CheckCircle size={16} />
                      <span>Applied</span>
                    </button>
                    <button
                      onClick={() => handleRemove(lead.id, lead.company)}
                      className={`${styles.actionButton} ${styles.removeButton}`}
                      title="Remove from list"
                    >
                      <Trash2 size={16} />
                    </button>
                    <button
                      onClick={() => handleMarkAsDoNotApply(lead.id, lead.company)}
                      className={`${styles.actionButton} ${styles.doNotApplyButton}`}
                      title="Mark as do not apply"
                    >
                      <XCircle size={16} />
                      <span>Do Not Apply</span>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Toast notification */}
      {showToast && (
        <div className={styles.toast}>
          <CheckCircle size={16} />
          <span>{toastMessage}</span>
        </div>
      )}
    </>
  )
}

export default JobLeadsPanel
