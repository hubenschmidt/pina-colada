"use client";

import { useEffect, useState } from "react";
import {
  Container,
  Stack,
  Title,
  Text,
  Switch,
  NumberInput,
  Textarea,
  TextInput,
  Button,
  Paper,
  Group,
  Badge,
  Table,
  Select,
  MultiSelect,
  TagsInput,
  Loader,
  Alert,
  Modal,
  ActionIcon,
  Menu,
  Collapse,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import {
  Play,
  Clock,
  Mail,
  Search,
  AlertCircle,
  CheckCircle,
  Plus,
  MoreVertical,
  Edit,
  Trash,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
import { usePageLoading } from "../../context/pageLoadingContext";
import { useUserContext } from "../../context/userContext";
import {
  getCrawlers,
  createCrawler,
  updateCrawler,
  deleteCrawler,
  toggleCrawler,
  getCrawlerRuns,
  testCrawler,
  searchIndividuals,
  searchOrganizations,
  searchContacts,
  getIndividual,
  getOrganization,
  getJob,
  getContact,
  getDocuments,
} from "../../api";

const ENTITY_TYPES = [
  { value: "job", label: "Jobs" },
  { value: "opportunity", label: "Opportunities" },
  { value: "partnership", label: "Partnerships" },
  { value: "individual", label: "Individuals" },
  { value: "contact", label: "Contacts" },
];

// Target type options based on primary entity type
const TARGET_TYPE_OPTIONS = {
  job: [{ value: "individual", label: "Individuals (match to profile)" }],
  opportunity: [{ value: "organization", label: "Organizations" }],
  partnership: [{ value: "organization", label: "Organizations" }],
  individual: [
    { value: "organization", label: "Organizations" },
    { value: "job", label: "Jobs" },
    { value: "contact", label: "Contacts" },
  ],
  contact: [
    { value: "organization", label: "Organizations" },
    { value: "job", label: "Jobs" },
    { value: "individual", label: "Individuals" },
  ],
};

const AGENT_MODEL_OPTIONS = [
  { value: "claude-sonnet-4-20250514", label: "Claude Sonnet 4 (Recommended)" },
  { value: "claude-haiku-4-20250514", label: "Claude Haiku 4 (Faster)" },
  { value: "gpt-5.2", label: "GPT 5.2" },
  { value: "gpt-5.1", label: "GPT 5.1" },
];

const emptyForm = {
  name: "",
  entity_type: "job",
  interval_minutes: 30,
  leads_per_run: 10,
  concurrent_searches: 1,
  compilation_target: 100,
  search_keywords: [],
  search_slots: [],
  ats_mode: true,
  time_filter: "week",
  target_type: null,
  target_ids: [],
  source_document_ids: [],
  system_prompt: "",
  digest_enabled: true,
  digest_emails: "",
  digest_time: "09:00",
  digest_model: null,
  use_agent: false,
  agent_model: "claude-sonnet-4-20250514",
};

const AutomationPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const { userState } = useUserContext();
  const [crawlers, setCrawlers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [documents, setDocuments] = useState([]);
  const [alert, setAlert] = useState(null);
  const [modalError, setModalError] = useState(null);

  const [modalOpened, { open: openModal, close: closeModal }] = useDisclosure(false);
  const [editingCrawler, setEditingCrawler] = useState(null);
  const [form, setForm] = useState(emptyForm);

  const [expandedRuns, setExpandedRuns] = useState({});
  const [crawlerRuns, setCrawlerRuns] = useState({});
  const [loadingRuns, setLoadingRuns] = useState({});

  const [targetOptions, setTargetOptions] = useState([]);
  const [searchingTargets, setSearchingTargets] = useState(false);

  const showAlert = (message, color = "blue") => {
    setAlert({ message, color });
    setTimeout(() => setAlert(null), 4000);
  };

  const fetchCrawlers = async () => {
    try {
      const data = await getCrawlers();
      setCrawlers(data.crawlers || []);
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const [crawlersData, docsData] = await Promise.all([
        getCrawlers(),
        getDocuments(1, 100).catch(() => ({ items: [] })),
      ]);
      setCrawlers(crawlersData.crawlers || []);
      setDocuments(docsData.items || []);
    } catch (error) {
      showAlert(error.message, "red");
    } finally {
      setLoading(false);
    }
  };

  const searchTargetsByType = async (targetType, query) => {
    if (targetType === "individual") {
      const results = await searchIndividuals(query);
      return (results || []).map((ind) => ({
        value: String(ind.id),
        label: `${ind.first_name} ${ind.last_name}`.trim() || ind.email || `ID: ${ind.id}`,
      }));
    }
    if (targetType === "organization") {
      const results = await searchOrganizations(query);
      return (results || []).map((org) => ({
        value: String(org.id),
        label: org.name || `ID: ${org.id}`,
      }));
    }
    if (targetType === "contact") {
      const results = await searchContacts(query);
      return (results || []).map((c) => ({
        value: String(c.id),
        label: c.name || c.email || `ID: ${c.id}`,
      }));
    }
    return [];
  };

  const handleTargetSearch = async (query) => {
    if (!query || query.length < 2 || !form.target_type) {
      return;
    }
    setSearchingTargets(true);
    try {
      const options = await searchTargetsByType(form.target_type, query);
      const safeOptions = Array.isArray(options) ? options : [];
      const currentOptions = Array.isArray(targetOptions) ? targetOptions : [];
      const currentIds = Array.isArray(form.target_ids) ? form.target_ids : [];
      // Merge with existing selected options to preserve selections
      const existingIds = new Set(currentOptions.map((o) => o.value));
      const newOptions = safeOptions.filter((o) => !existingIds.has(o.value));
      setTargetOptions([...currentOptions.filter((o) => currentIds.includes(parseInt(o.value, 10))), ...newOptions]);
    } catch {
      // Keep existing options on error
    } finally {
      setSearchingTargets(false);
    }
  };

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
    fetchData();
  }, [dispatchPageLoading]);

  const handleOpenCreate = () => {
    setEditingCrawler(null);
    setForm(emptyForm);
    setTargetOptions([]);
    setModalError(null);
    openModal();
  };

  const handleOpenEdit = async (crawler) => {
    setEditingCrawler(crawler);
    setForm({
      name: crawler.name || "",
      entity_type: crawler.entity_type || "job",
      interval_minutes: crawler.interval_minutes || 30,
      leads_per_run: crawler.leads_per_run || 10,
      concurrent_searches: crawler.concurrent_searches || 1,
      compilation_target: crawler.compilation_target || 100,
      search_keywords: crawler.search_keywords || [],
      search_slots: crawler.search_slots || [],
      ats_mode: crawler.ats_mode ?? true,
      time_filter: crawler.time_filter || "week",
      target_type: crawler.target_type || null,
      target_ids: crawler.target_ids || [],
      source_document_ids: crawler.source_document_ids || [],
      system_prompt: crawler.system_prompt || "",
      digest_enabled: crawler.digest_enabled ?? true,
      digest_emails: crawler.digest_emails || "",
      digest_time: crawler.digest_time || "09:00",
      digest_model: crawler.digest_model || null,
      use_agent: crawler.use_agent ?? false,
      agent_model: crawler.agent_model || "claude-sonnet-4-20250514",
    });

    // Load target entity names if set
    setTargetOptions([]);
    if (crawler.target_ids?.length && crawler.target_type) {
      try {
        const loadedOptions = await Promise.all(
          crawler.target_ids.map(async (id) => {
            if (crawler.target_type === "individual") {
              const ind = await getIndividual(id);
              return { value: String(id), label: `${ind.first_name} ${ind.last_name}`.trim() || ind.email || `ID: ${id}` };
            }
            if (crawler.target_type === "organization") {
              const org = await getOrganization(id);
              return { value: String(id), label: org.name || `ID: ${id}` };
            }
            if (crawler.target_type === "job") {
              const job = await getJob(id);
              return { value: String(id), label: job.job_title || `ID: ${id}` };
            }
            if (crawler.target_type === "contact") {
              const c = await getContact(id);
              return { value: String(id), label: c.name || c.email || `ID: ${id}` };
            }
            return { value: String(id), label: `ID: ${id}` };
          })
        );
        setTargetOptions(loadedOptions);
      } catch {
        setTargetOptions([]);
      }
    }

    setModalError(null);
    openModal();
  };

  const handleSave = async () => {
    setModalError(null);

    if (!form.name.trim()) {
      setModalError("Name is required");
      return;
    }

    if (form.search_keywords.length === 0) {
      setModalError("At least one search keyword is required");
      return;
    }

    // Filter out empty slots before saving
    const cleanedSlots = (form.search_slots || []).filter((slot) => slot && slot.length > 0);

    // When using multiple concurrent searches, at least one slot must have keywords
    if (form.concurrent_searches > 1 && cleanedSlots.length === 0) {
      setModalError("Assign at least one keyword to a search slot");
      return;
    }

    const cleanedForm = {
      ...form,
      search_slots: cleanedSlots,
    };

    setSaving(true);
    try {
      if (editingCrawler) {
        await updateCrawler(editingCrawler.id, cleanedForm);
        showAlert("Crawler updated", "lime");
      } else {
        await createCrawler(cleanedForm);
        showAlert("Crawler created", "lime");
      }
      await fetchCrawlers();
      closeModal();
    } catch (error) {
      setModalError(error.message);
    } finally {
      setSaving(false);
    }
  };

  const handleToggle = async (crawler, enabled) => {
    try {
      await toggleCrawler(crawler.id, enabled);
      await fetchCrawlers();
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const handleDelete = async (crawler) => {
    if (!window.confirm(`Delete crawler "${crawler.name}"?`)) return;
    try {
      await deleteCrawler(crawler.id);
      await fetchCrawlers();
      showAlert("Crawler deleted", "lime");
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const handleTestRun = async (crawler) => {
    try {
      await testCrawler(crawler.id);
      showAlert(`Test run initiated for "${crawler.name}"`, "blue");
      setTimeout(() => toggleRunsExpanded(crawler.id, true), 2000);
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const toggleRunsExpanded = async (crawlerId, forceExpand = false) => {
    const isExpanded = expandedRuns[crawlerId];

    if (!isExpanded || forceExpand) {
      setLoadingRuns((prev) => ({ ...prev, [crawlerId]: true }));
      try {
        const data = await getCrawlerRuns(crawlerId, 10);
        setCrawlerRuns((prev) => ({ ...prev, [crawlerId]: data.runs || [] }));
      } catch (error) {
        showAlert(error.message, "red");
      } finally {
        setLoadingRuns((prev) => ({ ...prev, [crawlerId]: false }));
      }
    }

    if (!forceExpand) {
      setExpandedRuns((prev) => ({ ...prev, [crawlerId]: !isExpanded }));
    } else {
      setExpandedRuns((prev) => ({ ...prev, [crawlerId]: true }));
    }
  };

  const updateForm = (field, value) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  };

  if (!userState.isAuthed) {
    return (
      <Container size="lg" py="xl">
        <Text>Please log in to access this page.</Text>
      </Container>
    );
  }

  if (loading) {
    return (
      <Container size="lg" py="xl">
        <Stack align="center" gap="md">
          <Loader />
          <Text>Loading automation crawlers...</Text>
        </Stack>
      </Container>
    );
  }

  const documentOptions = documents.map((doc) => ({
    value: String(doc.id),
    label: doc.filename,
  }));

  return (
    <Container size="lg" py="xl">
      <Stack gap="xl">
        {alert && (
          <Alert
            color={alert.color}
            icon={alert.color === "red" ? <AlertCircle size={16} /> : <CheckCircle size={16} />}
            withCloseButton
            onClose={() => setAlert(null)}
          >
            {alert.message}
          </Alert>
        )}

        <Group justify="space-between">
          <div>
            <Title order={1}>Automation Crawlers</Title>
            <Text c="dimmed">Configure automated lead sourcing crawlers for different entity types.</Text>
          </div>
          <Button leftSection={<Plus size={16} />} onClick={handleOpenCreate}>
            New Crawler
          </Button>
        </Group>

        {crawlers.length === 0 ? (
          <Paper p="xl" withBorder ta="center">
            <Text c="dimmed">No crawlers configured. Create one to get started.</Text>
          </Paper>
        ) : (
          <Stack gap="md">
            {crawlers.map((crawler) => (
              <Paper key={crawler.id} p="md" withBorder>
                <Stack gap="sm">
                  <Group justify="space-between">
                    <Group>
                      <ActionIcon
                        variant="subtle"
                        onClick={() => toggleRunsExpanded(crawler.id)}
                        size="sm"
                      >
                        {expandedRuns[crawler.id] ? (
                          <ChevronDown size={16} />
                        ) : (
                          <ChevronRight size={16} />
                        )}
                      </ActionIcon>
                      <div>
                        <Group gap="xs">
                          <Text fw={500}>{crawler.name}</Text>
                          <Text size="xs" c="dimmed">(Target: {crawler.compilation_target})</Text>
                        </Group>
                        <Text size="xs" c="dimmed">
                          {ENTITY_TYPES.find((t) => t.value === crawler.entity_type)?.label || crawler.entity_type}
                          {` • ${crawler.total_proposals_created || 0} proposals`}
                          {crawler.next_run_at && ` • Next: ${new Date(crawler.next_run_at).toLocaleString()}`}
                        </Text>
                      </div>
                    </Group>

                    <Group gap="xs">
                      {crawler.total_proposals_created >= crawler.compilation_target && (
                        <Badge color="lime" size="sm" leftSection={<CheckCircle size={12} />}>
                          Compiled
                        </Badge>
                      )}
                      <Badge color={crawler.enabled ? "lime" : "gray"} size="sm">
                        {crawler.enabled ? "Active" : "Inactive"}
                      </Badge>
                      <Switch
                        checked={crawler.enabled}
                        onChange={(e) => handleToggle(crawler, e.currentTarget.checked)}
                        size="sm"
                        color="lime"
                      />
                      <Menu position="bottom-end" shadow="md">
                        <Menu.Target>
                          <ActionIcon variant="subtle">
                            <MoreVertical size={16} />
                          </ActionIcon>
                        </Menu.Target>
                        <Menu.Dropdown>
                          <Menu.Item leftSection={<Edit size={14} />} onClick={() => handleOpenEdit(crawler)}>
                            Edit
                          </Menu.Item>
                          <Menu.Item leftSection={<Play size={14} />} onClick={() => handleTestRun(crawler)}>
                            Test Run
                          </Menu.Item>
                          <Menu.Divider />
                          <Menu.Item
                            leftSection={<Trash size={14} />}
                            color="red"
                            onClick={() => handleDelete(crawler)}
                          >
                            Delete
                          </Menu.Item>
                        </Menu.Dropdown>
                      </Menu>
                    </Group>
                  </Group>

                  <Collapse in={expandedRuns[crawler.id]}>
                    {loadingRuns[crawler.id] ? (
                      <Group justify="center" py="md">
                        <Loader size="sm" />
                      </Group>
                    ) : (
                      <RunHistoryTable runs={crawlerRuns[crawler.id] || []} />
                    )}
                  </Collapse>
                </Stack>
              </Paper>
            ))}
          </Stack>
        )}
      </Stack>

      <Modal
        opened={modalOpened}
        onClose={closeModal}
        title={editingCrawler ? "Edit Crawler" : "New Crawler"}
        size={"xl"}
      >
        <Stack gap="md">
          {modalError && (
            <Alert color="red" icon={<AlertCircle size={16} />} withCloseButton onClose={() => setModalError(null)}>
              {modalError}
            </Alert>
          )}
          <Group grow>
            <TextInput
              label="Name"
              placeholder="My Job Crawler"
              value={form.name}
              onChange={(e) => updateForm("name", e.currentTarget.value)}
              required
            />
            <Select
              label="Entity Type"
              value={form.entity_type}
              onChange={(val) => {
                updateForm("entity_type", val);
                updateForm("target_type", null);
                updateForm("target_ids", []);
                setTargetOptions([]);
              }}
              data={ENTITY_TYPES}
            />
          </Group>

          <Group grow>
            <NumberInput
              label="Interval (minutes)"
              value={form.interval_minutes}
              onChange={(val) => updateForm("interval_minutes", val)}
              min={1}
              max={1440}
              leftSection={<Clock size={16} />}
            />
            <NumberInput
              label="Leads per Run"
              value={form.leads_per_run}
              onChange={(val) => updateForm("leads_per_run", val)}
              min={1}
              max={50}
            />
          </Group>

          <Group grow>
            <NumberInput
              label="Concurrent Searches"
              value={form.concurrent_searches}
              onChange={(val) => updateForm("concurrent_searches", val)}
              min={1}
              max={10}
            />
            <NumberInput
              label="Compilation Target"
              value={form.compilation_target}
              onChange={(val) => updateForm("compilation_target", val)}
              min={10}
              max={1000}
            />
          </Group>

          <TagsInput
            label="Search Keywords"
            placeholder="Enter keyword and press Enter"
            value={form.search_keywords}
            onChange={(val) => updateForm("search_keywords", val)}
            leftSection={<Search size={16} />}
          />

          {form.search_keywords.length > 0 && form.concurrent_searches > 1 && (
            <Stack gap="xs">
              <Text size="sm" fw={500}>Assign Keywords to Search Slots</Text>
              <Text size="xs" c="dimmed">Each slot runs as a separate concurrent search. Unassigned keywords are not searched.</Text>
              {Array.from({ length: form.concurrent_searches }).map((_, slotIndex) => (
                <MultiSelect
                  key={slotIndex}
                  label={`Slot ${slotIndex + 1}`}
                  placeholder="Select keywords for this slot"
                  data={form.search_keywords.map((kw, i) => ({ value: String(i), label: kw }))}
                  value={(form.search_slots[slotIndex] || []).map(String)}
                  onChange={(vals) => {
                    const newSlots = [...(form.search_slots || [])];
                    while (newSlots.length < form.concurrent_searches) {
                      newSlots.push([]);
                    }
                    newSlots[slotIndex] = vals.map(Number);
                    updateForm("search_slots", newSlots);
                  }}
                  clearable
                />
              ))}
            </Stack>
          )}

          <Group grow>
            <Select
              label="Time Filter"
              value={form.time_filter}
              onChange={(val) => updateForm("time_filter", val)}
              data={[
                { value: "day", label: "Past Day" },
                { value: "week", label: "Past Week" },
                { value: "month", label: "Past Month" },
              ]}
            />
            <Switch
              label="ATS Mode"
              description="Focus on direct application links"
              checked={form.ats_mode}
              onChange={(e) => updateForm("ats_mode", e.currentTarget.checked)}
              color="lime"
              mt="md"
            />
          </Group>

          {TARGET_TYPE_OPTIONS[form.entity_type]?.length > 0 && (
            <>
              <Select
                label="Target Type"
                description="What entities to match against"
                value={form.target_type}
                onChange={(val) => {
                  updateForm("target_type", val);
                  updateForm("target_ids", []);
                  setTargetOptions([]);
                }}
                data={TARGET_TYPE_OPTIONS[form.entity_type] || []}
                clearable
                placeholder="Select target type..."
              />

              {form.target_type && (
                <MultiSelect
                  label="Targets"
                  description="Select one or more targets (type to search)"
                  value={(form.target_ids || []).map(String)}
                  onChange={(vals) => updateForm("target_ids", (vals || []).map((v) => parseInt(v, 10)))}
                  data={Array.isArray(targetOptions) ? targetOptions : []}
                  searchable
                  clearable
                  placeholder="Search and select..."
                  onSearchChange={handleTargetSearch}
                  nothingFoundMessage={searchingTargets ? "Searching..." : "Type at least 2 chars to search"}
                  filter={({ options }) => options}
                />
              )}
            </>
          )}

          <MultiSelect
            label="Source Documents"
            description="Load documents into context for the automation"
            value={(form.source_document_ids || []).map(String)}
            onChange={(vals) => updateForm("source_document_ids", (vals || []).map((v) => parseInt(v, 10)))}
            data={documentOptions}
            searchable
            clearable
            placeholder="Select documents..."
          />

          <Switch
            label="Use Agent Review"
            description="LLM reviews search results before creating proposals"
            checked={form.use_agent}
            onChange={(e) => updateForm("use_agent", e.currentTarget.checked)}
            color="lime"
          />

          {form.use_agent && (
            <Select
              label="Agent Model"
              value={form.agent_model}
              onChange={(val) => updateForm("agent_model", val)}
              data={AGENT_MODEL_OPTIONS}
            />
          )}

          <Textarea
            label="System Prompt"
            placeholder="Custom instructions for the automation agent..."
            value={form.system_prompt}
            onChange={(e) => updateForm("system_prompt", e.currentTarget.value)}
            minRows={3}
            autosize
          />

          <Switch
            label="Daily Digest"
            description="Send a daily summary email of automation activity"
            checked={form.digest_enabled}
            onChange={(e) => updateForm("digest_enabled", e.currentTarget.checked)}
            color="lime"
          />

          {form.digest_enabled && (
            <>
              <Group grow>
                <TextInput
                  label="Digest Recipients"
                  placeholder="email1@example.com, email2@example.com"
                  value={form.digest_emails}
                  onChange={(e) => updateForm("digest_emails", e.currentTarget.value)}
                  leftSection={<Mail size={16} />}
                />
                <TextInput
                  label="Send Time"
                  type="time"
                  value={form.digest_time}
                  onChange={(e) => updateForm("digest_time", e.currentTarget.value)}
                />
              </Group>

              <Select
                label="Digest AI Model (Optional)"
                description="Use AI to generate insightful digest summaries"
                value={form.digest_model}
                onChange={(val) => updateForm("digest_model", val)}
                data={AGENT_MODEL_OPTIONS}
                clearable
                placeholder="None - use simple format"
              />
            </>
          )}

          <Group justify="flex-end" mt="md">
            <Button variant="default" onClick={closeModal}>
              Cancel
            </Button>
            <Button onClick={handleSave} loading={saving}>
              {editingCrawler ? "Update" : "Create"}
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Container>
  );
};

const RunHistoryTable = ({ runs }) => {
  if (runs.length === 0) {
    return (
      <Text c="dimmed" size="sm" ta="center" py="md">
        No runs yet.
      </Text>
    );
  }

  return (
    <Table size="sm">
      <Table.Thead>
        <Table.Tr>
          <Table.Th>Started</Table.Th>
          <Table.Th>Status</Table.Th>
          <Table.Th>Leads</Table.Th>
          <Table.Th>Proposals</Table.Th>
          <Table.Th>Query</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {runs.map((run) => (
          <Table.Tr key={run.id}>
            <Table.Td>{new Date(run.started_at).toLocaleString()}</Table.Td>
            <Table.Td>
              <Badge
                size="xs"
                color={run.status === "completed" ? "lime" : run.status === "running" ? "blue" : "red"}
              >
                {run.status}
              </Badge>
            </Table.Td>
            <Table.Td>{run.leads_found}</Table.Td>
            <Table.Td>{run.proposals_created}</Table.Td>
            <Table.Td>
              <Text size="xs" lineClamp={1}>
                {run.search_query || "-"}
              </Text>
            </Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  );
};

export default AutomationPage;
