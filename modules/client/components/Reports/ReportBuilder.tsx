"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import {
  Stack,
  Card,
  Select,
  MultiSelect,
  Button,
  Group,
  Text,
  Table,
  TextInput,
  ActionIcon,
  Modal,
  Textarea,
  Anchor,
} from "@mantine/core";
import { Plus, Trash2, Play, Download, Save, Globe, FolderKanban } from "lucide-react";
import {
  ReportQueryRequest,
  ReportFilter,
  ReportQueryResult,
  EntityFields,
  getEntityFields,
  previewCustomReport,
  exportCustomReport,
} from "../../api";
import { useProjectContext } from "../../context/projectContext";

const ENTITY_ROUTES: Record<string, string> = {
  organizations: "/accounts/organizations",
  individuals: "/accounts/individuals",
  contacts: "/contacts",
  leads: "/leads",
};

type ReportBuilderProps = {
  initialQuery?: ReportQueryRequest;
  onSave?: (name: string, description: string, query: ReportQueryRequest, projectIds: number[]) => Promise<void>;
  reportName?: string;
  reportDescription?: string;
  initialProjectIds?: number[];
};

const ENTITIES = [
  { value: "organizations", label: "Organizations" },
  { value: "individuals", label: "Individuals" },
  { value: "contacts", label: "Contacts" },
  { value: "leads", label: "Leads" },
  { value: "notes", label: "Notes" },
];

const OPERATORS = [
  { value: "eq", label: "Equals" },
  { value: "neq", label: "Not Equals" },
  { value: "contains", label: "Contains" },
  { value: "starts_with", label: "Starts With" },
  { value: "gt", label: "Greater Than" },
  { value: "gte", label: "Greater or Equal" },
  { value: "lt", label: "Less Than" },
  { value: "lte", label: "Less or Equal" },
  { value: "is_null", label: "Is Empty" },
  { value: "is_not_null", label: "Is Not Empty" },
];

