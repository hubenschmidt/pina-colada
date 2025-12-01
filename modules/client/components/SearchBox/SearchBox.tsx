"use client";

import { useState } from "react";
import { TextInput } from "@mantine/core";
import { Search, X } from "lucide-react";

interface SearchBoxProps {
  placeholder?: string;
  onSearch: (query: string) => void;
  initialValue?: string;
}

const SearchBox = ({
  placeholder = "Search... (Enter to search)",
  onSearch,
  initialValue = "",
}: SearchBoxProps) => {
  const [input, setInput] = useState(initialValue);

  const handleSubmit = () => {
    onSearch(input);
  };

  const handleClear = () => {
    setInput("");
    onSearch("");
  };

  return (
    <TextInput
      placeholder={placeholder}
      value={input}
      onChange={(e) => setInput(e.target.value)}
      onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
      leftSection={<Search size={20} />}
      rightSection={
        input && (
          <button
            onClick={handleClear}
            className="text-zinc-400 hover:text-zinc-600 dark:text-zinc-500 dark:hover:text-zinc-400"
            aria-label="Clear search"
          >
            <X size={18} />
          </button>
        )
      }
      style={{ flex: 1 }}
    />
  );
};

export default SearchBox;
