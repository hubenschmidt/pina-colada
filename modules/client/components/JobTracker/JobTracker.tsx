"use client"

import { useEffect, useState, useMemo } from 'react'
import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from '../../lib/supabase'
import { fetchJobs, createJob, updateJob, deleteJob } from '../../lib/jobs-api'
import { DataTable, type Column, type PageData } from '../DataTable'
import JobForm from './JobForm'
import JobEditModal from './JobEditModal'
import { RefreshCw, ExternalLink, Search, X } from 'lucide-react'

// In development, use local Postgres via API routes
// In production, use Supabase directly
const USE_API_ROUTES = process.env.NODE_ENV === "development"

const JobTracker = () => {
  const [data, setData] = useState<PageData<AppliedJob> | null>(null)
  const [loading, setLoading] = useState(true)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(25)
  const [sortBy, setSortBy] = useState<string>("date")
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC")
  const [selectedJob, setSelectedJob] = useState<AppliedJob | null>(null)
  const [modalOpened, setModalOpened] = useState(false)
  const [showLoadingBar, setShowLoadingBar] = useState(false)
  const [searchQuery, setSearchQuery] = useState("")

  const loadJobs = async (showFullLoading = false, showBar = false) => {
    if (showFullLoading) {
      setLoading(true)
    }
    if (showBar) {
      setIsRefreshing(true)
      setShowLoadingBar(true)
    }
    setError(null)

    try {
      const pageData = await fetchJobs(page, limit, sortBy, sortDirection, searchQuery || undefined)
      setData(pageData)
    } catch (err) {
      console.error('Error fetching jobs:', err)
      const errorMsg = USE_API_ROUTES 
        ? 'Failed to load job applications. Please check your local Postgres connection.'
        : 'Failed to load job applications. Please check your Supabase connection.'
      setError(errorMsg)
    } finally {
      setLoading(false)
      setIsRefreshing(false)
      setTimeout(() => setShowLoadingBar(false), 300)
    }
  }

  useEffect(() => {
    loadJobs(true)
  }, [])

  useEffect(() => {
    if (data !== null && !loading) {
      const prevPage = data.currentPage || 1
      const prevLimit = data.pageSize || 25
      const isPagination = page !== prevPage || limit !== prevLimit
      loadJobs(false, isPagination)
    }
  }, [page, limit, sortBy, sortDirection, searchQuery])

  const handleAddJob = async (jobData: Omit<AppliedJob, 'id' | 'created_at' | 'updated_at' | 'date'>) => {
    const insertData: AppliedJobInsert = jobData as AppliedJobInsert
    await createJob(insertData)
    await loadJobs(false)
  }

  const handleUpdateJob = async (id: string, updates: Partial<AppliedJob>) => {
    const updateData: AppliedJobUpdate = updates as AppliedJobUpdate
    await updateJob(id, updateData)
    await loadJobs(false)
  }

  const handleDeleteJob = async (id: string) => {
    await deleteJob(id)
    await loadJobs(false)
  }

  const handleRowClick = (job: AppliedJob) => {
    setSelectedJob(job)
    setModalOpened(true)
  }

  const handleModalClose = () => {
    setModalOpened(false)
    setSelectedJob(null)
  }

  const handleSearchChange = (value: string) => {
    setSearchQuery(value)
    setPage(1) // Reset to first page when searching
  }

  const handleClearSearch = () => {
    setSearchQuery("")
    setPage(1)
  }

  const columns: Column<AppliedJob>[] = useMemo(() => [
    {
      header: 'Company',
      accessor: 'company',
      sortable: true,
      sortKey: 'company',
      width: '12%',
    },
    {
      header: 'Job Title',
      accessor: 'job_title',
      sortable: true,
      sortKey: 'job_title',
      width: '15%',
    },
    {
      header: 'Date',
      accessor: 'date',
      sortable: true,
      sortKey: 'date',
      width: '10%',
      render: (job) => job.date ? new Date(job.date).toLocaleDateString() : <span className="text-zinc-400">—</span>,
    },
    {
      header: 'Notes',
      accessor: 'notes',
      width: '20%',
      render: (job) => {
        if (!job.notes) return <span className="text-zinc-400">—</span>
        const truncated = job.notes.length > 50 
          ? `${job.notes.substring(0, 50)}...` 
          : job.notes
        return (
          <span 
            title={job.notes.length > 50 ? job.notes : undefined}
            className="text-sm"
          >
            {truncated}
          </span>
        )
      },
    },
    {
      header: 'Resume',
      accessor: 'resume',
      sortable: true,
      sortKey: 'resume',
      width: '10%',
      render: (job) => job.resume ? new Date(job.resume).toLocaleDateString() : <span className="text-zinc-400">—</span>,
    },
    {
      header: 'URL',
      accessor: 'job_url',
      width: '8%',
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
          {!job.job_url && <span className="text-zinc-400">—</span>}
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
          onClick={() => loadJobs(true)}
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

      {/* Search bar */}
      <div className="relative">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-zinc-400" size={20} />
          <input
            type="text"
            placeholder="Search by company or job title..."
            value={searchQuery}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="w-full pl-10 pr-10 py-2 border border-zinc-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-lime-500 focus:border-transparent"
          />
          {searchQuery && (
            <button
              onClick={handleClearSearch}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-zinc-400 hover:text-zinc-600"
              aria-label="Clear search"
            >
              <X size={18} />
            </button>
          )}
        </div>
        {searchQuery && (
          <p className="mt-2 text-sm text-zinc-600">
            Showing results for "{searchQuery}"
          </p>
        )}
      </div>

      {/* DataTable */}
      <div className="relative">
        {showLoadingBar && isRefreshing && (
          <div className="absolute top-0 left-0 right-0 h-0.5 bg-zinc-100 z-10 overflow-hidden">
            <div className="h-full bg-zinc-300" style={{ width: '40%', transition: 'width 0.3s ease' }} />
          </div>
        )}
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
          onRowClick={handleRowClick}
          rowKey={(job) => job.id}
          emptyText="No job applications yet. Add your first one above!"
        />
      </div>

      {/* Edit Modal */}
      <JobEditModal
        job={selectedJob}
        opened={modalOpened}
        onClose={handleModalClose}
        onUpdate={handleUpdateJob}
        onDelete={handleDeleteJob}
      />
    </div>
  )
}

export default JobTracker
