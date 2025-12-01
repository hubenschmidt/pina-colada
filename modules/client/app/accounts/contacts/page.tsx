"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { getContacts, Contact } from "../../../api";
import { Stack, Center, Loader, Text } from "@mantine/core";
import SearchHeader from "../../../components/SearchHeader/SearchHeader";
import { DataTable, Column, PageData } from "../../../components/DataTable/DataTable";

const columns: Column<Contact>[] = [
  {
    header: "Last Name",
    accessor: "last_name",
    sortable: true,
    sortKey: "last_name",
    render: (c) => c.last_name || "-",
  },
  {
    header: "First Name",
    accessor: "first_name",
    sortable: true,
    sortKey: "first_name",
    render: (c) => c.first_name || "-",
  },
  {
    header: "Title",
    accessor: "title",
    sortable: true,
    sortKey: "title",
    render: (c) => c.title || "-",
  },
  {
    header: "Account",
    sortable: false,
    render: (c) =>
      c.organizations && c.organizations.length > 0
        ? c.organizations.map((o) => o.name).join(", ")
        : "-",
  },
  {
    header: "Email",
    accessor: "email",
    sortable: true,
    sortKey: "email",
    render: (c) =>
      c.email ? (
        <a
          href={`mailto:${c.email}`}
          className="text-blue-600 dark:text-blue-400 hover:underline"
          onClick={(e) => e.stopPropagation()}
        >
          {c.email}
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
    render: (c) => c.phone || "-",
  },
];

const ContactsPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const [data, setData] = useState<PageData<Contact>>({
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

  const fetchContacts = useCallback(async () => {
    try {
      setLoading(true);
      const result = await getContacts(
        page,
        limit,
        sortBy,
        sortDirection,
        searchQuery || undefined
      );
      setData(result);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load contacts");
    } finally {
      setLoading(false);
      dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
    }
  }, [page, limit, sortBy, sortDirection, searchQuery, dispatchPageLoading]);

  useEffect(() => {
    fetchContacts();
  }, [fetchContacts]);

  const handleRowClick = (c: Contact) => {
    router.push(`/accounts/contacts/${c.id}`);
  };

  if (loading && data.items.length === 0) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading contacts...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Contacts
        </h1>
        <p className="text-red-600 dark:text-red-400">{error}</p>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Contacts
      </h1>

      <SearchHeader
        placeholder="Search contacts..."
        buttonLabel="New Contact"
        onSearch={(query) => {
          setSearchQuery(query);
          setPage(1);
        }}
        onAdd={() => router.push("/accounts/contacts/new")}
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
        rowKey={(c) => c.id}
        emptyText={
          searchQuery
            ? "No matching contacts found."
            : "No contacts yet. Add your first one above!"
        }
      />
    </Stack>
  );
};

export default ContactsPage;
