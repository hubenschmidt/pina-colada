"use client";

import { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../context/pageLoadingContext";
import { getProjects, Project } from "../../api";
import { Stack, Center, Loader, Text } from "@mantine/core";
import SearchHeader from "../../components/SearchHeader/SearchHeader";
import { DataTable, Column, PageData } from "../../components/DataTable/DataTable";

const columns: Column<Project>[] = [
  {
    header: "Name",
    accessor: "name",
    sortable: true,
    sortKey: "name",
  },
  {
    header: "Status",
    accessor: "status",
    sortable: true,
    sortKey: "status",
    render: (project) => project.status || "-",
  },
  {
    header: "Description",
    accessor: "description",
    sortable: true,
    sortKey: "description",
    render: (project) => project.description || "-",
  },
  {
    header: "Start Date",
    accessor: "start_date",
    sortable: true,
    sortKey: "start_date",
    render: (project) => project.start_date || "-",
  },
  {
    header: "End Date",
    accessor: "end_date",
    sortable: true,
    sortKey: "end_date",
    render: (project) => project.end_date || "-",
  },
];

const ProjectsPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const data = await getProjects();
        setProjects(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load projects");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchProjects();
  }, [dispatchPageLoading]);

  const filteredAndSortedData = useMemo((): PageData<Project> => {
    let filtered = projects;

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = projects.filter(
        (project) =>
          project.name?.toLowerCase().includes(query) ||
          project.description?.toLowerCase().includes(query) ||
          project.status?.toLowerCase().includes(query)
      );
    }

    const getComparison = (a: Project, b: Project): number => {
      if (sortBy === "updated_at") {
        return new Date(a.updated_at || 0).getTime() - new Date(b.updated_at || 0).getTime();
      }
      const aVal = (a[sortBy as keyof Project] as string || "").toLowerCase();
      const bVal = (b[sortBy as keyof Project] as string || "").toLowerCase();
      return aVal.localeCompare(bVal);
    };

    const sorted = [...filtered].sort((a, b) => {
      const comparison = getComparison(a, b);
      return sortDirection === "ASC" ? comparison : -comparison;
    });

    const totalPages = Math.ceil(sorted.length / limit);
    const startIndex = (page - 1) * limit;
    const items = sorted.slice(startIndex, startIndex + limit);

    return {
      items,
      currentPage: page,
      totalPages,
      total: sorted.length,
      pageSize: limit,
    };
  }, [projects, searchQuery, sortBy, sortDirection, page, limit]);

  const handleRowClick = (project: Project) => {
    router.push(`/projects/${project.id}`);
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading projects...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Projects
        </h1>
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Projects
      </h1>

      <SearchHeader
        placeholder="Search projects..."
        buttonLabel="New Project"
        onSearch={(query) => {
          setSearchQuery(query);
          setPage(1);
        }}
        onAdd={() => router.push("/projects/new")}
      />

      <DataTable
        data={filteredAndSortedData}
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
        rowKey={(project) => project.id}
        emptyText={
          searchQuery
            ? "No matching projects found."
            : "No projects yet. Add your first one above!"
        }
      />
    </Stack>
  );
};

export default ProjectsPage;
