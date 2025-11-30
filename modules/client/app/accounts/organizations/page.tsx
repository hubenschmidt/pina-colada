"use client";

import { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getOrganizations, Organization } from "../../../api";
import { Stack, Center, Loader, Text } from "@mantine/core";
import SearchHeader from "../../../components/SearchHeader/SearchHeader";
import { DataTable, Column, PageData } from "../../../components/DataTable/DataTable";

const columns: Column<Organization>[] = [
  {
    header: "Name",
    accessor: "name",
    sortable: true,
    sortKey: "name",
  },
  {
    header: "Industry",
    sortable: true,
    sortKey: "industries",
    render: (org) => (org.industries?.length > 0 ? org.industries.join(", ") : "-"),
  },
  {
    header: "Funding",
    sortable: true,
    sortKey: "funding_stage",
    render: (org) => org.funding_stage || "-",
  },
  {
    header: "Employees",
    sortable: true,
    sortKey: "employee_count_range",
    render: (org) => org.employee_count_range || "-",
  },
  {
    header: "Description",
    accessor: "description",
    sortable: true,
    sortKey: "description",
    render: (org) => org.description || "-",
  },
  {
    header: "Website",
    accessor: "website",
    sortable: true,
    sortKey: "website",
    render: (org) =>
      org.website ? (
        <a
          href={org.website}
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-600 dark:text-blue-400 hover:underline"
          onClick={(e) => e.stopPropagation()}
        >
          {org.website}
        </a>
      ) : (
        "-"
      ),
  },
];

const OrganizationsPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        const data = await getOrganizations();
        setOrganizations(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load organizations");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchOrganizations();
  }, [dispatchPageLoading]);

  const filteredAndSortedData = useMemo((): PageData<Organization> => {
    let filtered = organizations;

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = organizations.filter(
        (org) =>
          org.name?.toLowerCase().includes(query) ||
          org.website?.toLowerCase().includes(query) ||
          org.industries?.some((ind) => ind.toLowerCase().includes(query))
      );
    }

    const sorted = [...filtered].sort((a, b) => {
      let comparison: number;

      if (sortBy === "updated_at") {
        const aDate = new Date(a.updated_at || 0).getTime();
        const bDate = new Date(b.updated_at || 0).getTime();
        comparison = aDate - bDate;
      } else if (sortBy === "industries") {
        const aVal = (a.industries?.join(", ") || "").toLowerCase();
        const bVal = (b.industries?.join(", ") || "").toLowerCase();
        comparison = aVal.localeCompare(bVal);
      } else if (sortBy === "funding_stage") {
        const aVal = (a.funding_stage || "").toLowerCase();
        const bVal = (b.funding_stage || "").toLowerCase();
        comparison = aVal.localeCompare(bVal);
      } else if (sortBy === "employee_count_range") {
        const aVal = (a.employee_count_range || "").toLowerCase();
        const bVal = (b.employee_count_range || "").toLowerCase();
        comparison = aVal.localeCompare(bVal);
      } else {
        const aVal = (a[sortBy as keyof Organization] as string || "").toLowerCase();
        const bVal = (b[sortBy as keyof Organization] as string || "").toLowerCase();
        comparison = aVal.localeCompare(bVal);
      }

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
  }, [organizations, searchQuery, sortBy, sortDirection, page, limit]);

  const handleRowClick = (org: Organization) => {
    router.push(`/accounts/organizations/${org.id}`);
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading organizations...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Organizations
        </h1>
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Organizations
      </h1>

      <SearchHeader
        placeholder="Search organizations..."
        buttonLabel="New Organization"
        onSearch={(query) => {
          setSearchQuery(query);
          setPage(1);
        }}
        onAdd={() => router.push("/accounts/organizations/new")}
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
        rowKey={(org) => org.id}
        emptyText={
          searchQuery
            ? "No matching organizations found."
            : "No organizations yet. Add your first one above!"
        }
      />
    </Stack>
  );
};

export default OrganizationsPage;
