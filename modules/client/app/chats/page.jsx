"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Stack, Group, Button, Text, Box, Center, Loader } from "@mantine/core";
import { Plus } from "lucide-react";
import SearchBox from "../../components/SearchBox/SearchBox";
import { DataTable } from "../../components/DataTable/DataTable";
import { getAllConversations } from "../../api";
import { usePageLoading } from "../../context/pageLoadingContext";

const formatDate = (dateStr) => {
  if (!dateStr) return "—";
  return new Date(dateStr).toLocaleDateString();
};

const formatDateTime = (dateStr) => {
  if (!dateStr) return "—";
  const d = new Date(dateStr);
  return `${d.toLocaleDateString()} ${d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
};

const ChatsPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();

  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const [searchQuery, setSearchQuery] = useState("");

  const fetchChats = useCallback(async () => {
    setLoading(true);
    try {
      const offset = (page - 1) * pageSize;
      const result = await getAllConversations({
        search: searchQuery || undefined,
        limit: pageSize,
        offset,
      });
      setData({
        items: result.data,
        total: result.total,
        page,
        pageSize,
      });
    } catch (error) {
      console.error("Failed to fetch chats:", error);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, searchQuery]);

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  useEffect(() => {
    fetchChats();
  }, [fetchChats]);

  const handleSearch = (query) => {
    setSearchQuery(query);
    setPage(1);
  };

  const columns = [
    {
      header: "Title",
      accessor: (row) => row.title || "Untitled",
      sortable: false,
    },
    {
      header: "Created",
      accessor: (row) => formatDate(row.created_at),
      sortable: false,
    },
    {
      header: "Updated",
      accessor: (row) => formatDateTime(row.updated_at),
      sortable: false,
    },
    {
      header: "Created By",
      accessor: (row) => row.created_by?.email || "—",
      sortable: false,
    },
    {
      header: "Updated By",
      accessor: (row) => row.updated_by?.email || "—",
      sortable: false,
    },
  ];

  if (loading && !data) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading chats...</Text>
        </Stack>
      </Center>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Chats</h1>
      </Group>

      <Stack gap="xs">
        <Group gap="md">
          <SearchBox placeholder="Search chats... (Enter to search)" onSearch={handleSearch} />
          <Button onClick={() => router.push("/chat")} color="lime" leftSection={<Plus size={16} />}>
            New Chat
          </Button>
        </Group>
        {searchQuery && (
          <Text size="sm" c="dimmed">
            Showing results for &quot;{searchQuery}&quot;
          </Text>
        )}
      </Stack>

      <Box pos="relative">
        <DataTable
          data={data}
          columns={columns}
          rowKey={(row) => row.thread_id}
          pageValue={page}
          onPageChange={setPage}
          pageSizeValue={pageSize}
          onPageSizeChange={(size) => {
            setPageSize(size);
            setPage(1);
          }}
          onRowClick={(chat) => router.push(`/chat/${chat.thread_id}`)}
          emptyText="No chats yet. Start a new conversation!"
        />
      </Box>
    </Stack>
  );
};

export default ChatsPage;
