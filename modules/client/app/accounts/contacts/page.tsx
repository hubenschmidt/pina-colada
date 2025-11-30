"use client";

import { useEffect, useState, useMemo } from "react";
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
    sortable: true,
    sortKey: "account",
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
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(50);
  const [sortBy, setSortBy] = useState("updated_at");
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">("DESC");

  useEffect(() => {
    const fetchContacts = async () => {
      try {
        const data = await getContacts();
        setContacts(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load contacts");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchContacts();
  }, [dispatchPageLoading]);

  const filteredAndSortedData = useMemo((): PageData<Contact> => {
    let filtered = contacts;

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = contacts.filter(
        (c) =>
          c.first_name?.toLowerCase().includes(query) ||
          c.last_name?.toLowerCase().includes(query) ||
          c.email?.toLowerCase().includes(query) ||
          c.title?.toLowerCase().includes(query)
      );
    }

    const sorted = [...filtered].sort((a, b) => {
      let comparison: number;

      if (sortBy === "updated_at") {
        const aDate = new Date(a.updated_at || 0).getTime();
        const bDate = new Date(b.updated_at || 0).getTime();
        comparison = aDate - bDate;
      } else if (sortBy === "account") {
        const aVal = (a.organizations?.map((o) => o.name).join(", ") || "").toLowerCase();
        const bVal = (b.organizations?.map((o) => o.name).join(", ") || "").toLowerCase();
        comparison = aVal.localeCompare(bVal);
      } else {
        const aVal = (a[sortBy as keyof Contact] as string || "").toLowerCase();
        const bVal = (b[sortBy as keyof Contact] as string || "").toLowerCase();
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
  }, [contacts, searchQuery, sortBy, sortDirection, page, limit]);

  const handleRowClick = (c: Contact) => {
    router.push(`/accounts/contacts/${c.id}`);
  };

  if (loading) {
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
