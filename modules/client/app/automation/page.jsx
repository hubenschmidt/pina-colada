"use client";

import { useEffect, useState, useCallback, useMemo } from "react";
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
  Select,
  MultiSelect,
  TagsInput,
  Loader,
  Alert,
  Modal,
  ActionIcon,
  Menu,
  Tooltip,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { useMantineColorScheme } from "@mantine/core";
import {
  Play,
  Clock,
  AlertCircle,
  CheckCircle,
  Plus,
  MoreVertical,
  Edit,
  Trash,
  RefreshCw,
  ChevronDown,
  ChevronRight,
  Copy,
  HelpCircle,
} from "lucide-react";
import { DataTable } from "../../components/DataTable/DataTable";
import SSEIndicator from "../../components/SSEIndicator/SSEIndicator";
import { useSSE } from "../../hooks/useSSE";
import { usePageLoading } from "../../context/pageLoadingContext";
import { useUserContext } from "../../context/userContext";
import {
  getCrawlers,
  createCrawler,
  updateCrawler,
  deleteCrawler,
  toggleCrawler,
  testCrawler,
  clearRejectedJobs,
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
  { value: "claude-sonnet-4-5-20250929", label: "Claude Sonnet 4.5 (Recommended)" },
  { value: "claude-haiku-4-5-20251001", label: "Claude Haiku 4.5 (Faster)" },
  { value: "gpt-5.4", label: "GPT 5.4" },
  { value: "gpt-5.4-mini", label: "GPT 5.4 Mini" },
  { value: "gpt-5.2", label: "GPT 5.2" },
];

const US_STATES = [
  { value: "Remote", label: "Remote" },
  { value: "Alabama", label: "Alabama" },
  { value: "Alaska", label: "Alaska" },
  { value: "Arizona", label: "Arizona" },
  { value: "Arkansas", label: "Arkansas" },
  { value: "California", label: "California" },
  { value: "Colorado", label: "Colorado" },
  { value: "Connecticut", label: "Connecticut" },
  { value: "Delaware", label: "Delaware" },
  { value: "Florida", label: "Florida" },
  { value: "Georgia", label: "Georgia" },
  { value: "Hawaii", label: "Hawaii" },
  { value: "Idaho", label: "Idaho" },
  { value: "Illinois", label: "Illinois" },
  { value: "Indiana", label: "Indiana" },
  { value: "Iowa", label: "Iowa" },
  { value: "Kansas", label: "Kansas" },
  { value: "Kentucky", label: "Kentucky" },
  { value: "Louisiana", label: "Louisiana" },
  { value: "Maine", label: "Maine" },
  { value: "Maryland", label: "Maryland" },
  { value: "Massachusetts", label: "Massachusetts" },
  { value: "Michigan", label: "Michigan" },
  { value: "Minnesota", label: "Minnesota" },
  { value: "Mississippi", label: "Mississippi" },
  { value: "Missouri", label: "Missouri" },
  { value: "Montana", label: "Montana" },
  { value: "Nebraska", label: "Nebraska" },
  { value: "Nevada", label: "Nevada" },
  { value: "New Hampshire", label: "New Hampshire" },
  { value: "New Jersey", label: "New Jersey" },
  { value: "New Mexico", label: "New Mexico" },
  { value: "New York", label: "New York" },
  { value: "North Carolina", label: "North Carolina" },
  { value: "North Dakota", label: "North Dakota" },
  { value: "Ohio", label: "Ohio" },
  { value: "Oklahoma", label: "Oklahoma" },
  { value: "Oregon", label: "Oregon" },
  { value: "Pennsylvania", label: "Pennsylvania" },
  { value: "Rhode Island", label: "Rhode Island" },
  { value: "South Carolina", label: "South Carolina" },
  { value: "South Dakota", label: "South Dakota" },
  { value: "Tennessee", label: "Tennessee" },
  { value: "Texas", label: "Texas" },
  { value: "Utah", label: "Utah" },
  { value: "Vermont", label: "Vermont" },
  { value: "Virginia", label: "Virginia" },
  { value: "Washington", label: "Washington" },
  { value: "West Virginia", label: "West Virginia" },
  { value: "Wisconsin", label: "Wisconsin" },
  { value: "Wyoming", label: "Wyoming" },
  { value: "District of Columbia", label: "District of Columbia" },
];

const INTERVAL_UNITS = [
  { value: "seconds", label: "Seconds" },
  { value: "minutes", label: "Minutes" },
];

