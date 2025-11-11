"use client"

import { useState, useEffect } from 'react'
import { AppliedJob } from '../../lib/supabase'
import { Modal } from '@mantine/core'

type JobEditModalProps = {
  job: AppliedJob | null
  opened: boolean
  onClose: () => void
  onUpdate: (id: string, updates: Partial<AppliedJob>) => Promise<void>
  onDelete: (id: string) => Promise<void>
}

const STATUS_OPTIONS: AppliedJob['status'][] = [
  'applied',
  'interviewing',
  'rejected',
  'offer',
  'accepted'
]

export default function JobEditModal({ job, opened, onClose, onUpdate, onDelete }: JobEditModalProps) {
  const [formData, setFormData] = useState<Partial<AppliedJob>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    if (job) {
      setFormData({
        company: job.company,
        job_title: job.job_title,
        job_url: job.job_url || '',
        location: job.location || '',
        salary_range: job.salary_range || '',
        notes: job.notes || '',
        status: job.status,
      })
    }
  }, [job])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!job || !formData.company || !formData.job_title) {
      return
    }

    setIsSubmitting(true)
    try {
      await onUpdate(job.id, formData)
      onClose()
    } catch (error) {
      console.error('Failed to update job:', error)
      alert('Failed to update job. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleDelete = async () => {
    if (!job) return

    if (!isDeleting) {
      setIsDeleting(true)
      return
    }

    try {
      await onDelete(job.id)
      onClose()
    } catch (error) {
      console.error('Failed to delete job:', error)
      alert('Failed to delete job. Please try again.')
    } finally {
      setIsDeleting(false)
    }
  }

  if (!job) return null

  return (
    <Modal.Root opened={opened} onClose={onClose} size="lg">
      <Modal.Overlay backgroundOpacity={0.5} />
      <Modal.Content>
        <Modal.Header>
          <Modal.Title>Edit Job Application</Modal.Title>
          <Modal.CloseButton />
        </Modal.Header>
        <Modal.Body>
          <form onSubmit={handleSubmit}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Company <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={formData.company || ''}
                  onChange={(e) => setFormData({ ...formData, company: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Job Title <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={formData.job_title || ''}
                  onChange={(e) => setFormData({ ...formData, job_title: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Location
                </label>
                <input
                  type="text"
                  value={formData.location || ''}
                  onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Salary Range
                </label>
                <input
                  type="text"
                  value={formData.salary_range || ''}
                  onChange={(e) => setFormData({ ...formData, salary_range: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Job URL
                </label>
                <input
                  type="url"
                  value={formData.job_url || ''}
                  onChange={(e) => setFormData({ ...formData, job_url: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Status
                </label>
                <select
                  value={formData.status || 'applied'}
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
                  value={formData.notes || ''}
                  onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                  rows={3}
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6 justify-between">
              <button
                type="button"
                onClick={handleDelete}
                className={`px-6 py-3 rounded-lg font-semibold ${
                  isDeleting
                    ? 'bg-red-600 text-white hover:bg-red-700'
                    : 'bg-red-100 text-red-700 hover:bg-red-200'
                }`}
              >
                {isDeleting ? 'Click again to confirm delete' : 'Delete'}
              </button>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={onClose}
                  className="px-6 py-3 bg-zinc-200 text-zinc-700 rounded-lg hover:bg-zinc-300 font-semibold"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-6 py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 rounded-lg hover:brightness-95 font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isSubmitting ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          </form>
        </Modal.Body>
      </Modal.Content>
    </Modal.Root>
  )
}

