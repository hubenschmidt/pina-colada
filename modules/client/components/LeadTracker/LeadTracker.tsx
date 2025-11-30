"use client";

import { useEffect, useState, useContext } from "react";
import { useRouter } from "next/navigation";
import { DataTable, type PageData } from "../DataTable/DataTable";
import { Search, X } from "lucide-react";
import { LeadTrackerConfig, BaseLead } from "./types/LeadTrackerTypes";
import { ProjectContext } from "../../context/projectContext";
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
  Badge,
} from "@mantine/core";
import { FolderKanban } from "lucide-react";

interface LeadTrackerProps<T extends BaseLead> {
  config: LeadTrackerConfig<T>;
}

const LeadTracker = <T extends BaseLead>({ config }: LeadTrackerProps<T>) => {
  const router = useRouter();
  const { projectState } = useContext(ProjectContext);
  const selectedProjectId = projectState.selectedProject?.id ?? null;

  const [data, setData] = useState<PageData<T> | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(config.defaultPageSize || 50);
  const [sortBy, setSortBy] = useState<string>(
    config.defaultSortBy || "created_at"
  );
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">(
    config.defaultSortDirection || "DESC"
  );
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
      .getLeads(page, limit, sortBy, sortDirection, searchQuery || undefined, selectedProjectId)
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
  }, [selectedProjectId]);

  useEffect(() => {
    if (data !== null && !loading) {
      const prevPage = data.currentPage || 1;
      const prevLimit = data.pageSize || 50;
      const isPagination = page !== prevPage || limit !== prevLimit;
      loadLeads(false, isPagination);
    }
  }, [page, limit, sortBy, sortDirection, searchQuery]);

  const handleRowClick = (lead: T) => {
    if (config.detailPagePath) {
      router.push(`${config.detailPagePath}/${lead.id}`);
      return;
    }
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

  const selectedProject = projectState.selectedProject;

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          {config.entityNamePlural}
        </h1>
        {selectedProject ? (
          <Badge variant="light" color="lime" leftSection={<FolderKanban className="h-3 w-3" />}>
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>

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
            <Button
              onClick={() => {
                if (config.newPagePath) {
                  router.push(config.newPagePath);
                  return;
                }
              }}
              color="lime"
            >
              New {config.entityName}
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
    </Stack>
  );
};

export default LeadTracker;
