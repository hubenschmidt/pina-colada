"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Stack, Center, Loader, Text, Group, Badge } from "@mantine/core";
import { ReportBuilder } from "../../../../components/Reports/ReportBuilder";
import { getSavedReport, updateSavedReport } from "../../../../api";
import { useProjectContext } from "../../../../context/projectContext";

const EditReportPage = () => {
  const router = useRouter();
  const params = useParams();
  const reportId = Number(params.id);
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchReport = async () => {
      try {
        const data = await getSavedReport(reportId);
        setReport(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load report");
      } finally {
        setLoading(false);
      }
    };
    fetchReport();
  }, [reportId]);

  const handleSave = async (name, description, query, projectIds) => {
    await updateSavedReport(reportId, {
      name,
      description: description || undefined,
      query_definition: query,
      project_ids: projectIds,
    });
    router.push("/reports/custom");
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading report...</Text>
        </Stack>
      </Center>
    );
  }

  if (error || !report) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Edit Report
        </h1>
        <Text c="red">{error || "Report not found"}</Text>
      </Stack>
    );
  }

  // Only leads entity supports project filtering
  const entitySupportsProjectFilter =
    report.query_definition?.primary_entity === "leads";

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Edit Report: {report.name}
        </h1>
        {entitySupportsProjectFilter && selectedProject ? (
          <Badge variant="light" color="lime">
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>
      <ReportBuilder
        initialQuery={report.query_definition}
        reportName={report.name}
        reportDescription={report.description || ""}
        initialProjectIds={report.project_ids}
        onSave={handleSave}
      />
    </Stack>
  );
};

export default EditReportPage;