const emptyCrawlerForm = {
  name: "",
  entity_type: "job",
  interval_value: 30,
  interval_unit: "minutes",
  compilation_target: 100,
  disable_on_compiled: false,
  search_query: "",
  suggested_query: null,
  use_suggested_query: false,
  location: "",
  ats_mode: true,
  time_filter: "week",
  target_type: null,
  target_ids: [],
  source_document_ids: [],
  system_prompt: "",
  suggested_prompt: null,
  use_suggested_prompt: false,
  suggestion_threshold: 50,
  min_prospects_threshold: 5,
  use_agent: true,
  agent_model: "claude-opus-4-6",
  vector_prefilter_enabled: true,
  vector_similarity_threshold: 0.4,
  use_analytics: true,
  analytics_model: "claude-opus-4-6",
  empty_proposal_limit: null,
  analytics_run_limit: 20,
  prompt_cooldown_runs: 5,
  prompt_cooldown_prospects: 50,
  domain_blacklist: [],
};

const secondsToInterval = (seconds) => {
  const useMinutes = seconds >= 60 && seconds % 60 === 0;
  return {
    value: useMinutes ? seconds / 60 : seconds,
    unit: useMinutes ? "minutes" : "seconds",
  };
};

const intervalToSeconds = (value, unit) => (unit === "minutes" ? value * 60 : value);

