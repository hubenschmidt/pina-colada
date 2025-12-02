"use client";

import { useState, useEffect, useRef } from "react";
import { TextInput, Paper, Stack, Text } from "@mantine/core";
import { Search, X } from "lucide-react";

export interface SearchSuggestion {
  label: string;
  value: string;
}

interface SearchBoxProps {
  placeholder?: string;
  onSearch: (query: string) => void;
  initialValue?: string;
  fetchPreview?: (query: string) => Promise<SearchSuggestion[]>;
  debounceMs?: number;
}

const SearchBox = ({
  placeholder = "Search... (Enter to search)",
  onSearch,
  initialValue = "",
  fetchPreview,
  debounceMs = 1500,
}: SearchBoxProps) => {
  const [input, setInput] = useState(initialValue);
  const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [lastSearched, setLastSearched] = useState("");
  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!fetchPreview || !input.trim() || input === lastSearched) {
      setSuggestions([]);
      setShowSuggestions(false);
      return;
    }

    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    debounceRef.current = setTimeout(async () => {
      try {
        const results = await fetchPreview(input.trim());
        setSuggestions(results.slice(0, 4));
        setShowSuggestions(results.length > 0);
      } catch {
        setSuggestions([]);
        setShowSuggestions(false);
      }
    }, debounceMs);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [input, fetchPreview, debounceMs, lastSearched]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setShowSuggestions(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSubmit = () => {
    setShowSuggestions(false);
    setLastSearched(input);
    onSearch(input);
  };

  const handleClear = () => {
    setInput("");
    setSuggestions([]);
    setShowSuggestions(false);
    setLastSearched("");
    onSearch("");
  };

  const handleSelectSuggestion = (suggestion: SearchSuggestion) => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }
    setInput(suggestion.value);
    setSuggestions([]);
    setShowSuggestions(false);
    setLastSearched(suggestion.value);
    onSearch(suggestion.value);
  };

  return (
    <div ref={containerRef} style={{ position: "relative", flex: 1 }}>
      <TextInput
        placeholder={placeholder}
        value={input}
        onChange={(e) => {
          const val = e.target.value;
          setInput(val);
          if (val === "") {
            setLastSearched("");
            onSearch("");
          }
        }}
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
      />
      {showSuggestions && suggestions.length > 0 && (
        <Paper
          shadow="md"
          p="xs"
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            right: 0,
            zIndex: 100,
            marginTop: 4,
          }}
        >
          <Stack gap={4}>
            {suggestions.map((s, i) => (
              <Text
                key={i}
                size="sm"
                p="xs"
                style={{ cursor: "pointer", borderRadius: 4 }}
                className="hover:bg-zinc-100 dark:hover:bg-zinc-800"
                onClick={() => handleSelectSuggestion(s)}
              >
                {s.label}
              </Text>
            ))}
          </Stack>
        </Paper>
      )}
    </div>
  );
};

export default SearchBox;
