"use client";

import { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getIndividuals, Individual } from "../../../api";
import { Stack, Center, Loader, Text } from "@mantine/core";
import SearchHeader from "../../../components/SearchHeader";
import { DataTable, Column, PageData } from "../../../components/DataTable/DataTable";

const columns: Column<Individual>[] = [
  {
    header: "Name",
    sortable: true,
    sortKey: "name",
    render: (ind) => `${ind.first_name} ${ind.last_name}`,
  },
  {
    header: "Title",
    accessor: "title",
    sortable: true,
    sortKey: "title",
    render: (ind) => ind.title || "-",
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
  const [individuals, setIndividuals] = useState<Individual[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);
  const [sortBy, setSortBy] = useState("name");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("ASC");

  useEffect(() => {
    const fetchIndividuals = async () => {
      try {
        const data = await getIndividuals();
        setIndividuals(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load individuals");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchIndividuals();
  }, [dispatchPageLoading]);

  const filteredAndSortedData = useMemo((): PageData<Individual> => {
    let filtered = individuals;

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = individuals.filter(
        (ind) =>
          ind.first_name?.toLowerCase().includes(query) ||
          ind.last_name?.toLowerCase().includes(query) ||
          ind.email?.toLowerCase().includes(query) ||
          ind.title?.toLowerCase().includes(query)
      );
    }

    const sorted = [...filtered].sort((a, b) => {
      if (sortBy === "name") {
        const aVal = `${a.first_name} ${a.last_name}`.toLowerCase();
        const bVal = `${b.first_name} ${b.last_name}`.toLowerCase();
        if (sortDirection === "ASC") {
          return aVal.localeCompare(bVal);
        }
        return bVal.localeCompare(aVal);
      }
      
      const aVal = (a[sortBy as keyof Individual] as string || "").toLowerCase();
      const bVal = (b[sortBy as keyof Individual] as string || "").toLowerCase();
      if (sortDirection === "ASC") {
        return aVal.localeCompare(bVal);
      }
      return bVal.localeCompare(aVal);
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
  }, [individuals, searchQuery, sortBy, sortDirection, page, limit]);

  const handleRowClick = (ind: Individual) => {
    router.push(`/accounts/individuals/${ind.id}`);
  };

  if (loading) {
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
