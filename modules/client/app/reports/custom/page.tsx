"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Stack, Center, Loader, Text, Card, Button, Group, Table, ActionIcon, Anchor, Badge } from "@mantine/core";
import { Plus, Trash2 } from "lucide-react";
import { getSavedReports, deleteSavedReport, SavedReport } from "../../../api";
import { useProjectContext } from "../../../context/projectContext";

const CustomReportsPage = () => {
  const router = useRouter();
  const [reports, setReports] = useState<SavedReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

  const fetchReports = async () => {
    try {
      const data = await getSavedReports(selectedProject?.id, true);
      setReports(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load saved reports");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setLoading(true);
    fetchReports();
  }, [selectedProject?.id]);

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure you want to delete this report?")) {
      return;
    }
    try {
      await deleteSavedReport(id);
      setReports((prev) => prev.filter((r) => r.id !== id));
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to delete report");
    }
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading saved reports...</Text>
        </Stack>
      </Center>
    );
  }

  if (error) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Custom Reports
        </h1>
        <Text c="red">{error}</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Custom Reports
        </h1>
        <Button
          leftSection={<Plus className="h-4 w-4" />}
          color="lime"
          onClick={() => router.push("/reports/custom/new")}
        >
          New Report
        </Button>
      </Group>

      {reports.length === 0 ? (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack align="center" gap="md" py="xl">
            <Text c="dimmed">No saved reports yet</Text>
            <Button
              variant="light"
              color="lime"
              onClick={() => router.push("/reports/custom/new")}
            >
              Create your first report
            </Button>
          </Stack>
        </Card>
      ) : (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>Scope</Table.Th>
                <Table.Th>Description</Table.Th>
                <Table.Th>Entity</Table.Th>
                <Table.Th>Created</Table.Th>
                <Table.Th></Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {reports.map((report) => (
                <Table.Tr key={report.id}>
                  <Table.Td>
                    <Anchor component={Link} href={`/reports/custom/${report.id}`} fw={500} c="blue">
                      {report.name}
                    </Anchor>
                  </Table.Td>
                  <Table.Td>
                    {report.query_definition?.primary_entity === "leads" && selectedProject ? (
                      <Badge variant="light" color="lime">
                        {selectedProject.name}
                      </Badge>
                    ) : (
                      <Badge variant="light" color="gray">
                        Global
                      </Badge>
                    )}
                  </Table.Td>
                  <Table.Td>
                    <Text c="dimmed" size="sm" lineClamp={1}>
                      {report.description || "-"}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm" tt="capitalize">
                      {report.query_definition?.primary_entity || "-"}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm" c="dimmed">
                      {report.created_at
                        ? new Date(report.created_at).toLocaleDateString()
                        : "-"}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <ActionIcon
                      variant="subtle"
                      color="red"
                      onClick={() => handleDelete(report.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </ActionIcon>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Card>
      )}
    </Stack>
  );
};

export default CustomReportsPage;
