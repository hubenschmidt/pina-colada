"use client"

import { useEffect, useState } from 'react'
import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from '../../lib/supabase'
import { fetchJobs, createJob, updateJob, deleteJob } from '../../lib/jobs-api'
import JobRow from './JobRow'
import JobForm from './JobForm'
import { Filter, RefreshCw } from 'lucide-react'

// In development, use local Postgres via API routes
// In production, use Supabase directly
const USE_API_ROUTES = process.env.NODE_ENV === "development"

export default function JobTracker() {
  const [jobs, setJobs] = useState<AppliedJob[]>([])
  const [filteredJobs, setFilteredJobs] = useState<AppliedJob[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [searchTerm, setSearchTerm] = useState('')

  const loadJobs = async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await fetchJobs()
      setJobs(data)
      setFilteredJobs(data)
    } catch (err) {
      console.error('Error fetching jobs:', err)
      const errorMsg = USE_API_ROUTES 
        ? 'Failed to load job applications. Please check your local Postgres connection.'
        : 'Failed to load job applications. Please check your Supabase connection.'
      setError(errorMsg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadJobs()
  }, [])

  useEffect(() => {
    let filtered = jobs

    // Filter by status
    if (filterStatus !== 'all') {
      filtered = filtered.filter((job) => job.status === filterStatus)
    }

    // Filter by search term
    if (searchTerm) {
      const search = searchTerm.toLowerCase()
      filtered = filtered.filter(
        (job) =>
          job.company.toLowerCase().includes(search) ||
          job.job_title.toLowerCase().includes(search) ||
          job.location?.toLowerCase().includes(search) ||
          job.notes?.toLowerCase().includes(search)
      )
    }

    setFilteredJobs(filtered)
  }, [jobs, filterStatus, searchTerm])

  const handleAddJob = async (jobData: Omit<AppliedJob, 'id' | 'created_at' | 'updated_at' | 'application_date'>) => {
    // Convert to AppliedJobInsert (company and job_title are required, others optional)
    const insertData: AppliedJobInsert = jobData as AppliedJobInsert
    const newJob = await createJob(insertData)
    setJobs((prev) => [newJob, ...prev])
  }

  const handleUpdateJob = async (id: string, updates: Partial<AppliedJob>) => {
    // Convert to AppliedJobUpdate (all fields optional)
    const updateData: AppliedJobUpdate = updates as AppliedJobUpdate
    const updatedJob = await updateJob(id, updateData)
    setJobs((prev) =>
      prev.map((job) => (job.id === id ? updatedJob : job))
    )
  }

  const handleDeleteJob = async (id: string) => {
    await deleteJob(id)
    setJobs((prev) => prev.filter((job) => job.id !== id))
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <RefreshCw className="animate-spin mx-auto mb-4 text-lime-500" size={48} />
          <p className="text-zinc-600">Loading job applications...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <p className="text-red-800 font-semibold mb-2">Error</p>
        <p className="text-red-600">{error}</p>
          <button
            onClick={loadJobs}
            className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            Retry
          </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header with stats */}
      <div className="bg-gradient-to-r from-lime-500 to-yellow-400 rounded-lg p-6 text-blue-900">
        <h1 className="text-3xl font-bold mb-2">Job Applications Tracker</h1>
        <p className="text-blue-800">
          Tracking {jobs.length} application{jobs.length !== 1 ? 's' : ''}
        </p>
      </div>

      {/* Add job form */}
      <JobForm onAdd={handleAddJob} />

      {/* Filters */}
      <div className="bg-white rounded-lg border border-zinc-200 p-4">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1">
            <label className="block text-sm font-medium text-zinc-700 mb-2">
              Search
            </label>
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Search by company, title, location..."
              className="w-full px-4 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-2 flex items-center gap-2">
              <Filter size={16} />
              Filter by Status
            </label>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="px-4 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
            >
              <option value="all">All Statuses</option>
              <option value="applied">Applied</option>
              <option value="interviewing">Interviewing</option>
              <option value="rejected">Rejected</option>
              <option value="offer">Offer</option>
              <option value="accepted">Accepted</option>
            </select>
          </div>

          <div className="flex items-end">
            <button
              onClick={loadJobs}
              className="flex items-center gap-2 px-4 py-2 bg-zinc-200 text-zinc-700 rounded hover:bg-zinc-300"
              title="Refresh"
            >
              <RefreshCw size={18} />
              Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Job list */}
      {filteredJobs.length === 0 ? (
        <div className="bg-zinc-50 border border-zinc-200 rounded-lg p-12 text-center">
          <p className="text-zinc-600 text-lg">
            {searchTerm || filterStatus !== 'all'
              ? 'No jobs match your filters'
              : 'No job applications yet. Add your first one above!'}
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredJobs.map((job) => (
            <JobRow
              key={job.id}
              job={job}
              onUpdate={handleUpdateJob}
              onDelete={handleDeleteJob}
            />
          ))}
        </div>
      )}
    </div>
  )
}