const AutomationPage = () => {
  const { dispatchPageLoading } = usePageLoading();
  const { userState } = useUserContext();

  // Shared state
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [documents, setDocuments] = useState([]);
  const [alert, setAlert] = useState(null);
  const [modalError, setModalError] = useState(null);

  // Crawler state
  const [crawlers, setCrawlers] = useState([]);
  const [crawlerModalOpened, { open: openCrawlerModal, close: closeCrawlerModal }] =
    useDisclosure(false);
  const [editingCrawler, setEditingCrawler] = useState(null);
  const [crawlerForm, setCrawlerForm] = useState(emptyCrawlerForm);
  const [initialCrawlerForm, setInitialCrawlerForm] = useState(emptyCrawlerForm);
  const [expandedCrawlerRuns, setExpandedCrawlerRuns] = useState({});
  const [crawlerPages, setCrawlerPages] = useState({});
  const [crawlerPageSizes, setCrawlerPageSizes] = useState({});
  const [crawlerSseStatus, setCrawlerSseStatus] = useState({});
  const [targetOptions, setTargetOptions] = useState([]);
  const [searchingTargets, setSearchingTargets] = useState(false);

  const showAlert = (message, color = "blue") => {
    setAlert({ message, color });
    setTimeout(() => setAlert(null), 4000);
  };

  // ── Data fetching ──

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

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
    fetchData();
  }, [dispatchPageLoading]);

  // ── Target search (crawler-specific) ──

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
    if (!query || query.length < 2 || !crawlerForm.target_type) return;
    setSearchingTargets(true);
    try {
      const options = await searchTargetsByType(crawlerForm.target_type, query);
      const safeOptions = Array.isArray(options) ? options : [];
      const currentOptions = Array.isArray(targetOptions) ? targetOptions : [];
      const currentIds = Array.isArray(crawlerForm.target_ids) ? crawlerForm.target_ids : [];
      const existingIds = new Set(currentOptions.map((o) => o.value));
      const newOptions = safeOptions.filter((o) => !existingIds.has(o.value));
      setTargetOptions([
        ...currentOptions.filter((o) => currentIds.includes(parseInt(o.value, 10))),
        ...newOptions,
      ]);
    } catch {
      // Keep existing options on error
    } finally {
      setSearchingTargets(false);
    }
  };

  const TARGET_LABEL_FETCHERS = {
    individual: async (id) => {
      const ind = await getIndividual(id);
      return `${ind.first_name} ${ind.last_name}`.trim() || ind.email || `ID: ${id}`;
    },
    organization: async (id) => {
      const org = await getOrganization(id);
      return org.name || `ID: ${id}`;
    },
    job: async (id) => {
      const job = await getJob(id);
      return job.job_title || `ID: ${id}`;
    },
    contact: async (id) => {
      const c = await getContact(id);
      return c.name || c.email || `ID: ${id}`;
    },
  };

  const loadTargetOptions = async (crawler) => {
    setTargetOptions([]);
    if (!crawler.target_ids?.length || !crawler.target_type) return;
    const fetcher = TARGET_LABEL_FETCHERS[crawler.target_type];
    try {
      const loadedOptions = await Promise.all(
        crawler.target_ids.map(async (id) => ({
          value: String(id),
          label: fetcher ? await fetcher(id) : `ID: ${id}`,
        }))
      );
      setTargetOptions(loadedOptions);
    } catch {
      setTargetOptions([]);
    }
  };

  // ── Crawler handlers ──

  const crawlerToFormData = (crawler, nameOverride) => {
    const { value, unit } = secondsToInterval(crawler.interval_seconds || 1800);
    return {
      name: nameOverride ?? (crawler.name || ""),
      entity_type: crawler.entity_type || "job",
      interval_value: value,
      interval_unit: unit,
      compilation_target: crawler.compilation_target || 100,
      disable_on_compiled: crawler.disable_on_compiled ?? false,
      search_query: crawler.search_query || "",
      suggested_query: crawler.suggested_query || null,
      use_suggested_query: crawler.use_suggested_query ?? false,
      location: crawler.location || "",
      ats_mode: crawler.ats_mode ?? true,
      time_filter: crawler.time_filter || "week",
      target_type: crawler.target_type || null,
      target_ids: crawler.target_ids || [],
      source_document_ids: crawler.source_document_ids || [],
      system_prompt: crawler.system_prompt || "",
      suggested_prompt: crawler.suggested_prompt || null,
      use_suggested_prompt: crawler.use_suggested_prompt ?? false,
      suggestion_threshold: crawler.suggestion_threshold ?? 50,
      min_prospects_threshold: crawler.min_prospects_threshold ?? 5,
      use_agent: crawler.use_agent ?? false,
      agent_model: crawler.agent_model || "claude-sonnet-4-5-20250929",
      vector_prefilter_enabled: crawler.vector_prefilter_enabled ?? false,
      vector_similarity_threshold: crawler.vector_similarity_threshold ?? 0.3,
      use_analytics: crawler.use_analytics ?? false,
      analytics_model: crawler.analytics_model || "claude-3-haiku-20240307",
      empty_proposal_limit: crawler.empty_proposal_limit || null,
      analytics_run_limit: crawler.analytics_run_limit ?? 20,
      prompt_cooldown_runs: crawler.prompt_cooldown_runs ?? 5,
      prompt_cooldown_prospects: crawler.prompt_cooldown_prospects ?? 50,
      domain_blacklist: crawler.domain_blacklist || [],
    };
  };

  const handleOpenCrawlerCreate = () => {
    setEditingCrawler(null);
    setCrawlerForm(emptyCrawlerForm);
    setInitialCrawlerForm(emptyCrawlerForm);
    setTargetOptions([]);
    setModalError(null);
    openCrawlerModal();
  };

  const handleOpenCrawlerEdit = async (crawler) => {
    setEditingCrawler(crawler);
    const formData = crawlerToFormData(crawler);
    setCrawlerForm(formData);
    setInitialCrawlerForm(formData);
    await loadTargetOptions(crawler);
    setModalError(null);
    openCrawlerModal();
  };

  const handleDuplicateCrawler = async (crawler) => {
    setEditingCrawler(null);
    const formData = crawlerToFormData(crawler, `Copy of ${crawler.name || ""}`);
    setCrawlerForm(formData);
    setInitialCrawlerForm(formData);
    await loadTargetOptions(crawler);
    setModalError(null);
    openCrawlerModal();
  };

  const handleSaveCrawler = async () => {
    setModalError(null);
    if (!crawlerForm.name.trim()) {
      setModalError("Name is required");
      return;
    }
    const query =
      crawlerForm.use_suggested_query && crawlerForm.suggested_query
        ? crawlerForm.suggested_query
        : crawlerForm.search_query;
    if (!query || !query.trim()) {
      setModalError("Search query is required");
      return;
    }
    const cleanedForm = {
      ...crawlerForm,
      interval_seconds: intervalToSeconds(crawlerForm.interval_value, crawlerForm.interval_unit),
      empty_proposal_limit: crawlerForm.empty_proposal_limit || 0,
    };
    delete cleanedForm.interval_value;
    delete cleanedForm.interval_unit;

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
      closeCrawlerModal();
    } catch (error) {
      setModalError(error.message);
    } finally {
      setSaving(false);
    }
  };

  const handleToggleCrawler = async (crawler, enabled) => {
    try {
      await toggleCrawler(crawler.id, enabled);
      await fetchCrawlers();
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const handleDeleteCrawler = async (crawler) => {
    if (!window.confirm(`Delete crawler "${crawler.name}"?`)) return;
    try {
      await deleteCrawler(crawler.id);
      await fetchCrawlers();
      showAlert("Crawler deleted", "lime");
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const handleTestCrawler = async (crawler) => {
    try {
      await testCrawler(crawler.id);
      showAlert(`Test run initiated for "${crawler.name}"`, "blue");
      setTimeout(() => setExpandedCrawlerRuns((prev) => ({ ...prev, [crawler.id]: true })), 2000);
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const handleClearCrawlerRejected = async (crawler) => {
    if (
      !window.confirm(
        `Clear all rejected jobs for "${crawler.name}"? This will allow previously rejected results to be re-evaluated.`
      )
    )
      return;
    try {
      await clearRejectedJobs(crawler.id);
      showAlert("Rejected jobs cleared", "lime");
    } catch (error) {
      showAlert(error.message, "red");
    }
  };

  const handleCrawlerSseStatus = useCallback((crawlerId, status) => {
    setCrawlerSseStatus((prev) => ({ ...prev, [crawlerId]: status }));
  }, []);

  const handleCrawlerUpdate = useCallback((crawlerId, updates) => {
    setCrawlers((prev) => prev.map((c) => (c.id === crawlerId ? { ...c, ...updates } : c)));
  }, []);

  const updateCrawlerForm = (field, value) => {
    setCrawlerForm((prev) => ({ ...prev, [field]: value }));
  };

  const isCrawlerFormDirty = JSON.stringify(crawlerForm) !== JSON.stringify(initialCrawlerForm);

  // ── Shared ──

  const documentOptions = useMemo(
    () =>
      documents.map((doc) => ({
        value: String(doc.id),
        label: `${doc.filename} (v${doc.version_number || 1})`,
      })),
    [documents]
  );

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
          <Text>Loading automation...</Text>
        </Stack>
      </Container>
    );
  }

  return (
    <Container fluid py="xl" px="xl">
      <Stack gap="xl">
        {alert && (
          <Alert
            color={alert.color}
            icon={alert.color === "red" ? <AlertCircle size={16} /> : <CheckCircle size={16} />}
            withCloseButton
            onClose={() => setAlert(null)}>
            {alert.message}
          </Alert>
        )}

        <Stack gap="md">
          <Group justify="space-between">
            <div>
              <Title order={2}>Crawlers</Title>
              <Text c="dimmed" size="sm">
                Automated lead sourcing via search engines.
              </Text>
            </div>
            <Button leftSection={<Plus size={16} />} onClick={handleOpenCrawlerCreate}>
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
                <AutomationCard
                  key={crawler.id}
                  item={crawler}
                  entityLabel={
                    ENTITY_TYPES.find((t) => t.value === crawler.entity_type)?.label ||
                    crawler.entity_type
                  }
                  suggestedLine={
                    crawler.suggested_query && crawler.use_suggested_query
                      ? `Suggested query: ${crawler.suggested_query}${crawler.location ? ` "${crawler.location}"` : ""}`
                      : null
                  }
                  sseStatus={crawlerSseStatus[crawler.id]}
                  isExpanded={expandedCrawlerRuns[crawler.id]}
                  onToggleExpand={() =>
                    setExpandedCrawlerRuns((prev) => ({ ...prev, [crawler.id]: !prev[crawler.id] }))
                  }
                  onToggle={(enabled) => handleToggleCrawler(crawler, enabled)}
                  onEdit={() => handleOpenCrawlerEdit(crawler)}
                  onDuplicate={() => handleDuplicateCrawler(crawler)}
                  onTestRun={() => handleTestCrawler(crawler)}
                  onClearRejected={() => handleClearCrawlerRejected(crawler)}
                  onDelete={() => handleDeleteCrawler(crawler)}
                  runsComponent={
                    expandedCrawlerRuns[crawler.id] && (
                      <CrawlerRunsSSE
                        crawlerId={crawler.id}
                        isExpanded={expandedCrawlerRuns[crawler.id]}
                        pageValue={crawlerPages[crawler.id] || 1}
                        pageSizeValue={crawlerPageSizes[crawler.id] || 10}
                        onPageChange={(page) =>
                          setCrawlerPages((prev) => ({ ...prev, [crawler.id]: page }))
                        }
                        onPageSizeChange={(size) => {
                          setCrawlerPageSizes((prev) => ({ ...prev, [crawler.id]: size }));
                          setCrawlerPages((prev) => ({ ...prev, [crawler.id]: 1 }));
                        }}
                        onStatusChange={handleCrawlerSseStatus}
                        onCrawlerUpdate={handleCrawlerUpdate}
                      />
                    )
                  }
                />
              ))}
            </Stack>
          )}
        </Stack>
      </Stack>

      {/* ── Crawler Modal ── */}
      <Modal
        opened={crawlerModalOpened}
        onClose={closeCrawlerModal}
        title={
          <Group justify="space-between" w="100%">
            <Text fw={500}>{editingCrawler ? "Edit Crawler" : "New Crawler"}</Text>
            <Group gap="xs" mr="xl">
              <Button variant="default" size="sm" onClick={closeCrawlerModal}>
                Cancel
              </Button>
              <Button
                size="sm"
                color={isCrawlerFormDirty ? "lime" : "gray"}
                onClick={handleSaveCrawler}
                loading={saving}>
                Save
              </Button>
            </Group>
          </Group>
        }
        size="1600"
        styles={{ title: { width: "100%" } }}>
        <Stack gap="md">
          {modalError && (
            <Alert
              color="red"
              icon={<AlertCircle size={16} />}
              withCloseButton
              onClose={() => setModalError(null)}>
              {modalError}
            </Alert>
          )}
          <Group grow>
            <TextInput
              label="Name"
              placeholder="My Job Crawler"
              value={crawlerForm.name}
              onChange={(e) => updateCrawlerForm("name", e.currentTarget.value)}
              required
            />
            <Select
              label="Entity Type"
              value={crawlerForm.entity_type}
              onChange={(val) => {
                updateCrawlerForm("entity_type", val);
                updateCrawlerForm("target_type", null);
                updateCrawlerForm("target_ids", []);
                setTargetOptions([]);
              }}
              data={ENTITY_TYPES}
            />
          </Group>

          <Group grow gap="xs">
            <NumberInput
              label="Interval"
              value={crawlerForm.interval_value}
              onChange={(val) => updateCrawlerForm("interval_value", val)}
              min={1}
              max={crawlerForm.interval_unit === "minutes" ? 1440 : 86400}
              leftSection={<Clock size={16} />}
              style={{ flex: 2 }}
            />
            <Select
              label="Unit"
              value={crawlerForm.interval_unit}
              onChange={(val) => updateCrawlerForm("interval_unit", val)}
              data={INTERVAL_UNITS}
              style={{ flex: 1 }}
            />
          </Group>

          <Group grow>
            <NumberInput
              label="Compilation Target"
              value={crawlerForm.compilation_target}
              onChange={(val) => updateCrawlerForm("compilation_target", val)}
              min={1}
              max={1000}
            />
            <Switch
              label="Disable on target"
              description="Else crawler will pause on target"
              checked={crawlerForm.disable_on_compiled}
              onChange={(e) => updateCrawlerForm("disable_on_compiled", e.currentTarget.checked)}
              color="lime"
            />
          </Group>

          <NumberInput
            label="Empty Proposal Limit"
            description="Auto-disable after N consecutive runs with 0 proposals (blank = disabled)"
            placeholder="Disabled"
            value={crawlerForm.empty_proposal_limit ?? ""}
            onChange={(val) => updateCrawlerForm("empty_proposal_limit", val === "" ? null : val)}
            min={1}
            max={100}
            allowDecimal={false}
          />
          <NumberInput
            label="Analytics Run Limit"
            description="Number of recent runs to analyze for historical trends"
            value={crawlerForm.analytics_run_limit}
            onChange={(val) => updateCrawlerForm("analytics_run_limit", val)}
            min={1}
            max={100}
            allowDecimal={false}
          />

          <Stack gap="xs">
            <Textarea
              label="Search Query"
              placeholder="e.g. (typescript OR javascript) node.js intitle:engineer"
              description={'Operators: OR, (), "", intitle:, -exclude'}
              value={crawlerForm.search_query || ""}
              onChange={(e) => updateCrawlerForm("search_query", e.currentTarget.value)}
              minRows={2}
            />
            {crawlerForm.suggested_query && crawlerForm.use_suggested_query && (
              <Group gap="xs" align="center">
                <Text size="xs" c="lime">
                  Suggested: {crawlerForm.suggested_query}
                </Text>
                <Button
                  size="compact-xs"
                  variant="light"
                  color="lime"
                  onClick={() => {
                    updateCrawlerForm("search_query", crawlerForm.suggested_query);
                    updateCrawlerForm("suggested_query", null);
                    updateCrawlerForm("use_suggested_query", false);
                  }}>
                  Accept
                </Button>
                <Button
                  size="compact-xs"
                  variant="subtle"
                  color="gray"
                  onClick={() => {
                    updateCrawlerForm("suggested_query", null);
                    updateCrawlerForm("use_suggested_query", false);
                  }}>
                  Dismiss
                </Button>
              </Group>
            )}
            <Switch
              label="Use suggested query"
              description="Auto-generate and apply AI-suggested search queries"
              checked={crawlerForm.use_suggested_query}
              onChange={(e) => updateCrawlerForm("use_suggested_query", e.currentTarget.checked)}
              color="lime"
              size="sm"
            />
          </Stack>

          <Select
            label="Location"
            description="US state — appended to search queries"
            placeholder="Select state..."
            value={crawlerForm.location || null}
            onChange={(val) => updateCrawlerForm("location", val || "")}
            data={US_STATES}
            searchable
            clearable
          />

          <Group grow>
            <Select
              label="Time Filter"
              value={crawlerForm.time_filter}
              onChange={(val) => updateCrawlerForm("time_filter", val)}
              data={[
                { value: "day", label: "Past Day" },
                { value: "week", label: "Past Week" },
                { value: "month", label: "Past Month" },
              ]}
            />
            <Switch
              label="ATS Mode"
              description="Focus on direct application links"
              checked={crawlerForm.ats_mode}
              onChange={(e) => updateCrawlerForm("ats_mode", e.currentTarget.checked)}
              color="lime"
              mt="md"
            />
          </Group>

          {TARGET_TYPE_OPTIONS[crawlerForm.entity_type]?.length > 0 && (
            <>
              <Select
                label="Target Type"
                description="What entities to match against"
                value={crawlerForm.target_type}
                onChange={(val) => {
                  updateCrawlerForm("target_type", val);
                  updateCrawlerForm("target_ids", []);
                  setTargetOptions([]);
                }}
                data={TARGET_TYPE_OPTIONS[crawlerForm.entity_type] || []}
                clearable
                placeholder="Select target type..."
              />
              {crawlerForm.target_type && (
                <MultiSelect
                  label="Targets"
                  description="Select one or more targets (type to search)"
                  value={(crawlerForm.target_ids || []).map(String)}
                  onChange={(vals) =>
                    updateCrawlerForm(
                      "target_ids",
                      (vals || []).map((v) => parseInt(v, 10))
                    )
                  }
                  data={Array.isArray(targetOptions) ? targetOptions : []}
                  searchable
                  clearable
                  placeholder="Search and select..."
                  onSearchChange={handleTargetSearch}
                  nothingFoundMessage={
                    searchingTargets ? "Searching..." : "Type at least 2 chars to search"
                  }
                  filter={({ options }) => options}
                />
              )}
            </>
          )}

          <SharedFormFields
            form={crawlerForm}
            updateForm={updateCrawlerForm}
            documentOptions={documentOptions}
          />
        </Stack>
      </Modal>
    </Container>
  );
};

