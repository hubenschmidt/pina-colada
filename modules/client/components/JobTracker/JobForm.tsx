"use client"

import { useState, useEffect } from 'react'
import { AppliedJob } from '../../lib/supabase'
import { getMostRecentResumeDate } from '../../api/jobs'
import { Plus, X } from 'lucide-react'

type JobFormProps = {
  onAdd: (job: Omit<AppliedJob, 'id' | 'created_at' | 'updated_at' | 'date'>) => Promise<void>
}

const STATUS_OPTIONS: AppliedJob['status'][] = [
  'applied',
  'interviewing',
  'rejected',
  'offer',
  'accepted'
]

const JobForm = ({ onAdd }: JobFormProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const [formData, setFormData] = useState({
    company: '',
    job_title: '',
    date: new Date().toISOString().split('T')[0], // Default to today
    job_url: '',
    salary_range: '',
    notes: '',
    resume: '',
    status: 'applied' as AppliedJob['status'],
    source: 'manual' as const,
    lead_status_id: null as string | null
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [useLatestResume, setUseLatestResume] = useState(true)

  // Fetch most recent resume date when form opens
  useEffect(() => {
    if (isOpen) {
      setUseLatestResume(true) // Reset checkbox when form opens
      getMostRecentResumeDate().then((resumeDate) => {
        if (resumeDate) {
          setFormData(prev => ({ ...prev, resume: resumeDate }))
        }
      }).catch((error) => {
        console.error('Failed to fetch recent resume date:', error)
        // Silently fail - user can still enter resume date manually
      })
    }
  }, [isOpen])

  const handleUseLatestResumeChange = (checked: boolean) => {
    setUseLatestResume(checked)
    if (!checked) {
      // Clear resume date when unchecked
      setFormData(prev => ({ ...prev, resume: '' }))
    } else {
      // Re-fetch and populate when checked
      getMostRecentResumeDate().then((resumeDate) => {
        if (resumeDate) {
          setFormData(prev => ({ ...prev, resume: resumeDate }))
        }
      }).catch((error) => {
        console.error('Failed to fetch recent resume date:', error)
      })
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.company || !formData.job_title) {
      alert('Company and Job Title are required')
      return
    }

    setIsSubmitting(true)
    try {
      await onAdd(formData)

      // Reset form
      setFormData({
        company: '',
        job_title: '',
        date: new Date().toISOString().split('T')[0],
        job_url: '',
        salary_range: '',
        notes: '',
        resume: '',
        status: 'applied',
        source: 'manual',
        lead_status_id: null
      })
      setUseLatestResume(true)
      setIsOpen(false)
    } catch (error: any) {
      console.error('Failed to add job:', error)
      const errorMessage = error?.message || error?.error || 'Failed to add job. Please try again.'
      alert(errorMessage)
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 rounded-lg hover:brightness-95 font-semibold shadow-md"
      >
        <Plus size={20} />
        Add New Application
      </button>
    )
  }

  return (
    <div className="border border-zinc-300 rounded-lg p-6 bg-blue-50 shadow-lg">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-xl font-semibold text-blue-900">Add New Job Application</h3>
        <button
          onClick={() => setIsOpen(false)}
          className="p-2 text-zinc-600 hover:bg-zinc-200 rounded"
        >
          <X size={20} />
        </button>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Company <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              required
              value={formData.company}
              onChange={(e) => setFormData({ ...formData, company: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
              placeholder="e.g., Google"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Job Title <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              required
              value={formData.job_title}
              onChange={(e) => setFormData({ ...formData, job_title: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
              placeholder="e.g., Software Engineer"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Date <span className="text-red-500">*</span>
            </label>
            <input
              type="date"
              required
              value={formData.date}
              onChange={(e) => setFormData({ ...formData, date: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Resume Date
            </label>
            <input
              type="date"
              value={formData.resume}
              onChange={(e) => setFormData({ ...formData, resume: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
            />
            <label className="flex items-center gap-2 mt-2 text-sm text-zinc-600">
              <input
                type="checkbox"
                checked={useLatestResume}
                onChange={(e) => handleUseLatestResumeChange(e.target.checked)}
                className="w-4 h-4 text-lime-500 border-zinc-300 rounded focus:ring-lime-500"
              />
              <span>Use latest resume on file?</span>
            </label>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Salary Range
            </label>
            <input
              type="text"
              value={formData.salary_range}
              onChange={(e) => setFormData({ ...formData, salary_range: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
              placeholder="e.g., $120k - $150k"
            />
          </div>

          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Job URL
            </label>
            <input
              type="url"
              value={formData.job_url}
              onChange={(e) => setFormData({ ...formData, job_url: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
              placeholder="https://..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Status
            </label>
            <select
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value as AppliedJob['status'] })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
            >
              {STATUS_OPTIONS.map((status) => (
                <option key={status} value={status}>
                  {status.charAt(0).toUpperCase() + status.slice(1)}
                </option>
              ))}
            </select>
          </div>

          <div className="md:col-span-2">
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Notes
            </label>
            <textarea
              value={formData.notes}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
              rows={3}
              placeholder="Additional notes about this application..."
            />
          </div>
        </div>

        <div className="flex gap-3 mt-6">
          <button
            type="submit"
            disabled={isSubmitting}
            className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 rounded-lg hover:brightness-95 font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus size={18} />
            {isSubmitting ? 'Adding...' : 'Add Application'}
          </button>
          <button
            type="button"
            onClick={() => setIsOpen(false)}
            className="px-6 py-3 bg-zinc-200 text-zinc-700 rounded-lg hover:bg-zinc-300 font-semibold"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  )
}

export default JobForm
