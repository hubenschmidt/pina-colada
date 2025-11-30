"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import {
  Stack,
  Group,
  TextInput,
  Button,
  Badge,
  Anchor,
  Text,
  Box,
  Center,
  Loader,
} from "@mantine/core";
import { Search, X, FolderKanban } from "lucide-react";
import { DataTable, Column } from "../../components/DataTable/DataTable";
import { getTasks, Task, TasksPageData } from "../../api";
import { useProjectContext } from "../../context/projectContext";
import { usePageLoading } from "../../context/pageLoadingContext";

type SortDir = "ASC" | "DESC";

const TasksPage = () => {
  const router = useRouter();
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;
  const { dispatchPageLoading } = usePageLoading();

  const [data, setData] = useState<TasksPageData | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [sortBy, setSortBy] = useState("created_at");
  const [sortDirection, setSortDirection] = useState<SortDir>("DESC");
  const [searchQuery, setSearchQuery] = useState("");

  const scope = selectedProject ? "project" : "global";

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    try {
      const projectId = selectedProject?.id;
      const result = await getTasks(page, pageSize, sortBy, sortDirection, scope, projectId);
      setData(result);
    } catch (error) {
      console.error("Failed to fetch tasks:", error);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, sortBy, sortDirection, scope, selectedProject?.id]);

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  useEffect(() => {
    setPage(1);
  }, [selectedProject?.id]);

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    setPage(1);
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    setPage(1);
  };

  const columns: Column<Task>[] = [
    {
      header: "Task",
      accessor: "title",
      sortable: true,
      sortKey: "title",
    },
    {
      header: "Entity",
      render: (row) => {
        if (!row.entity.type || !row.entity.display_name) {
          return <Text c="dimmed">—</Text>;
        }
        return (
          <Group gap="xs">
            <Text size="xs" c="dimmed">
              {row.entity.type}
            </Text>
            {row.entity.url ? (
              <Anchor href={row.entity.url} size="sm">
                {row.entity.display_name}
              </Anchor>
            ) : (
              <Text size="sm">{row.entity.display_name}</Text>
            )}
          </Group>
        );
      },
    },
    {
      header: "Status",
      render: (row) =>
        row.status ? (
          <Badge size="sm" variant="light">
            {row.status.name}
          </Badge>
        ) : (
          <Text c="dimmed">—</Text>
        ),
      sortable: true,
      sortKey: "current_status_id",
    },
    {
      header: "Priority",
      render: (row) => {
        if (!row.priority) return <Text c="dimmed">—</Text>;
        const colorMap: Record<string, string> = {
          Low: "gray",
          Medium: "blue",
          High: "orange",
          Urgent: "red",
        };
        return (
          <Badge size="sm" variant="outline" color={colorMap[row.priority.name] || "gray"}>
            {row.priority.name}
          </Badge>
        );
      },
      sortable: true,
      sortKey: "priority_id",
    },
    {
      header: "Due Date",
      accessor: (row) => row.due_date || "—",
      sortable: true,
      sortKey: "due_date",
    },
    {
      header: "Created",
      accessor: (row) => (row.created_at ? row.created_at.slice(0, 10) : "—"),
      sortable: true,
      sortKey: "created_at",
    },
  ];

  if (loading && !data) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading tasks...</Text>
        </Stack>
      </Center>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Tasks</h1>
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

      <Stack gap="xs">
        <Group gap="md">
          <TextInput
            placeholder="Search tasks..."
            value={searchQuery}
            onChange={(e) => handleSearchChange(e.target.value)}
            leftSection={<Search size={20} />}
            rightSection={
              searchQuery ? (
                <button
                  onClick={handleClearSearch}
                  className="text-zinc-400 hover:text-zinc-600 dark:text-zinc-500 dark:hover:text-zinc-400"
                  aria-label="Clear search"
                >
                  <X size={18} />
                </button>
              ) : null
            }
            style={{ flex: 1 }}
          />
          <Button onClick={() => router.push("/tasks/new")} color="lime">
            New Task
          </Button>
        </Group>
        {searchQuery && (
          <Text size="sm" c="dimmed">
            Showing results for "{searchQuery}"
          </Text>
        )}
      </Stack>

      <Box pos="relative">
        <DataTable
          data={data}
          columns={columns}
          rowKey={(row) => row.id}
          pageValue={page}
          onPageChange={setPage}
          pageSizeValue={pageSize}
          onPageSizeChange={(size) => {
            setPageSize(size);
            setPage(1);
          }}
          sortBy={sortBy}
          sortDirection={sortDirection}
          onSortChange={({ sortBy: newSortBy, direction }) => {
            setSortBy(newSortBy);
            setSortDirection(direction);
            setPage(1);
          }}
          onRowClick={(task) => router.push(`/tasks/${task.id}`)}
          emptyText="No tasks yet. Add your first one above!"
        />
      </Box>
    </Stack>
  );
};

export default TasksPage;
