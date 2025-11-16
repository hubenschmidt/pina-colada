"use client";

import { useEffect, useState } from "react";
import { DataTable, type PageData } from "../DataTable";
import { RefreshCw, Search, X } from "lucide-react";
import { LeadTrackerConfig, BaseLead } from "./LeadTrackerConfig";

interface LeadTrackerProps<T extends BaseLead> {
  config: LeadTrackerConfig<T>;
}

function LeadTracker<T extends BaseLead>({ config }: LeadTrackerProps<T>) {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [data, setData] = useState<PageData<T> | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(config.defaultPageSize || 25);
  const [sortBy, setSortBy] = useState<string>(
    config.defaultSortBy || "created_at"
  );
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">(
    config.defaultSortDirection || "DESC"
  );
  const [selectedLead, setSelectedLead] = useState<T | null>(null);
  const [modalOpened, setModalOpened] = useState(false);
  const [showLoadingBar, setShowLoadingBar] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const enableSearch = config.enableSearch !== false;

  const loadLeads = async (showFullLoading = false, showBar = false) => {
    if (showFullLoading) {
      setLoading(true);
    }
    if (showBar) {
      setIsRefreshing(true);
      setShowLoadingBar(true);
    }
    setError(null);

    try {
      const pageData = await config.api.getLeads(
        page,
        limit,
        sortBy,
        sortDirection,
        searchQuery || undefined
      );
      setData(pageData);
    } catch (err) {
      console.error(`Error fetching ${config.entityNamePlural}:`, err);
      setError(
        `Failed to load ${config.entityNamePlural.toLowerCase()}. Please try again.`
      );
    } finally {
      setLoading(false);
      setIsRefreshing(false);
      setTimeout(() => setShowLoadingBar(false), 300);
    }
  };

  useEffect(() => {
    loadLeads(true);
  }, []);

  useEffect(() => {
    if (data !== null && !loading) {
      const prevPage = data.currentPage || 1;
      const prevLimit = data.pageSize || 25;
      const isPagination = page !== prevPage || limit !== prevLimit;
      loadLeads(false, isPagination);
    }
  }, [page, limit, sortBy, sortDirection, searchQuery]);

  const handleAddLead = async (
    leadData: Omit<T, "id" | "created_at" | "updated_at">
  ) => {
    await config.api.createLead(leadData);
    await loadLeads(false);
    setIsFormOpen(false);
  };

  const handleUpdateLead = async (id: string, updates: Partial<T>) => {
    await config.api.updateLead(id, updates);
    await loadLeads(false);
  };

  const handleDeleteLead = async (id: string) => {
    await config.api.deleteLead(id);
    await loadLeads(false);
  };

  const handleRowClick = (lead: T) => {
    setSelectedLead(lead);
    setModalOpened(true);
  };

  const handleModalClose = () => {
    setModalOpened(false);
    setSelectedLead(null);
  };

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    setPage(1);
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    setPage(1);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <RefreshCw
            className="animate-spin mx-auto mb-4 text-lime-500"
            size={48}
          />
          <p className="text-zinc-600">
            Loading {config.entityNamePlural.toLowerCase()}...
          </p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
        <p className="text-red-800 font-semibold mb-2">Error</p>
        <p className="text-red-600">{error}</p>
        <button
          onClick={() => loadLeads(true)}
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  const FormComponent = config.FormComponent;
  const EditModalComponent = config.EditModalComponent;

  return (
    <div className="space-y-6">
      {/* Lead form */}
      <FormComponent
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        onAdd={handleAddLead}
      />

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
                placeholder={
                  config.searchPlaceholder ||
                  `Search ${config.entityNamePlural.toLowerCase()}...`
                }
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
              Add {config.entityName}
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
          columns={config.columns}
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
          onRowClick={handleRowClick}
          rowKey={(lead) => lead.id}
          emptyText={
            config.emptyMessage ||
            `No ${config.entityNamePlural.toLowerCase()} yet. Add your first one above!`
          }
        />
      </div>

      {/* Edit Modal */}
      <EditModalComponent
        lead={selectedLead}
        opened={modalOpened}
        onClose={handleModalClose}
        onUpdate={handleUpdateLead}
        onDelete={handleDeleteLead}
      />
    </div>
  );
}

export default LeadTracker;
