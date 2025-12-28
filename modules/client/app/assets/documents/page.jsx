"use client";

import { useState, useCallback, useEffect } from "react";
import { Stack, Collapse, Button } from "@mantine/core";
import { ChevronDown, ChevronUp, Upload } from "lucide-react";
import { DocumentUpload, DocumentList } from "../../../components/Documents";

import { usePageLoading } from "../../../context/pageLoadingContext";

const DocumentsPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const [uploadOpen, setUploadOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const handleUploadComplete = useCallback((doc) => {
    setUploadOpen(false);
    setRefreshKey((k) => k + 1);
  }, []);

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">Documents</h1>

      {uploadOpen && (
        <Stack gap="xs">
          <Collapse in={uploadOpen}>
            <DocumentUpload onUploadComplete={handleUploadComplete} />
          </Collapse>
        </Stack>
      )}

      <DocumentList
        key={refreshKey}
        headerRight={
          <Button
            variant={uploadOpen ? "light" : "filled"}
            color="lime"
            leftSection={<Upload className="h-4 w-4" />}
            rightSection={
              uploadOpen ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />
            }
            onClick={() => setUploadOpen(!uploadOpen)}>
            Upload
          </Button>
        }
      />
    </Stack>
  );
};

export default DocumentsPage;
