"use client";

import { useEffect, useState, useContext } from "react";
import { useRouter } from "next/navigation";
import { DataTable } from "../DataTable/DataTable";
import SearchBox from "../SearchBox/SearchBox";

import { ProjectContext } from "../../context/projectContext";
import { Stack, Center, Loader, Alert, Group, Button, Box, Text, Badge } from "@mantine/core";
import { FolderKanban } from "lucide-react";

const LeadTracker = ({ config }) => {
  const router = useRouter();
  const { projectState } = useContext(ProjectContext);
  const selectedProjectId = projectState.selectedProject?.id ?? null;

  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(config.defaultPageSize || 50);
  const [sortBy, setSortBy] = useState(config.defaultSortBy || "created_at");
  const [sortDirection, setSortDirection] = useState(config.defaultSortDirection || "DESC");
  const [showLoadingBar, setShowLoadingBar] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const enableSearch = config.enableSearch !== false;

  const loadLeads = (showFullLoading = false, showBar = false, overrideSearch) => {
    if (showFullLoading) {
      setLoading(true);
    }
    if (showBar) {
      setIsRefreshing(true);
      setShowLoadingBar(true);
    }
    setError(null);

    const search = overrideSearch !== undefined ? overrideSearch : searchQuery;
    config.api
      .getLeads(page, limit, sortBy, sortDirection, search || undefined, selectedProjectId)
      .then((pageData) => {
        setData(pageData);
      })
      .catch((err) => {
        console.error(`Error fetching ${config.entityNamePlural}:`, err);
        setError(`Failed to load ${config.entityNamePlural.toLowerCase()}. Please try again.`);
      })
      .finally(() => {
        setLoading(false);
        setIsRefreshing(false);
        setTimeout(() => setShowLoadingBar(false), 300);
      });
  };

  useEffect(() => {
    setSearchQuery("");
    setPage(1);
    loadLeads(true, false, "");
  }, [selectedProjectId]);

  useEffect(() => {
    if (data !== null && !loading) {
      const prevPage = data.currentPage || 1;
      const prevLimit = data.pageSize || 50;
      const isPagination = page !== prevPage || limit !== prevLimit;
      loadLeads(false, isPagination);
    }
  }, [page, limit, sortBy, sortDirection, searchQuery]);

  const handleRowClick = (lead) => {
    if (!config.detailPagePath) return;
    router.push(`${config.detailPagePath}/${lead.id}`);
  };

  const handleCellUpdate = async (row, field, value) => {
    await config.api.updateLead(row.id, { [field]: value });
  };

  const handleSearch = (query) => {
    setSearchQuery(query);
    setPage(1);
    if (query === "") {
      loadLeads(false, true, "");
    }
  };

  const fetchPreview = async (query) => {
    if (!config.getSuggestionLabel) return [];
    const result = await config.api.getLeads(1, 4, sortBy, sortDirection, query, selectedProjectId);
    return result.items.map((item) => {
      const label = config.getSuggestionLabel(item);
      const value = config.getSuggestionValue ? config.getSuggestionValue(item) : label;
      return { label, value };
    });
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading {config.entityNamePlural.toLowerCase()}...</Text>
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
        <Group gap="md">
          <SearchBox
            key={selectedProjectId ?? "global"}
            placeholder={
              config.searchPlaceholder ||
              `Search ${config.entityNamePlural.toLowerCase()}... (Enter to search)`
            }
            onSearch={handleSearch}
            fetchPreview={config.getSuggestionLabel ? fetchPreview : undefined}
          />

          <Button
            onClick={() => {
              if (config.newPagePath) {
                router.push(config.newPagePath);
                return;
              }
            }}
            color="lime">
            New {config.entityName}
          </Button>
        </Group>
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
            style={{ zIndex: 10, overflow: "hidden" }}>
            <Box h="100%" bg="gray.3" style={{ width: "40%", transition: "width 0.3s ease" }} />
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
          onCellUpdate={handleCellUpdate}
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
