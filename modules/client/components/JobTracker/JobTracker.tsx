"use client"

import { useEffect, useState, useMemo } from 'react'
import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from '../../lib/supabase'
import { fetchJobs, createJob, updateJob, deleteJob } from '../../lib/jobs-api'
import { DataTable, type Column, type PageData } from '../DataTable'
import JobForm from './JobForm'
import { Filter, RefreshCw, ExternalLink } from 'lucide-react'

// In development, use local Postgres via API routes
// In production, use Supabase directly
const USE_API_ROUTES = process.env.NODE_ENV === "development"

export default function JobTracker() {
  const [data, setData] = useState<PageData<AppliedJob> | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(25)
  const [sortBy, setSortBy] = useState<string>("application_date")
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC")

  const loadJobs = async () => {
    try {
      setLoading(true)
      setError(null)
      const pageData = await fetchJobs(page, limit, sortBy, sortDirection)
      setData(pageData)
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
  }, [page, limit, sortBy, sortDirection])

  const handleAddJob = async (jobData: Omit<AppliedJob, 'id' | 'created_at' | 'updated_at' | 'application_date'>) => {
    const insertData: AppliedJobInsert = jobData as AppliedJobInsert
    await createJob(insertData)
    await loadJobs()
  }

  const handleUpdateJob = async (id: string, updates: Partial<AppliedJob>) => {
    const updateData: AppliedJobUpdate = updates as AppliedJobUpdate
    await updateJob(id, updateData)
    await loadJobs()
  }

  const handleDeleteJob = async (id: string) => {
    await deleteJob(id)
    await loadJobs()
  }

  const columns: Column<AppliedJob>[] = useMemo(() => [
    {
      header: 'Company',
      accessor: 'company',
      sortable: true,
      sortKey: 'company',
    },
    {
      header: 'Job Title',
      accessor: 'job_title',
      sortable: true,
      sortKey: 'job_title',
    },
    {
      header: 'Location',
      accessor: 'location',
    },
    {
      header: 'Status',
      accessor: 'status',
      sortable: true,
      sortKey: 'status',
      render: (job) => (
        <span className={`px-2 py-1 rounded-full text-xs font-medium ${
          job.status === 'applied' ? 'bg-blue-100 text-blue-800' :
          job.status === 'interviewing' ? 'bg-yellow-100 text-yellow-800' :
          job.status === 'rejected' ? 'bg-red-100 text-red-800' :
          job.status === 'offer' ? 'bg-green-100 text-green-800' :
          'bg-purple-100 text-purple-800'
        }`}>
          {job.status.charAt(0).toUpperCase() + job.status.slice(1)}
        </span>
      ),
    },
    {
      header: 'Applied Date',
      accessor: 'application_date',
      sortable: true,
      sortKey: 'application_date',
      render: (job) => new Date(job.application_date).toLocaleDateString(),
    },
    {
      header: 'Actions',
      render: (job) => (
        <div className="flex items-center gap-2">
          {job.job_url && (
            <a
              href={job.job_url}
              target="_blank"
              rel="noopener noreferrer"
              className="p-1 text-blue-600 hover:bg-blue-50 rounded"
              title="View job posting"
            >
              <ExternalLink size={16} />
            </a>
          )}
        </div>
      ),
    },
  ], [])

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
          Tracking {data?.total || 0} application{(data?.total || 0) !== 1 ? 's' : ''}
        </p>
      </div>

      {/* Add job form */}
      <JobForm onAdd={handleAddJob} />

      {/* DataTable */}
      <DataTable<AppliedJob>
        data={data}
        columns={columns}
        onPageChange={setPage}
        pageValue={page}
        onPageSizeChange={(size) => {
          setLimit(size)
          setPage(1)
        }}
        pageSizeValue={limit}
        sortBy={sortBy}
        sortDirection={sortDirection}
        onSortChange={({ sortBy: newSortBy, direction }) => {
          setSortBy(newSortBy)
          setSortDirection(direction)
          setPage(1)
        }}
        rowKey={(job) => job.id}
        emptyText="No job applications yet. Add your first one above!"
      />
    </div>
  )
}
