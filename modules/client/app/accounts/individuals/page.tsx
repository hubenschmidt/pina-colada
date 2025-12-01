"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getIndividuals, Individual } from "../../../api";
import { Stack, Center, Loader, Text } from "@mantine/core";
import SearchHeader from "../../../components/SearchHeader/SearchHeader";
import { DataTable, Column, PageData } from "../../../components/DataTable/DataTable";

const columns: Column<Individual>[] = [
  {
    header: "Last Name",
    accessor: "last_name",
    sortable: true,
    sortKey: "last_name",
    render: (ind) => ind.last_name || "-",
  },
  {
    header: "First Name",
    accessor: "first_name",
    sortable: true,
    sortKey: "first_name",
    render: (ind) => ind.first_name || "-",
  },
  {
    header: "Title",
    accessor: "title",
    sortable: true,
    sortKey: "title",
    render: (ind) => ind.title || "-",
  },
  {
    header: "Industry",
    sortable: false,
    render: (ind) => ind.industries?.join(", ") || "-",
  },
  {
    header: "Email",
    accessor: "email",
    sortable: true,
    sortKey: "email",
    render: (ind) =>
      ind.email ? (
        <a
          href={`mailto:${ind.email}`}
          className="text-blue-600 dark:text-blue-400 hover:underline"
          onClick={(e) => e.stopPropagation()}
        >
          {ind.email}
        </a>
      ) : (
        "-"
      ),
  },
  {
    header: "Phone",
    accessor: "phone",
    sortable: true,
    sortKey: "phone",
    render: (ind) => ind.phone || "-",
  },
];

const IndividualsPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const [data, setData] = useState<PageData<Individual>>({
    items: [],
    currentPage: 1,
    totalPages: 1,
    total: 0,
    pageSize: 50,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  const fetchIndividuals = useCallback(async () => {
    try {
      setLoading(true);
      const result = await getIndividuals(
        page,
        limit,
        sortBy,
        sortDirection,
        searchQuery || undefined
      );
      setData(result);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load individuals");
    } finally {
      setLoading(false);
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
    }
  }, [page, limit, sortBy, sortDirection, searchQuery, dispatchPageLoading]);

  useEffect(() => {
    fetchIndividuals();
  }, [fetchIndividuals]);

  const handleRowClick = (ind: Individual) => {
    router.push(`/accounts/individuals/${ind.id}`);
  };

  if (loading && data.items.length === 0) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading individuals...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Individuals
        </h1>
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Individuals
      </h1>

      <SearchHeader
        placeholder="Search individuals..."
        buttonLabel="New Individual"
        onSearch={(query) => {
          setSearchQuery(query);
          setPage(1);
        }}
        onAdd={() => router.push("/accounts/individuals/new")}
      />

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
        rowKey={(ind) => ind.id}
        emptyText={
          searchQuery
            ? "No matching individuals found."
            : "No individuals yet. Add your first one above!"
        }
      />
    </Stack>
  );
};

export default IndividualsPage;
