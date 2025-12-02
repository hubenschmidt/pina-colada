"use client";

import { Button, Group } from "@mantine/core";
import { SearchBox, SearchSuggestion } from "../SearchBox";

interface SearchHeaderProps {
  placeholder?: string;
  buttonLabel: string;
  onSearch: (query: string) => void;
  onAdd: () => void;
  fetchPreview?: (query: string) => Promise<SearchSuggestion[]>;
}

const SearchHeader = ({
  placeholder = "Search... (Enter to search)",
  buttonLabel,
  onSearch,
  onAdd,
  fetchPreview,
}: SearchHeaderProps) => {
  return (
    <Group gap="md">
      <SearchBox
        placeholder={placeholder}
        onSearch={onSearch}
        fetchPreview={fetchPreview}
      />
      <Button onClick={onAdd} color="lime">
        {buttonLabel}
      </Button>
    </Group>
  );
};

export default SearchHeader;
