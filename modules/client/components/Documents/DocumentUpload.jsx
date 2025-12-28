"use client";

import { useState, useCallback } from "react";
import { useDropzone } from "react-dropzone";
import {
  Stack,
  Text,
  Progress,
  TagsInput,
  Textarea,
  Button,
  Group,
  Paper,
  ActionIcon,
  Modal,
} from "@mantine/core";
import { Upload, X, FileText, AlertTriangle } from "lucide-react";
import { uploadDocument, getTags, checkDocumentFilename, createDocumentVersion } from "../../api";

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
const ACCEPTED_TYPES = {
  "application/pdf": [".pdf"],
  "text/plain": [".txt"],
  "text/markdown": [".md"],
  "application/msword": [".doc"],
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [".docx"],
};

export const DocumentUpload = ({ onUploadComplete, entityType, entityId }) => {
  const [file, setFile] = useState(null);
  const [tags, setTags] = useState([]);
  const [availableTags, setAvailableTags] = useState([]);
  const [description, setDescription] = useState("");
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState(null);
  const [versionModalOpen, setVersionModalOpen] = useState(false);
  const [existingDocument, setExistingDocument] = useState(null);

  const loadTags = useCallback(async () => {
    try {
      const tagList = await getTags();
      setAvailableTags(tagList.map((t) => t.name));
    } catch {
      // Tags are optional, ignore errors
    }
  }, []);

  const onDrop = useCallback(
    (acceptedFiles) => {
      setError(null);
      const selected = acceptedFiles[0];
      if (!selected) return;

      if (selected.size > MAX_FILE_SIZE) {
        setError("File exceeds 10MB limit");
        return;
      }
      setFile(selected);
      loadTags();
    },
    [loadTags]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: ACCEPTED_TYPES,
    maxFiles: 1,
    multiple: false,
  });

  const handleUpload = async () => {
    if (!file) return;

    setError(null);

    // Check for existing filename if entity context is provided
    if (entityType && entityId) {
      try {
        const { exists, document: existing } = await checkDocumentFilename(
          file.name,
          entityType,
          entityId
        );
        if (exists && existing) {
          setExistingDocument(existing);
          setVersionModalOpen(true);
          return;
        }
      } catch {
        // If check fails, proceed with upload anyway
      }
    }

    await performUpload();
  };

  const performUpload = async (asVersion = false) => {
    if (!file) return;

    setUploading(true);
    setProgress(0);
    setError(null);
    setVersionModalOpen(false);

    try {
      // Simulate progress since axios doesn't expose it easily for FormData
      const progressInterval = setInterval(() => {
        setProgress((prev) => Math.min(prev + 10, 90));
      }, 100);

      const document =
        asVersion && existingDocument
          ? await createDocumentVersion(existingDocument.id, file)
          : await uploadDocument(
              file,
              tags.length > 0 ? tags : undefined,
              entityType,
              entityId,
              description || undefined
            );

      clearInterval(progressInterval);
      setProgress(100);

      // Reset form
      setFile(null);
      setTags([]);
      setDescription("");
      setExistingDocument(null);

      onUploadComplete?.(document);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setUploading(false);
      setProgress(0);
    }
  };

  const handleClear = () => {
    setFile(null);
    setTags([]);
    setDescription("");
    setError(null);
  };

  const formatFileSize = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <>
      {/* Version Confirmation Modal */}
      <Modal
        opened={versionModalOpen}
        onClose={() => setVersionModalOpen(false)}
        title={
          <Group gap="xs">
            <AlertTriangle className="h-5 w-5 text-yellow-500" />
            <Text fw={600}>File Already Exists</Text>
          </Group>
        }
        centered>
        <Stack gap="md">
          <Text size="sm">
            A file named &quot;{file?.name}&quot; already exists on this entity (currently at
            version {existingDocument?.version_number}).
          </Text>
          <Text size="sm" c="dimmed">
            Would you like to create a new version of this document?
          </Text>
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => {
                setVersionModalOpen(false);
                setExistingDocument(null);
              }}>
              Cancel
            </Button>
            <Button color="lime" onClick={() => performUpload(true)}>
              Create New Version
            </Button>
          </Group>
        </Stack>
      </Modal>

      <Stack gap="md">
        {!file ? (
          <Paper
            {...getRootProps()}
            p="xl"
            withBorder
            style={{
              borderStyle: "dashed",
              cursor: "pointer",
              backgroundColor: isDragActive ? "var(--mantine-color-lime-0)" : undefined,
            }}>
            <input {...getInputProps()} />
            <Stack align="center" gap="sm">
              <Upload className={`h-8 w-8 ${isDragActive ? "text-lime-600" : "text-zinc-400"}`} />

              <Text size="sm" c="dimmed" ta="center">
                {isDragActive ? "Drop file here..." : "Drag & drop a file, or click to select"}
              </Text>
              <Text size="xs" c="dimmed">
                PDF, TXT, MD, DOC, DOCX (max 10MB)
              </Text>
            </Stack>
          </Paper>
        ) : (
          <Paper p="md" withBorder>
            <Group justify="space-between" mb="sm">
              <Group gap="sm">
                <FileText className="h-5 w-5 text-lime-600" />
                <div>
                  <Text size="sm" fw={500}>
                    {file.name}
                  </Text>
                  <Text size="xs" c="dimmed">
                    {formatFileSize(file.size)}
                  </Text>
                </div>
              </Group>
              <ActionIcon variant="subtle" color="gray" onClick={handleClear}>
                <X className="h-4 w-4" />
              </ActionIcon>
            </Group>

            <Stack gap="sm">
              <TagsInput
                label="Tags"
                placeholder="Add tags..."
                value={tags}
                onChange={setTags}
                data={availableTags}
                clearable
                size="sm"
              />

              <Textarea
                label="Description"
                placeholder="Optional description..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                size="sm"
                minRows={2}
              />

              {uploading && <Progress value={progress} size="sm" color="lime" />}

              {error && (
                <Text size="sm" c="red">
                  {error}
                </Text>
              )}

              <Group justify="flex-end">
                <Button variant="subtle" onClick={handleClear} disabled={uploading}>
                  Cancel
                </Button>
                <Button
                  onClick={handleUpload}
                  loading={uploading}
                  color="lime"
                  leftSection={<Upload className="h-4 w-4" />}>
                  Upload
                </Button>
              </Group>
            </Stack>
          </Paper>
        )}
      </Stack>
    </>
  );
};
