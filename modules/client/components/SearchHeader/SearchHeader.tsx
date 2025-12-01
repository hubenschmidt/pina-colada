"use client";

import { useState } from "react";
import { Button, Stack, Group, Text } from "@mantine/core";
import { SearchBox } from "../SearchBox";

interface SearchHeaderProps {
  placeholder?: string;
  buttonLabel: string;
  onSearch: (query: string) => void;
  onAdd: () => void;
}

const SearchHeader = ({
  placeholder = "Search... (Enter to search)",
  buttonLabel,
  onSearch,
  onAdd,
}: SearchHeaderProps) => {
  const [searchQuery, setSearchQuery] = useState("");

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    onSearch(query);
  };

  return (
    <Stack gap="xs">
      <Group gap="md">
        <SearchBox placeholder={placeholder} onSearch={handleSearch} />
        <Button onClick={onAdd} color="lime">
          {buttonLabel}
        </Button>
      </Group>
      {searchQuery && (
        <Text size="sm" c="dimmed">
          Showing results for &quot;{searchQuery}&quot;
        </Text>
      )}
    </Stack>
  );
};

export default SearchHeader;
