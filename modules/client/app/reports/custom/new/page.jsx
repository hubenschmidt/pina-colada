"use client";

import { useRouter } from "next/navigation";
import { Stack, Group, Badge } from "@mantine/core";
import { ReportBuilder } from "../../../../components/Reports/ReportBuilder";
import { createSavedReport } from "../../../../api";
import { useProjectContext } from "../../../../context/projectContext";

const NewReportPage = () => {
  const router = useRouter();
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

  const handleSave = async (
  name,
  description,
  query,
  projectIds) =>
  {
    await createSavedReport({
      name,
      description: description || undefined,
      query_definition: query,
      project_ids: projectIds.length > 0 ? projectIds : undefined
    });
    router.push("/reports/custom");
  };

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          New Custom Report
        </h1>
        {selectedProject ?
        <Badge variant="light" color="lime">
            {selectedProject.name}
          </Badge> :

        <Badge variant="light" color="gray">
            Global
          </Badge>
        }
      </Group>
      <ReportBuilder onSave={handleSave} />
    </Stack>);

};

export default NewReportPage;