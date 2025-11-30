"use client";

import { useState, useCallback } from "react";
import { Stack, Title, Collapse, Button, Group } from "@mantine/core";
import { ChevronDown, ChevronUp, Upload } from "lucide-react";
import { DocumentUpload, DocumentList } from "../../../components/Documents";
import { Document } from "../../../api";

const DocumentsPage = () => {
  const [uploadOpen, setUploadOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleUploadComplete = useCallback((doc: Document) => {
    setUploadOpen(false);
    setRefreshKey((k) => k + 1);
  }, []);

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Title order={2}>Documents</Title>
        <Button
          variant={uploadOpen ? "light" : "filled"}
          color="lime"
          leftSection={<Upload className="h-4 w-4" />}
          rightSection={
            uploadOpen ? (
              <ChevronUp className="h-4 w-4" />
            ) : (
              <ChevronDown className="h-4 w-4" />
            )
          }
          onClick={() => setUploadOpen(!uploadOpen)}
        >
          Upload
        </Button>
      </Group>

      <Collapse in={uploadOpen}>
        <DocumentUpload onUploadComplete={handleUploadComplete} />
      </Collapse>

      <DocumentList key={refreshKey} />
    </Stack>
  );
};

export default DocumentsPage;
