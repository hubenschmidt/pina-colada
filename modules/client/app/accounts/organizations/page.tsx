"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getOrganizations, Organization } from "../../../api";
import { Stack, Center, Loader, Text } from "@mantine/core";
import SearchHeader from "../../../components/SearchHeader/SearchHeader";
import { DataTable, Column, PageData } from "../../../components/DataTable/DataTable";

type SortDir = "ASC" | "DESC";

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
  const [data, setData] = useState<PageData<Organization> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState<SortDir>("DESC");

  const fetchOrganizations = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await getOrganizations(page, limit, sortBy, sortDirection, searchQuery || undefined);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load organizations");
    } finally {
      setLoading(false);
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
    }
  }, [page, limit, sortBy, sortDirection, searchQuery, dispatchPageLoading]);

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  useEffect(() => {
    fetchOrganizations();
  }, [fetchOrganizations]);

  const handleSearchChange = (query: string) => {
    setSearchQuery(query);
    setPage(1);
  };

  const handleRowClick = (org: Organization) => {
    router.push(`/accounts/organizations/${org.id}`);
  };

  if (loading && !data) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading organizations...</Text>
        </Stack>
      </Center>
    );
  }

  if (error && !data) {
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
        onSearch={handleSearchChange}
        onAdd={() => router.push("/accounts/organizations/new")}
      />

      <DataTable
        data={data || { items: [], currentPage: 1, totalPages: 1, total: 0, pageSize: limit }}
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
