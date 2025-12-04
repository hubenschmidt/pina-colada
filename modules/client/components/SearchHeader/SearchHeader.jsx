"use client";

import { Button, Group } from "@mantine/core";
import { SearchBox } from "../SearchBox";

const SearchHeader = ({
  placeholder = "Search... (Enter to search)",
  buttonLabel,
  onSearch,
  onAdd,
  fetchPreview,
}) => {
  return (
    <Group gap="md">
      <SearchBox placeholder={placeholder} onSearch={onSearch} fetchPreview={fetchPreview} />

      <Button onClick={onAdd} color="lime">
        {buttonLabel}
      </Button>
    </Group>
  );
};

export default SearchHeader;
