"use client";

import { useState } from "react";
import { TextInput, Button, Stack, Group, Text } from "@mantine/core";
import { Search, X } from "lucide-react";

interface SearchHeaderProps {
  placeholder?: string;
  buttonLabel: string;
  onSearch: (query: string) => void;
  onAdd: () => void;
  debounceMs?: number;
}

const SearchHeader = ({
  placeholder = "Search...",
  buttonLabel,
  onSearch,
  onAdd,
  debounceMs = 300,
}: SearchHeaderProps) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [debounceTimer, setDebounceTimer] = useState<NodeJS.Timeout | null>(null);

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);

    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    const timer = setTimeout(() => {
      onSearch(value);
    }, debounceMs);

    setDebounceTimer(timer);
  };

  const handleClearSearch = () => {
    setSearchQuery("");
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }
    onSearch("");
  };

  return (
    <Stack gap="xs">
      <Group gap="md">
        <TextInput
          placeholder={placeholder}
          value={searchQuery}
          onChange={(e) => handleSearchChange(e.target.value)}
          leftSection={<Search size={20} />}
          rightSection={
            searchQuery && (
              <button
                onClick={handleClearSearch}
                className="text-zinc-400 hover:text-zinc-600 dark:text-zinc-500 dark:hover:text-zinc-400"
                aria-label="Clear search"
              >
                <X size={18} />
              </button>
            )
          }
          style={{ flex: 1 }}
          styles={{
            input: {
              transition: "background-color 0.2s ease",
              "&:hover": {
                backgroundColor: "var(--input-background)",
                filter: "brightness(0.97)",
              },
            },
          }}
        />
        <Button onClick={onAdd} variant="default">
          {buttonLabel}
        </Button>
      </Group>
      {searchQuery && (
        <Text size="sm" c="dimmed">
          Showing results for "{searchQuery}"
        </Text>
      )}
    </Stack>
  );
};

export default SearchHeader;
