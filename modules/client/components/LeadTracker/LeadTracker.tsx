"use client";

import { useEffect, useState } from "react";
import { DataTable, type PageData } from "../DataTable";
import { Search, X } from "lucide-react";
import { LeadTrackerConfig, BaseLead } from "./LeadTrackerConfig";
import {
  Stack,
  Center,
  Loader,
  Alert,
  Group,
  TextInput,
  Button,
  Box,
  Text,
} from "@mantine/core";

interface LeadTrackerProps<T extends BaseLead> {
  config: LeadTrackerConfig<T>;
}

const LeadTracker = <T extends BaseLead>({ config }: LeadTrackerProps<T>) => {
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

  const loadLeads = (showFullLoading = false, showBar = false) => {
    if (showFullLoading) {
      setLoading(true);
    }
    if (showBar) {
      setIsRefreshing(true);
      setShowLoadingBar(true);
    }
    setError(null);

    config.api
      .getLeads(page, limit, sortBy, sortDirection, searchQuery || undefined)
      .then((pageData) => {
        setData(pageData);
      })
      .catch((err) => {
        console.error(`Error fetching ${config.entityNamePlural}:`, err);
        setError(
          `Failed to load ${config.entityNamePlural.toLowerCase()}. Please try again.`
        );
      })
      .finally(() => {
        setLoading(false);
        setIsRefreshing(false);
        setTimeout(() => setShowLoadingBar(false), 300);
      });
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
    return config.api
      .createLead(leadData)
      .then(() => {
        loadLeads(false);
        setIsFormOpen(false);
      })
      .catch((err) => {
        console.error(`Error creating ${config.entityName}:`, err);
      });
  };

  const handleUpdateLead = async (id: string, updates: Partial<T>) => {
    return config.api
      .updateLead(id, updates)
      .then(() => {
        loadLeads(false);
      })
      .catch((err) => {
        console.error(`Error updating ${config.entityName}:`, err);
      });
  };

  const handleDeleteLead = async (id: string) => {
    return config.api
      .deleteLead(id)
      .then(() => {
        loadLeads(false);
      })
      .catch((err) => {
        console.error(`Error deleting ${config.entityName}:`, err);
      });
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
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">
            Loading {config.entityNamePlural.toLowerCase()}...
          </Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Alert color="red" title="Error">
        <Stack gap="md" align="center">
          <Text>{error}</Text>
          <Button color="red" onClick={() => loadLeads(true)}>
            Retry
          </Button>
        </Stack>
      </Alert>
    );
  }

  const FormComponent = config.FormComponent;

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        {config.entityName} Tracker
      </h1>

      {/* Lead form */}
      <FormComponent
        isOpen={isFormOpen}
        onClose={() => setIsFormOpen(false)}
        onAdd={handleAddLead}
      />

      {/* Search bar and Add button */}
      {enableSearch && (
        <Stack gap="xs">
          <Group gap="md">
            <TextInput
              placeholder={
                config.searchPlaceholder ||
                `Search ${config.entityNamePlural.toLowerCase()}...`
              }
              value={searchQuery}
              onChange={(e) => handleSearchChange(e.target.value)}
              leftSection={<Search size={20} />}
              rightSection={
                searchQuery && (
                  <button
                    onClick={handleClearSearch}
                    className="text-zinc-400 hover:text-zinc-600 dark:text-zinc-500 dark:hover:text-zinc-400"
                    aria-label="Clear search"
                  >
                    <X size={18} />
                  </button>
                )
              }
              style={{ flex: 1 }}
              styles={{
                input: {
                  transition: "background-color 0.2s ease",
                  "&:hover": {
                    backgroundColor: "var(--input-background)",
                    filter: "brightness(0.97)",
                  },
                },
              }}
            />
            <Button onClick={() => setIsFormOpen(true)} variant="default">
              Add {config.entityName}
            </Button>
          </Group>
          {searchQuery && (
            <Text size="sm" c="dimmed">
              Showing results for "{searchQuery}"
            </Text>
          )}
        </Stack>
      )}

      {/* DataTable */}
      <Box pos="relative">
        {showLoadingBar && isRefreshing && (
          <Box
            pos="absolute"
            top={0}
            left={0}
            right={0}
            h={2}
            bg="gray.1"
            style={{ zIndex: 10, overflow: "hidden" }}
          >
            <Box
              h="100%"
              bg="gray.3"
              style={{ width: "40%", transition: "width 0.3s ease" }}
            />
          </Box>
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
      </Box>

      {/* Edit Modal */}
      <FormComponent
        isOpen={modalOpened}
        onClose={handleModalClose}
        lead={selectedLead}
        onUpdate={handleUpdateLead}
        onDelete={handleDeleteLead}
      />
    </Stack>
  );
};

export default LeadTracker;
