"use client"

import { useState } from 'react'
import { AppliedJob } from '../../lib/supabase'
import { Edit2, Trash2, ExternalLink, Save, X } from 'lucide-react'

type JobRowProps = {
  job: AppliedJob
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

const STATUS_COLORS = {
  applied: 'bg-blue-100 text-blue-800',
  interviewing: 'bg-yellow-100 text-yellow-800',
  rejected: 'bg-red-100 text-red-800',
  offer: 'bg-green-100 text-green-800',
  accepted: 'bg-purple-100 text-purple-800'
}

const JobRow = ({ job, onUpdate, onDelete }: JobRowProps) => {
  const [isEditing, setIsEditing] = useState(false)
  const [editData, setEditData] = useState<Partial<AppliedJob>>({
    company: job.company,
    job_title: job.job_title,
    job_url: job.job_url || '',
    location: job.location || '',
    salary_range: job.salary_range || '',
    notes: job.notes || '',
    status: job.status
  })
  const [isDeleting, setIsDeleting] = useState(false)

  const handleSave = async () => {
    await onUpdate(job.id, editData)
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditData({
      company: job.company,
      job_title: job.job_title,
      job_url: job.job_url || '',
      location: job.location || '',
      salary_range: job.salary_range || '',
      notes: job.notes || '',
      status: job.status
    })
    setIsEditing(false)
  }

  const handleDelete = async () => {
    if (!isDeleting) {
      setIsDeleting(true)
      return
    }
    await onDelete(job.id)
  }

  if (isEditing) {
    return (
      <div className="border border-zinc-300 rounded-lg p-4 bg-blue-50">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <input
            type="text"
            placeholder="Company"
            value={editData.company}
            onChange={(e) => setEditData({ ...editData, company: e.target.value })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
          />
          <input
            type="text"
            placeholder="Job Title"
            value={editData.job_title}
            onChange={(e) => setEditData({ ...editData, job_title: e.target.value })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
          />
          <input
            type="text"
            placeholder="Location"
            value={editData.location}
            onChange={(e) => setEditData({ ...editData, location: e.target.value })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
          />
          <input
            type="text"
            placeholder="Salary Range"
            value={editData.salary_range}
            onChange={(e) => setEditData({ ...editData, salary_range: e.target.value })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
          />
          <input
            type="url"
            placeholder="Job URL"
            value={editData.job_url}
            onChange={(e) => setEditData({ ...editData, job_url: e.target.value })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 md:col-span-2"
          />
          <select
            value={editData.status}
            onChange={(e) => setEditData({ ...editData, status: e.target.value as AppliedJob['status'] })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
          >
            {STATUS_OPTIONS.map((status) => (
              <option key={status} value={status}>
                {status.charAt(0).toUpperCase() + status.slice(1)}
              </option>
            ))}
          </select>
          <textarea
            placeholder="Note"
            value={editData.notes}
            onChange={(e) => setEditData({ ...editData, notes: e.target.value })}
            className="px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 md:col-span-2"
            rows={2}
          />
        </div>
        <div className="flex gap-2 mt-3">
          <button
            onClick={handleSave}
            className="flex items-center gap-1 px-4 py-2 bg-lime-500 text-blue-900 rounded hover:bg-lime-400 font-semibold"
          >
            <Save size={16} />
            Save
          </button>
          <button
            onClick={handleCancel}
            className="flex items-center gap-1 px-4 py-2 bg-zinc-200 text-zinc-700 rounded hover:bg-zinc-300"
          >
            <X size={16} />
            Cancel
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="border border-zinc-200 rounded-lg p-4 bg-white hover:shadow-md transition-shadow">
      <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-3">
        <div className="flex-1">
          <div className="flex items-start gap-2 mb-2">
            <div>
              <h3 className="font-semibold text-lg text-blue-900">{job.company}</h3>
              <p className="text-zinc-700">{job.job_title}</p>
            </div>
          </div>

          {job.location && (
            <p className="text-sm text-zinc-600 mb-1">📍 {job.location}</p>
          )}

          {job.salary_range && (
            <p className="text-sm text-zinc-600 mb-1">💰 {job.salary_range}</p>
          )}

          {job.notes && (
            <p className="text-sm text-zinc-600 mt-2 italic">{job.notes}</p>
          )}

          <div className="flex items-center gap-2 mt-2 text-xs text-zinc-500">
            <span>
              Applied: {new Date(job.application_date).toLocaleDateString()}
            </span>
            <span>•</span>
            <span className={`px-2 py-1 rounded-full font-medium ${STATUS_COLORS[job.status]}`}>
              {job.status.charAt(0).toUpperCase() + job.status.slice(1)}
            </span>
            {job.source === 'agent' && (
              <>
                <span>•</span>
                <span className="text-lime-600">Added by agent</span>
              </>
            )}
          </div>
        </div>

        <div className="flex items-center gap-2">
          {job.job_url && (
            <a
              href={job.job_url}
              target="_blank"
              rel="noopener noreferrer"
              className="p-2 text-blue-600 hover:bg-blue-50 rounded"
              title="View job posting"
            >
              <ExternalLink size={18} />
            </a>
          )}
          <button
            onClick={() => setIsEditing(true)}
            className="p-2 text-zinc-600 hover:bg-zinc-100 rounded"
            title="Edit"
          >
            <Edit2 size={18} />
          </button>
          <button
            onClick={handleDelete}
            className={`p-2 rounded ${
              isDeleting
                ? 'bg-red-100 text-red-700'
                : 'text-zinc-600 hover:bg-zinc-100'
            }`}
            title={isDeleting ? 'Click again to confirm' : 'Delete'}
          >
            <Trash2 size={18} />
          </button>
        </div>
      </div>
    </div>
  )
}

export default JobRow
