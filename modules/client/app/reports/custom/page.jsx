"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  Stack,
  Center,
  Loader,
  Text,
  Button,
  Group,
  ActionIcon,
  Anchor,
  Badge,
  TextInput,
  Box,
} from "@mantine/core";
import { Plus, Trash2, FolderKanban, Search, X } from "lucide-react";
import { getSavedReports, deleteSavedReport } from "../../../api";
import { useProjectContext } from "../../../context/projectContext";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { DataTable } from "../../../components/DataTable/DataTable";
import { DeleteConfirmBanner } from "../../../components/DeleteConfirmBanner";

const CustomReportsPage = () => {
  const router = useRouter();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState("DESC");
  const [searchQuery, setSearchQuery] = useState("");
  const [showLoadingBar, setShowLoadingBar] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState(null);
  const [deleting, setDeleting] = useState(false);

  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const fetchReports = async (showFullLoading = false, showBar = false) => {
    if (showFullLoading) {
      setLoading(true);
    }
    if (showBar) {
      setIsRefreshing(true);
      setShowLoadingBar(true);
    }
    setError(null);

    try {
      const result = await getSavedReports(
        selectedProject?.id,
        true,
        page,
        limit,
        sortBy,
        sortDirection,
        searchQuery || undefined,
      );
      setData({
        items: result.items,
        currentPage: result.currentPage,
        totalPages: result.totalPages,
        total: result.total,
        pageSize: result.pageSize,
      });
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load saved reports",
      );
    } finally {
      setLoading(false);
      setIsRefreshing(false);
      setTimeout(() => setShowLoadingBar(false), 300);
    }
  };

  useEffect(() => {
    fetchReports(true);
  }, [selectedProject?.id]);

  useEffect(() => {
    if (data !== null && !loading) {
      const isPagination =
        page !== (data.currentPage || 1) || limit !== (data.pageSize || 50);
      fetchReports(false, isPagination);
    }
  }, [page, limit, sortBy, sortDirection, searchQuery]);

  const handleDeleteClick = (report, e) => {
    e.stopPropagation();
    setDeleteConfirm(report);
  };

  const handleDeleteConfirm = async () => {
    if (!deleteConfirm) return;

    setDeleting(true);
    try {
      await deleteSavedReport(deleteConfirm.id);
      setDeleteConfirm(null);
      fetchReports(false, true);
    } catch (err) {
      console.error("Delete failed:", err);
    } finally {
      setDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteConfirm(null);
  };

  const handleRowClick = (report) => {
    router.push(`/reports/custom/${report.id}`);
  };

  const handleSearchChange = (value) => {
    setSearchQuery(value);
    setPage(1);
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    setPage(1);
  };

  const columns = [
    {
      header: "Name",
      sortable: true,
      sortKey: "name",
      render: (report) => (
        <Anchor
          component={Link}
          href={`/reports/custom/${report.id}`}
          fw={500}
          c="inherit"
          underline="hover"
          onClick={(e) => e.stopPropagation()}
        >
          {report.name}
        </Anchor>
      ),
    },
    {
      header: "Scope",
      render: (report) =>
        report.query_definition?.primary_entity === "leads" &&
        selectedProject ? (
          <Badge variant="light" color="lime">
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        ),
    },
    {
      header: "Description",
      sortable: true,
      sortKey: "description",
      render: (report) => (
        <Text size="sm" lineClamp={1}>
          {report.description || "-"}
        </Text>
      ),
    },
    {
      header: "Entity",
      render: (report) => (
        <Text size="sm" tt="capitalize">
          {report.query_definition?.primary_entity || "-"}
        </Text>
      ),
    },
    {
      header: "Updated",
      sortable: true,
      sortKey: "updated_at",
      render: (report) => (
        <Text size="sm">
          {report.updated_at
            ? new Date(report.updated_at).toLocaleDateString()
            : "-"}
        </Text>
      ),
    },
    {
      header: "",
      width: 50,
      render: (report) => (
        <ActionIcon
          variant="subtle"
          color="red"
          onClick={(e) => handleDeleteClick(report, e)}
        >
          <Trash2 className="h-4 w-4" />
        </ActionIcon>
      ),
    },
  ];

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading saved reports...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Custom Reports
        </h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Custom Reports
        </h1>
        {selectedProject ? (
          <Badge
            variant="light"
            color="lime"
            leftSection={<FolderKanban className="h-3 w-3" />}
          >
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>

      <Stack gap="xs">
        <Group gap="md">
          <TextInput
            placeholder="Search reports..."
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
          />

          <Button
            leftSection={<Plus className="h-4 w-4" />}
            color="lime"
            onClick={() => router.push("/reports/custom/new")}
          >
            New Report
          </Button>
        </Group>
        {searchQuery && (
          <Text size="sm" c="dimmed">
            Showing results for "{searchQuery}"
          </Text>
        )}
      </Stack>

      {deleteConfirm && (
        <DeleteConfirmBanner
          itemName={deleteConfirm.name}
          onConfirm={handleDeleteConfirm}
          onCancel={handleDeleteCancel}
          loading={deleting}
        />
      )}

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
          columns={columns}
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
          rowKey={(report) => report.id}
          emptyText="No saved reports yet. Create your first report above!"
        />
      </Box>
    </Stack>
  );
};

export default CustomReportsPage;
