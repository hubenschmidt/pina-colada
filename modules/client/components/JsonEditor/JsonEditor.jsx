"use client";

import { useState, useEffect, useCallback } from "react";
import { Textarea, Alert, Stack, Text, Group, ActionIcon, Tooltip, Box } from "@mantine/core";
import { Copy, Check, RotateCcw, Pencil } from "lucide-react";

const URL_REGEX = /(https?:\/\/[^\s"',}\]]+)/g;

const renderTextWithLinks = (text) => {
  const parts = text.split(URL_REGEX);
  return parts.map((part, i) => {
    if (URL_REGEX.test(part)) {
      URL_REGEX.lastIndex = 0; // Reset regex state
      return (
        <a
          key={i}
          href={part}
          target="_blank"
          rel="noopener noreferrer"
          style={{ color: "var(--mantine-color-lime-5)", textDecoration: "underline" }}
          onClick={(e) => e.stopPropagation()}
        >
          {part}
        </a>
      );
    }
    return part;
  });
};

const JsonEditor = ({ value, onChange, readOnly = false, minRows = 10, error }) => {
  const [text, setText] = useState("");
  const [parseError, setParseError] = useState(null);
  const [copied, setCopied] = useState(false);
  const [editing, setEditing] = useState(false);

  // Serialize value to text on mount or when value changes from outside
  useEffect(() => {
    if (value === undefined || value === null) {
      setText("");
      return;
    }
    try {
      const formatted = typeof value === "string" ? value : JSON.stringify(value, null, 2);
      setText(formatted);
      setParseError(null);
    } catch (e) {
      setText(String(value));
      setParseError("Invalid JSON value");
    }
  }, [value]);

  const handleChange = useCallback(
    (newText) => {
      setText(newText);

      if (!newText.trim()) {
        setParseError(null);
        onChange?.(null);
        return;
      }

      try {
        const parsed = JSON.parse(newText);
        setParseError(null);
        onChange?.(parsed);
      } catch (e) {
        setParseError(e.message);
      }
    },
    [onChange]
  );

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (e) {
      console.error("Failed to copy:", e);
    }
  }, [text]);

  const handleFormat = useCallback(() => {
    if (!text.trim()) return;
    try {
      const parsed = JSON.parse(text);
      const formatted = JSON.stringify(parsed, null, 2);
      setText(formatted);
      setParseError(null);
    } catch (e) {
      // Can't format invalid JSON
    }
  }, [text]);

  const showRenderedView = !editing && !readOnly && text.includes("http");

  return (
    <Stack gap="xs">
      <Group justify="space-between" align="center">
        <Text size="sm" c="dimmed">
          JSON Payload
        </Text>
        <Group gap="xs">
          {!readOnly && showRenderedView && (
            <Tooltip label="Edit">
              <ActionIcon variant="subtle" size="sm" onClick={() => setEditing(true)}>
                <Pencil size={14} />
              </ActionIcon>
            </Tooltip>
          )}
          <Tooltip label="Format JSON">
            <ActionIcon variant="subtle" size="sm" onClick={handleFormat} disabled={!!parseError}>
              <RotateCcw size={14} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label={copied ? "Copied!" : "Copy"}>
            <ActionIcon variant="subtle" size="sm" onClick={handleCopy}>
              {copied ? <Check size={14} /> : <Copy size={14} />}
            </ActionIcon>
          </Tooltip>
        </Group>
      </Group>

      {showRenderedView ? (
        <Box
          onClick={() => setEditing(true)}
          style={{
            fontFamily: "monospace",
            fontSize: "13px",
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            padding: "10px",
            borderRadius: "4px",
            border: "1px solid var(--mantine-color-dark-4)",
            backgroundColor: "var(--mantine-color-dark-6)",
            minHeight: `${minRows * 1.5}em`,
            cursor: "text",
          }}
        >
          {renderTextWithLinks(text)}
        </Box>
      ) : (
        <Textarea
          value={text}
          onChange={(e) => handleChange(e.currentTarget.value)}
          onBlur={() => setEditing(false)}
          readOnly={readOnly}
          minRows={minRows}
          autosize
          autoFocus={editing}
          styles={{
            input: {
              fontFamily: "monospace",
              fontSize: "13px",
            },
          }}
          error={parseError || error}
        />
      )}

      {parseError && (
        <Alert color="red" variant="light" p="xs">
          <Text size="xs">{parseError}</Text>
        </Alert>
      )}
    </Stack>
  );
};

export default JsonEditor;
