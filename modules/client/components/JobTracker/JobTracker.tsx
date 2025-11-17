"use client";

import { useEffect, useState } from "react";
import { DataTable, type PageData } from "../DataTable";
import { RefreshCw, Search, X } from "lucide-react";
import { CreatedJob } from "../../types/types";
import * as api from "../../api";
import { JobForm } from "./JobForm";
import { JobEditModal } from "./JobEditModal";

interface JobTrackerProps {
  // Optional props for customization
  defaultPageSize?: number;
  enableSearch?: boolean;
  searchPlaceholder?: string;
}

export function JobTracker({
  defaultPageSize = 25,
  enableSearch = true,
  searchPlaceholder = "Search jobs...",
}: JobTrackerProps) {
  // Data state
  const [data, setData] = useState<PageData<CreatedJob> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Pagination & sorting state
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(defaultPageSize);
  const [sortBy, setSortBy] = useState("date");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  // Search state
  const [searchQuery, setSearchQuery] = useState("");

  // UI state
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showLoadingBar, setShowLoadingBar] = useState(false);
  const [selectedJob, setSelectedJob] = useState<CreatedJob | null>(null);
  const [modalOpened, setModalOpened] = useState(false);
  const [isFormOpen, setIsFormOpen] = useState(false);

  // Load jobs from API
  const loadJobs = async (showFullLoading = false, showBar = false) => {
    if (showFullLoading) setLoading(true);
    if (showBar) {
      setIsRefreshing(true);
      setShowLoadingBar(true);
    }
    setError(null);

    try {
      const pageData = await api.getJobs(
        page,
        limit,
        sortBy,
        sortDirection,
        searchQuery || undefined
      );
      setData(pageData);
    } catch (err) {
      console.error("Error fetching jobs:", err);
      setError("Failed to load jobs. Please try again.");
    } finally {
      setLoading(false);
      setIsRefreshing(false);
      setTimeout(() => setShowLoadingBar(false), 300);
    }
  };

  // Initial load
  useEffect(() => {
    loadJobs(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Reload when params change
  useEffect(() => {
    if (data !== null && !loading) {
      const prevPage = data.currentPage || 1;
      const prevLimit = data.pageSize || 25;
      const isPagination = page !== prevPage || limit !== prevLimit;
      loadJobs(false, isPagination);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, limit, sortBy, sortDirection, searchQuery]);

  // Search handlers
  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    setPage(1);
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    setPage(1);
  };

  // Loading state
  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <RefreshCw
            className="animate-spin mx-auto mb-4 text-lime-500"
            size={48}
          />
          <p className="text-zinc-600">Loading jobs...</p>
        </div>
      </div>
    );
  }

  // Error state
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
    );
  }

  return (
    <div className="space-y-6">
      {/* Search bar and Add button */}
      {enableSearch && (
        <div className="relative">
          <div className="flex gap-3 items-center">
            <div className="relative flex-1">
              <Search
                className="absolute left-3 top-1/2 transform -translate-y-1/2 text-zinc-400"
                size={20}
              />
              <input
                type="text"
                placeholder={searchPlaceholder}
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
            <button
              onClick={() => setIsFormOpen(true)}
              className="px-4 py-2 bg-zinc-700 text-white rounded-lg hover:bg-zinc-600 font-medium whitespace-nowrap"
            >
              Add Job
            </button>
          </div>
          {searchQuery && (
            <p className="mt-2 text-sm text-zinc-600">
              Showing results for "{searchQuery}"
            </p>
          )}
        </div>
      )}

      {/* DataTable */}
      <div className="relative">
        {showLoadingBar && isRefreshing && (
          <div className="absolute top-0 left-0 right-0 h-0.5 bg-zinc-100 z-10 overflow-hidden">
            <div
              className="h-full bg-zinc-300"
              style={{ width: "40%", transition: "width 0.3s ease" }}
            />
          </div>
        )}
        <DataTable
          data={data}
          columns={[
            {
              header: "Company",
              accessor: "company",
              sortable: true,
              sortKey: "company",
            },
            {
              header: "Job Title",
              accessor: "job_title",
              sortable: true,
              sortKey: "job_title",
            },
            {
              header: "Date",
              accessor: (job) => job.date ? job.date.split("T")[0] : "—",
              sortable: true,
              sortKey: "date",
            },
            {
              header: "Status",
              accessor: "status",
              sortable: true,
              sortKey: "status",
            },
          ]}
          onPageChange={setPage}
          pageValue={page}
          onPageSizeChange={(size) => {
            setLimit(size);
            setPage(1);
          }}
          pageSizeValue={limit}
          sortBy={sortBy}
          sortDirection={sortDirection}
          onSortChange={({ sortBy: newSortBy, direction }) => {
            setSortBy(newSortBy);
            setSortDirection(direction);
            setPage(1);
          }}
          onRowClick={(job) => {
            setSelectedJob(job);
            setModalOpened(true);
          }}
          rowKey={(job) => job.id}
          emptyText="No jobs yet."
        />
      </div>

      {/* Job Form */}
      <JobForm
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        onAdd={async () => {
          await loadJobs(false);
        }}
      />

      {/* Edit Modal */}
      <JobEditModal
        job={selectedJob}
        opened={modalOpened}
        onClose={() => {
          setModalOpened(false);
          setSelectedJob(null);
        }}
        onUpdate={async () => {
          await loadJobs(false);
        }}
        onDelete={async () => {
          await loadJobs(false);
        }}
      />
    </div>
  );
}