// ── Shared card component ──

const AutomationCard = ({
  item,
  entityLabel,
  suggestedLine,
  sourcesLine,
  sseStatus,
  isExpanded,
  onToggleExpand,
  onToggle,
  onEdit,
  onDuplicate,
  onTestRun,
  onClearRejected,
  onDelete,
  runsComponent,
}) => (
  <Paper p="md" withBorder>
    <Stack gap="sm">
      <Group justify="space-between" wrap="nowrap">
        <Group style={{ minWidth: 0 }}>
          <ActionIcon variant="subtle" onClick={onToggleExpand} size="sm">
            {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
          </ActionIcon>
          <div style={{ minWidth: 0 }}>
            <Group gap="xs">
              <SSEIndicator status={sseStatus} />
              <Text fw={500}>{item.name}</Text>
              <Text size="xs" c="dimmed">
                (Target: {item.compilation_target})
              </Text>
            </Group>
            <Text size="xs" c="dimmed">
              {entityLabel}
              {sourcesLine && ` • ${sourcesLine}`}
              {` • ${item.active_proposals || 0} proposals`}
              {item.next_run_at && ` • Next: ${new Date(item.next_run_at).toLocaleString()}`}
            </Text>
            {suggestedLine && (
              <Text size="xs" c="lime" lineClamp={1}>
                {suggestedLine}
              </Text>
            )}
          </div>
        </Group>

        <Group gap="xs">
          {item.active_proposals >= item.compilation_target && (
            <Badge color="lime" size="sm" leftSection={<CheckCircle size={12} />}>
              Compiled
            </Badge>
          )}
          <Badge color={item.enabled ? "lime" : "gray"} size="sm">
            {item.enabled ? "Active" : "Inactive"}
          </Badge>
          <Switch
            checked={item.enabled}
            onChange={(e) => onToggle(e.currentTarget.checked)}
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
              <Menu.Item leftSection={<Edit size={14} />} onClick={onEdit}>
                Edit
              </Menu.Item>
              <Menu.Item leftSection={<Copy size={14} />} onClick={onDuplicate}>
                Duplicate
              </Menu.Item>
              <Menu.Item leftSection={<Play size={14} />} onClick={onTestRun}>
                Test Run
              </Menu.Item>
              <Menu.Item leftSection={<RefreshCw size={14} />} onClick={onClearRejected}>
                Clear Rejected Jobs
              </Menu.Item>
              <Menu.Divider />
              <Menu.Item leftSection={<Trash size={14} />} color="red" onClick={onDelete}>
                Delete
              </Menu.Item>
            </Menu.Dropdown>
          </Menu>
        </Group>
      </Group>
      {runsComponent}
    </Stack>
  </Paper>
);

// ── Shared form fields (used by both crawler and targeter modals) ──

const SharedFormFields = ({ form, updateForm, documentOptions }) => (
  <>
    <MultiSelect
      label="Source Documents"
      description="Load documents into context for the automation"
      value={(form.source_document_ids || []).map(String)}
      onChange={(vals) =>
        updateForm(
          "source_document_ids",
          (vals || []).map((v) => parseInt(v, 10))
        )
      }
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

    <TagsInput
      label="Domain Blacklist"
      description="Domains to auto-reject (e.g., jobleads.com). Press Enter or comma to add."
      placeholder="Add domain..."
      value={form.domain_blacklist}
      onChange={(val) => updateForm("domain_blacklist", val)}
      splitChars={[","]}
    />

    <Switch
      label="Vector Pre-filter"
      description="Embed results and filter by cosine similarity before LLM review"
      checked={form.vector_prefilter_enabled}
      onChange={(e) => updateForm("vector_prefilter_enabled", e.currentTarget.checked)}
      color="lime"
    />
    {form.vector_prefilter_enabled && (
      <NumberInput
        label={
          <Group gap={4}>
            <Text size="sm" fw={500}>
              Similarity Threshold
            </Text>
            <Tooltip
              label="Each result's full text is embedded and compared against your source documents using cosine similarity. Results below this threshold are filtered out before LLM review."
              multiline
              w={300}
              withArrow>
              <HelpCircle size={14} style={{ cursor: "help", opacity: 0.5 }} />
            </Tooltip>
          </Group>
        }
        description="Minimum cosine similarity (0-1)"
        value={form.vector_similarity_threshold}
        onChange={(val) => updateForm("vector_similarity_threshold", val)}
        min={0}
        max={1}
        step={0.05}
        decimalScale={2}
      />
    )}

    <Switch
      label="Use Prompt Analytics"
      description="LLM analyzes historical runs to improve prompt suggestions"
      checked={form.use_analytics}
      onChange={(e) => updateForm("use_analytics", e.currentTarget.checked)}
      color="lime"
    />
    {form.use_analytics && (
      <Select
        label="Analytics Model"
        description="Cheaper models recommended to reduce costs"
        value={form.analytics_model}
        onChange={(val) => updateForm("analytics_model", val)}
        data={AGENT_MODEL_OPTIONS}
      />
    )}

    <Stack gap="xs">
      <Text size="sm" fw={500}>
        System Prompt
      </Text>
      <Textarea
        placeholder="Custom instructions for the automation agent..."
        value={
          form.use_suggested_prompt && form.suggested_prompt
            ? form.suggested_prompt
            : form.system_prompt
        }
        onChange={(e) => {
          if (form.use_suggested_prompt && form.suggested_prompt) {
            updateForm("suggested_prompt", e.currentTarget.value);
          } else {
            updateForm("system_prompt", e.currentTarget.value);
          }
        }}
        minRows={3}
        autosize
        styles={
          form.use_suggested_prompt && form.suggested_prompt
            ? { input: { borderColor: "var(--mantine-color-lime-6)" } }
            : {}
        }
      />
      {form.suggested_prompt && (
        <Text size="xs" c="dimmed">
          {form.use_suggested_prompt
            ? "Editing suggested prompt. Toggle off to switch back to original."
            : "A suggested prompt is available. Toggle on to use it."}
        </Text>
      )}
    </Stack>

    <Group grow align="flex-start">
      <Switch
        label="Use suggested prompt"
        description="Auto-generate and apply AI-suggested system prompts"
        checked={form?.use_suggested_prompt}
        onChange={(e) => updateForm("use_suggested_prompt", e.currentTarget.checked)}
        color="lime"
        size="sm"
      />
      <NumberInput
        label="Suggestion Threshold"
        description="Trigger suggestions when conversion rate below this %"
        value={form.suggestion_threshold}
        onChange={(val) => updateForm("suggestion_threshold", val)}
        min={0}
        max={100}
        suffix="%"
      />
      <NumberInput
        label="Min Prospects"
        description="Also trigger suggestions if prospects below this count"
        value={form.min_prospects_threshold}
        onChange={(val) => updateForm("min_prospects_threshold", val)}
        min={1}
        max={50}
      />
    </Group>

    <Group grow align="flex-start">
      <NumberInput
        label="Prompt Cooldown Runs"
        description="Min runs after prompt update before suggesting again"
        value={form.prompt_cooldown_runs}
        onChange={(val) => updateForm("prompt_cooldown_runs", val)}
        min={1}
        max={100}
      />
      <NumberInput
        label="Prompt Cooldown Prospects"
        description="Min total prospects during cooldown before suggesting again"
        value={form.prompt_cooldown_prospects}
        onChange={(val) => updateForm("prompt_cooldown_prospects", val)}
        min={1}
        max={500}
      />
    </Group>
  </>
);

// ── SSE-enabled run history for crawlers ──

const CrawlerRunsSSE = ({
  crawlerId,
  isExpanded,
  pageValue,
  pageSizeValue,
  onPageChange,
  onPageSizeChange,
  onStatusChange,
  onCrawlerUpdate,
}) => {
  const [localData, setLocalData] = useState(null);

  const handleSSEEvent = useCallback(
    (type, payload) => {
      if (type === "config_updated") {
        const updates = {
          ...(payload.enabled != null && { enabled: payload.enabled }),
          ...(payload.active_proposals != null && { active_proposals: payload.active_proposals }),
          ...(payload.suggested_query != null && { suggested_query: payload.suggested_query }),
          ...(payload.use_suggested_query != null && {
            use_suggested_query: payload.use_suggested_query,
          }),
          ...(payload.suggested_prompt != null && { suggested_prompt: payload.suggested_prompt }),
          ...(payload.use_suggested_prompt != null && {
            use_suggested_prompt: payload.use_suggested_prompt,
          }),
        };
        onCrawlerUpdate?.(crawlerId, updates);
        return;
      }

      if (type !== "run_started" && type !== "run_completed") return;

      const updates = {
        ...(payload.config_active_proposals != null && {
          active_proposals: payload.config_active_proposals,
        }),
        ...(payload.config_enabled != null && { enabled: payload.config_enabled }),
      };
      if (type === "run_completed" && Object.keys(updates).length > 0) {
        onCrawlerUpdate?.(crawlerId, updates);
      }

      setLocalData((prev) => {
        if (!prev) return prev;
        const items = prev.items || [];
        const existingIndex = items.findIndex((r) => r.id === payload.id);
        if (existingIndex >= 0) {
          const updated = [...items];
          updated[existingIndex] = payload;
          return { ...prev, items: updated };
        }
        return {
          ...prev,
          items: [payload, ...items].slice(0, pageSizeValue),
          totalCount: prev.totalCount + 1,
        };
      });
    },
    [crawlerId, pageSizeValue, onCrawlerUpdate]
  );

  const sseUrl = `/automation/crawlers/${crawlerId}/runs/stream?page=${pageValue}&limit=${pageSizeValue}`;
  const {
    data: sseData,
    isConnected,
    lastPulse,
  } = useSSE(sseUrl, {
    enabled: isExpanded && !!crawlerId,
    onEvent: handleSSEEvent,
  });

  useEffect(() => {
    if (sseData) setLocalData(sseData);
  }, [sseData]);

  const isReconnecting = !isConnected && !!localData;

  useEffect(() => {
    onStatusChange(crawlerId, {
      connected: isConnected,
      reconnecting: isReconnecting,
      pulsing: false,
    });
    return () => onStatusChange(crawlerId, null);
  }, [crawlerId, isConnected, isReconnecting, onStatusChange]);

  useEffect(() => {
    if (!lastPulse) return;
    onStatusChange(crawlerId, { connected: isConnected, reconnecting: false, pulsing: true });
    const timer = setTimeout(
      () =>
        onStatusChange(crawlerId, { connected: isConnected, reconnecting: false, pulsing: false }),
      500
    );
    return () => clearTimeout(timer);
  }, [crawlerId, lastPulse, isConnected, onStatusChange]);

  if (!localData)
    return (
      <Group justify="center" py="md">
        <Loader size="sm" />
      </Group>
    );

  return (
    <Stack gap={4}>
      <CrawlerRunHistoryTable
        data={localData}
        pageValue={pageValue}
        onPageChange={onPageChange}
        pageSizeValue={pageSizeValue}
        onPageSizeChange={onPageSizeChange}
      />
    </Stack>
  );
};

// ── Crawler run history table ──

const CrawlerRunHistoryTable = ({
  data,
  pageValue,
  onPageChange,
  pageSizeValue,
  onPageSizeChange,
}) => {
  const { colorScheme } = useMantineColorScheme();

  const getStatusColor = (status) => {
    if (status === "done") return "lime";
    if (status === "running") return "blue";
    return "red";
  };

  const columns = [
    { header: "Started", accessor: (row) => new Date(row.started_at).toLocaleString() },
    {
      header: "Status",
      render: (row) => (
        <Badge size="xs" color={getStatusColor(row.status)}>
          {row.status}
        </Badge>
      ),
    },
    {
      header: "Compiled",
      render: (row) => (row.compiled ? <CheckCircle size={16} color="lime" /> : null),
    },
    {
      header: "Q+",
      render: (row) => (row.query_updated ? <CheckCircle size={16} color="#228be6" /> : null),
    },
    {
      header: "P+",
      render: (row) => (row.prompt_updated ? <CheckCircle size={16} color="#228be6" /> : null),
    },
    { header: "Prospects", accessor: "prospects_found" },
    { header: "Proposals", accessor: "proposals_created" },
    {
      header: "Query",
      render: (row) => (
        <Text size="xs" style={{ whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
          {row.executed_query || "-"}
        </Text>
      ),
    },
    {
      header: "Prompt",
      render: (row) => (
        <Tooltip
          label={row.executed_system_prompt || "-"}
          multiline
          w={400}
          withArrow
          position="left"
          disabled={!row.executed_system_prompt}
          color={colorScheme === "dark" ? "dark" : "gray"}>
          <Text
            size="xs"
            c="dimmed"
            lineClamp={2}
            style={{ maxWidth: 300, cursor: row.executed_system_prompt ? "help" : "default" }}>
            {row.executed_system_prompt || "-"}
          </Text>
        </Tooltip>
      ),
    },
    { header: "P#", accessor: "executed_system_prompt_charcount" },
  ];

  return (
    <DataTable
      data={data}
      columns={columns}
      rowKey={(row) => row.id}
      pageValue={pageValue}
      onPageChange={onPageChange}
      pageSizeValue={pageSizeValue}
      onPageSizeChange={onPageSizeChange}
      pageSizeOptions={[5, 10, 25, 50]}
      emptyText="No runs yet."
      withTableBorder={false}
      withColumnBorders={false}
    />
  );
};

export default AutomationPage;