export const ReportBuilder = ({
  initialQuery,
  onSave,
  reportName: initialName = "",
  reportDescription: initialDescription = "",
  initialProjectIds = [],
}: ReportBuilderProps) => {
  const [entity, setEntity] = useState<string>(initialQuery?.primary_entity || "");
  const [availableFields, setAvailableFields] = useState<EntityFields | null>(null);
  const [selectedColumns, setSelectedColumns] = useState<string[]>(initialQuery?.columns || []);
  const [selectedJoins, setSelectedJoins] = useState<string[]>(initialQuery?.joins || []);
  const [filters, setFilters] = useState<ReportFilter[]>(initialQuery?.filters || []);
  const [preview, setPreview] = useState<ReportQueryResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saveModalOpen, setSaveModalOpen] = useState(false);
  const [saveName, setSaveName] = useState(initialName);
  const [saveDescription, setSaveDescription] = useState(initialDescription);
  const [saving, setSaving] = useState(false);
  const [selectedProjectIds, setSelectedProjectIds] = useState<string[]>(
    initialProjectIds.map(String)
  );
  const { projectState } = useProjectContext();
  const { projects, selectedProject } = projectState;

  useEffect(() => {
    if (!entity) {
      setAvailableFields(null);
      return;
    }
    const fetchFields = async () => {
      try {
        const fields = await getEntityFields(entity);
        setAvailableFields(fields);
      } catch (err) {
        console.error("Failed to fetch fields:", err);
      }
    };
    fetchFields();
  }, [entity]);

  // Re-run preview when selected project changes (only for entities that support project filtering)
  useEffect(() => {
    if (preview && entity === "leads" && selectedColumns.length > 0) {
      runPreview();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProject?.id]);

  const allFields = availableFields
    ? [...availableFields.base, ...availableFields.joins]
    : [];

  const fieldOptions = allFields.map((f) => ({ value: f, label: f }));

  const addFilter = () => {
    setFilters([...filters, { field: "", operator: "eq", value: "" }]);
  };

  const updateFilter = (index: number, updates: Partial<ReportFilter>) => {
    setFilters((prev) =>
      prev.map((f, i) => (i === index ? { ...f, ...updates } : f))
    );
  };

  const removeFilter = (index: number) => {
    setFilters((prev) => prev.filter((_, i) => i !== index));
  };

  // Only leads entity supports project filtering
  const entitySupportsProjectFilter = entity === "leads";

  const buildQuery = (): ReportQueryRequest => ({
    primary_entity: entity as ReportQueryRequest["primary_entity"],
    columns: selectedColumns,
    joins: selectedJoins.length > 0 ? selectedJoins : undefined,
    filters: filters.filter((f) => f.field && f.operator),
    limit: 100,
    offset: 0,
    project_id: entitySupportsProjectFilter ? selectedProject?.id : undefined,
  });

  const runPreview = async () => {
    if (!entity || selectedColumns.length === 0) {
      setError("Please select an entity and at least one column");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const result = await previewCustomReport(buildQuery());
      setPreview(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to run preview");
    } finally {
      setLoading(false);
    }
  };

  const handleExport = async () => {
    if (!entity || selectedColumns.length === 0) {
      setError("Please select an entity and at least one column");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const blob = await exportCustomReport(buildQuery());
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "report.xlsx";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to export");
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!saveName.trim()) {
      return;
    }
    if (!onSave) {
      return;
    }
    setSaving(true);
    try {
      const projectIds = selectedProjectIds.map(Number);
      await onSave(saveName, saveDescription, buildQuery(), projectIds);
      setSaveModalOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save report");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Stack gap="lg">
      {/* Entity Selection */}
      <Card shadow="sm" padding="lg" radius="md" withBorder>
        <Text fw={600} mb="md">1. Select Entity</Text>
        <Select
          label="Primary Entity"
          placeholder="Choose an entity to report on"
          data={ENTITIES}
          value={entity}
          onChange={(val) => {
            setEntity(val || "");
            setSelectedColumns([]);
            setSelectedJoins([]);
            setFilters([]);
            setPreview(null);
          }}
        />
      </Card>

      {/* Join Selection */}
      {entity && availableFields && availableFields.available_joins.length > 0 && (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">2. Link Related Entities (Optional)</Text>
          <MultiSelect
            label="Join with"
            placeholder="Select entities to join"
            data={availableFields.available_joins.map((j) => ({
              value: j.name,
              label: j.name.charAt(0).toUpperCase() + j.name.slice(1),
            }))}
            value={selectedJoins}
            onChange={setSelectedJoins}
          />
          {selectedJoins.length > 0 && (
            <Text c="dimmed" size="xs" mt="xs">
              Fields from joined entities will be available in column selection
            </Text>
          )}
        </Card>
      )}

      {/* Column Selection */}
      {entity && availableFields && (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            {availableFields.available_joins.length > 0 ? "3" : "2"}. Select Columns
          </Text>
          <MultiSelect
            label="Columns to include"
            placeholder="Select columns"
            data={fieldOptions}
            value={selectedColumns}
            onChange={setSelectedColumns}
            searchable
          />
        </Card>
      )}

      {/* Filters */}
      {entity && availableFields && (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Group justify="space-between" mb="md">
            <Text fw={600}>
              {availableFields.available_joins.length > 0 ? "4" : "3"}. Filters (Optional)
            </Text>
            <Button
              variant="light"
              size="xs"
              leftSection={<Plus className="h-3 w-3" />}
              onClick={addFilter}
            >
              Add Filter
            </Button>
          </Group>

          {filters.length === 0 ? (
            <Text c="dimmed" size="sm">No filters applied</Text>
          ) : (
            <Stack gap="sm">
              {filters.map((filter, index) => (
                <Group key={index} gap="sm">
                  <Select
                    placeholder="Field"
                    data={fieldOptions}
                    value={filter.field}
                    onChange={(val) => updateFilter(index, { field: val || "" })}
                    style={{ flex: 1 }}
                  />
                  <Select
                    placeholder="Operator"
                    data={OPERATORS}
                    value={filter.operator}
                    onChange={(val) =>
                      updateFilter(index, {
                        operator: (val || "eq") as ReportFilter["operator"],
                      })
                    }
                    style={{ width: 150 }}
                  />
                  <TextInput
                    placeholder="Value"
                    value={filter.value || ""}
                    onChange={(e) => updateFilter(index, { value: e.target.value })}
                    style={{ flex: 1 }}
                    disabled={filter.operator === "is_null" || filter.operator === "is_not_null"}
                  />
                  <ActionIcon variant="subtle" color="red" onClick={() => removeFilter(index)}>
                    <Trash2 className="h-4 w-4" />
                  </ActionIcon>
                </Group>
              ))}
            </Stack>
          )}
        </Card>
      )}

      {/* Actions */}
      {entity && selectedColumns.length > 0 && (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Group>
            <Button
              leftSection={<Play className="h-4 w-4" />}
              color="lime"
              onClick={runPreview}
              loading={loading}
            >
              Preview
            </Button>
            <Button
              leftSection={<Download className="h-4 w-4" />}
              variant="light"
              onClick={handleExport}
              loading={loading}
            >
              Export to Excel
            </Button>
            {onSave && (
              <Button
                leftSection={<Save className="h-4 w-4" />}
                variant="outline"
                onClick={() => setSaveModalOpen(true)}
              >
                Save Report
              </Button>
            )}
          </Group>
        </Card>
      )}

      {/* Error */}
      {error && (
        <Text c="red" size="sm">{error}</Text>
      )}

      {/* Preview Results */}
      {preview && (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Text fw={600} mb="md">
            Preview Results ({preview.total} total rows, showing {preview.data.length})
          </Text>

          {preview.data.length === 0 ? (
            <Text c="dimmed" size="sm">No data found</Text>
          ) : (
            <div style={{ overflowX: "auto" }}>
              <Table>
                <Table.Thead>
                  <Table.Tr>
                    {selectedColumns.map((col) => (
                      <Table.Th key={col}>{col}</Table.Th>
                    ))}
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {preview.data.map((row, idx) => {
                    const entityRoute = ENTITY_ROUTES[entity];
                    const entityId = row["id"];
                    const entityLink = entityRoute && entityId ? `${entityRoute}/${entityId}` : null;

                    return (
                      <Table.Tr key={idx}>
                        {selectedColumns.map((col) => {
                          const value = row[col];
                          const displayValue = value !== null && value !== undefined ? String(value) : "-";

                          // Make the ID column a clickable link
                          if (col === "id" && entityLink) {
                            return (
                              <Table.Td key={col}>
                                <Anchor component={Link} href={entityLink} size="sm" c="blue">
                                  {displayValue}
                                </Anchor>
                              </Table.Td>
                            );
                          }

                          return (
                            <Table.Td key={col}>
                              {displayValue}
                            </Table.Td>
                          );
                        })}
                      </Table.Tr>
                    );
                  })}
                </Table.Tbody>
              </Table>
            </div>
          )}
        </Card>
      )}

      {/* Save Modal */}
      <Modal
        opened={saveModalOpen}
        onClose={() => setSaveModalOpen(false)}
        title="Save Report"
      >
        <Stack gap="md">
          <TextInput
            label="Report Name"
            placeholder="Enter a name for this report"
            value={saveName}
            onChange={(e) => setSaveName(e.target.value)}
            required
          />
          <Textarea
            label="Description (optional)"
            placeholder="Describe what this report shows"
            value={saveDescription}
            onChange={(e) => setSaveDescription(e.target.value)}
          />
          <div>
            <Text size="sm" fw={500} mb="xs">Report Scope</Text>
            <MultiSelect
              placeholder="Select projects (leave empty for global)"
              data={projects
                .filter((p) => p.status !== "Inactive")
                .map((p) => ({ value: p.id.toString(), label: p.name }))}
              value={selectedProjectIds}
              onChange={setSelectedProjectIds}
              clearable
              searchable
              leftSection={
                selectedProjectIds.length === 0 ? (
                  <Globe className="h-4 w-4 text-gray-400" />
                ) : (
                  <FolderKanban className="h-4 w-4 text-lime-500" />
                )
              }
            />
            <Text size="xs" c="dimmed" mt="xs">
              {selectedProjectIds.length === 0
                ? "Global reports are visible regardless of project selection"
                : `This report will be visible when any of the selected projects is active`}
            </Text>
          </div>
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setSaveModalOpen(false)}>
              Cancel
            </Button>
            <Button
              color="lime"
              onClick={handleSave}
              loading={saving}
              disabled={!saveName.trim()}
            >
              Save
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  );
};
