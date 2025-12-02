"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { DataTable, type PageData } from "../DataTable/DataTable";
import { SearchBox, SearchSuggestion } from "../SearchBox";
import { AccountListConfig, BaseAccount } from "./types/AccountListTypes";
import {
  Stack,
  Center,
  Loader,
  Alert,
  Group,
  Button,
  Box,
  Text,
} from "@mantine/core";

interface AccountListProps<T extends BaseAccount> {
  config: AccountListConfig<T>;
}

const AccountList = <T extends BaseAccount>({ config }: AccountListProps<T>) => {
  const router = useRouter();

  const [data, setData] = useState<PageData<T> | null>(null);
  const [loading, setLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(config.defaultPageSize || 50);
  const [sortBy, setSortBy] = useState<string>(
    config.defaultSortBy || "updated_at"
  );
  const [sortDirection, setSortDirection] = useState<"ASC" | "DESC">(
    config.defaultSortDirection || "DESC"
  );
  const [showLoadingBar, setShowLoadingBar] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const enableSearch = config.enableSearch !== false;

  const loadItems = useCallback(
    (showFullLoading = false, showBar = false, overrideSearch?: string) => {
      if (showFullLoading) {
        setLoading(true);
      }
      if (showBar) {
        setIsRefreshing(true);
        setShowLoadingBar(true);
      }
      setError(null);

      const search = overrideSearch !== undefined ? overrideSearch : searchQuery;
      config.api
        .getItems(page, limit, sortBy, sortDirection, search || undefined)
        .then((pageData) => {
          setData(pageData);
        })
        .catch((err) => {
          console.error(`Error fetching ${config.entityNamePlural}:`, err);
          setError(
            `Failed to load ${config.entityNamePlural.toLowerCase()}. Please try again.`
          );
        })
        .finally(() => {
          setLoading(false);
          setIsRefreshing(false);
          setTimeout(() => setShowLoadingBar(false), 300);
        });
    },
    [page, limit, sortBy, sortDirection, searchQuery, config]
  );

  useEffect(() => {
    loadItems(true);
  }, []);

  useEffect(() => {
    if (data !== null && !loading) {
      const prevPage = data.currentPage || 1;
      const prevLimit = data.pageSize || 50;
      const isPagination = page !== prevPage || limit !== prevLimit;
      loadItems(false, isPagination);
    }
  }, [page, limit, sortBy, sortDirection, searchQuery]);

  const handleRowClick = (item: T) => {
    if (config.detailPagePath) {
      router.push(`${config.detailPagePath}/${item.id}`);
    }
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setPage(1);
    if (query === "") {
      loadItems(false, true, "");
    }
  };

  const fetchPreview = async (query: string): Promise<SearchSuggestion[]> => {
    if (!config.getSuggestionLabel) return [];
    const result = await config.api.getItems(1, 4, sortBy, sortDirection, query);
    return result.items.map((item) => {
      const label = config.getSuggestionLabel!(item);
      const value = config.getSuggestionValue ? config.getSuggestionValue(item) : label;
      return { label, value };
    });
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">
            Loading {config.entityNamePlural.toLowerCase()}...
          </Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Alert color="red" title="Error">
        <Stack gap="md" align="center">
          <Text>{error}</Text>
          <Button color="red" onClick={() => loadItems(true)}>
            Retry
          </Button>
        </Stack>
      </Alert>
    );
  }

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        {config.entityNamePlural}
      </h1>

      {enableSearch && (
        <Group gap="md">
          <SearchBox
            placeholder={
              config.searchPlaceholder ||
              `Search ${config.entityNamePlural.toLowerCase()}... (Enter to search)`
            }
            onSearch={handleSearch}
            fetchPreview={config.getSuggestionLabel ? fetchPreview : undefined}
          />
          <Button
            onClick={() => {
              if (config.newPagePath) {
                router.push(config.newPagePath);
              }
            }}
            color="lime"
          >
            New {config.entityName}
          </Button>
        </Group>
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
          columns={config.columns}
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
          rowKey={(item) => item.id}
          emptyText={
            config.emptyMessage ||
            `No ${config.entityNamePlural.toLowerCase()} yet. Add your first one above!`
          }
        />
      </Box>
    </Stack>
  );
};

export default AccountList;
