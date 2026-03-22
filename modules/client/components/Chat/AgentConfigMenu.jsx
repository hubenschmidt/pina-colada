import React, { useState, useEffect, useId } from "react";
import { Settings, ChevronDown, ChevronUp, Info, Save, Circle, Square, X } from "lucide-react";
import {
  updateAgentNodeConfig,
  resetAgentNodeConfig,
  createAgentConfigPreset,
  applyAgentConfigPreset,
  applyCostTier,
  startMetricRecording,
  stopMetricRecording,
} from "../../api";
import { useUserContext } from "../../context/userContext";
import { useAgentConfig } from "../../context/agentConfigContext";
import DeveloperFeature from "../DeveloperFeature/DeveloperFeature";
import styles from "./AgentConfigMenu.module.css";

const AgentConfigMenu = () => {
  const { userState } = useUserContext();
  const { config, loading, error: fetchError, fetchConfig, setConfig, isRecording, setIsRecording } = useAgentConfig();
  const [isOpen, setIsOpen] = useState(false);
  const [saving, setSaving] = useState({});
  const [expandedNodes, setExpandedNodes] = useState({});
  const [presetName, setPresetName] = useState("");
  const [showSavePreset, setShowSavePreset] = useState(false);
  const [savingPreset, setSavingPreset] = useState(false);
  const [applyingCostTier, setApplyingCostTier] = useState(false);
  const [recordingLoading, setRecordingLoading] = useState(false);
  const [changingProvider, setChangingProvider] = useState(false);
  const [actionError, setActionError] = useState(null);
  const menuId = useId();

  const error = fetchError || actionError;

  useEffect(() => {
    if (!isOpen || !userState.roles?.includes("developer")) return;
    fetchConfig();
  }, [isOpen, userState.roles, fetchConfig]);

  // Sidebar uses backdrop click instead of click-outside detection

  const handleNodeUpdate = async (nodeName, updates) => {
    setSaving((prev) => ({ ...prev, [nodeName]: true }));

    try {
      const node = config.nodes.find((n) => n.node_name === nodeName);
      const payload = {
        model: updates.model ?? node.model,
        temperature: updates.temperature ?? node.temperature,
        max_tokens: updates.max_tokens ?? node.max_tokens,
        top_p: updates.top_p ?? node.top_p,
        top_k: updates.top_k ?? node.top_k,
        frequency_penalty: updates.frequency_penalty ?? node.frequency_penalty,
        presence_penalty: updates.presence_penalty ?? node.presence_penalty,
      };
      const updated = await updateAgentNodeConfig(nodeName, payload);
      setConfig((prev) => ({
        ...prev,
        nodes: prev.nodes.map((n) =>
          n.node_name === nodeName
            ? {
                ...n,
                model: updated.model,
                provider: updated.provider,
                temperature: updated.temperature,
                max_tokens: updated.max_tokens,
                top_p: updated.top_p,
                top_k: updated.top_k,
                frequency_penalty: updated.frequency_penalty,
                presence_penalty: updated.presence_penalty,
                is_default: updated.is_default,
              }
            : n
        ),
      }));
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving((prev) => ({ ...prev, [nodeName]: false }));
    }
  };

  const handleModelChange = (nodeName, model) => {
    handleNodeUpdate(nodeName, { model });
  };

  const toggleNodeExpanded = (nodeName) => {
    setExpandedNodes((prev) => ({ ...prev, [nodeName]: !prev[nodeName] }));
  };

  const handleReset = async (nodeName) => {
    setSaving((prev) => ({ ...prev, [nodeName]: true }));

    try {
      const updated = await resetAgentNodeConfig(nodeName);
      setConfig((prev) => ({
        ...prev,
        nodes: prev.nodes.map((n) =>
          n.node_name === nodeName
            ? {
                ...n,
                model: updated.model,
                provider: updated.provider,
                temperature: updated.temperature,
                max_tokens: updated.max_tokens,
                top_p: updated.top_p,
                top_k: updated.top_k,
                frequency_penalty: updated.frequency_penalty,
                presence_penalty: updated.presence_penalty,
                is_default: updated.is_default,
              }
            : n
        ),
      }));
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving((prev) => ({ ...prev, [nodeName]: false }));
    }
  };

  const getModelsForProvider = (provider) => {
    const models = config?.available_models;
    if (!models) return [];
    return provider === "anthropic" ? models.anthropic : models.openai;
  };

  const handleSavePreset = async () => {
    if (!presetName.trim()) return;

    setSavingPreset(true);
    setActionError(null);
    try {
      const firstNode = config?.nodes?.[0];
      const preset = {
        name: presetName.trim(),
        temperature: firstNode?.temperature,
        max_tokens: firstNode?.max_tokens,
        top_p: firstNode?.top_p,
        top_k: firstNode?.top_k,
        frequency_penalty: firstNode?.frequency_penalty,
        presence_penalty: firstNode?.presence_penalty,
      };
      const created = await createAgentConfigPreset(preset);
      setConfig({ ...config, presets: [...(config?.presets || []), created] });
      setPresetName("");
      setShowSavePreset(false);
    } catch (err) {
      setActionError(err.message);
    } finally {
      setSavingPreset(false);
    }
  };

  const handleApplyPreset = async (presetId) => {
    setSavingPreset(true);
    setActionError(null);
    try {
      const updatedConfig = await applyAgentConfigPreset(presetId);
      setConfig(updatedConfig);
    } catch (err) {
      setActionError(err.message);
    } finally {
      setSavingPreset(false);
    }
  };

  const handleApplyCostTier = async (tier) => {
    setApplyingCostTier(true);
    setActionError(null);
    try {
      const updatedConfig = await applyCostTier(tier, config?.selected_provider);
      setConfig(updatedConfig);
    } catch (err) {
      setActionError(err.message);
    } finally {
      setApplyingCostTier(false);
    }
  };

  const handleToggleRecording = async () => {
    setRecordingLoading(true);
    try {
      const action = isRecording ? stopMetricRecording : startMetricRecording;
      await action();
      setIsRecording(!isRecording);
    } catch (err) {
      setActionError(err.message);
    } finally {
      setRecordingLoading(false);
    }
  };

  const handleGlobalProviderChange = async (provider) => {
    if (!config) return;

    setChangingProvider(true);
    setActionError(null);

    try {
      const updatedConfig = await applyCostTier(config.selected_cost_tier, provider);
      setConfig(updatedConfig);
    } catch (err) {
      setActionError(err.message);
    } finally {
      setChangingProvider(false);
    }
  };

  const paramTooltips = {
    temperature: "Controls randomness. Higher values (e.g., 1.5) make output more random, lower values (e.g., 0.2) make it more focused and deterministic.",
    max_tokens: "Maximum number of tokens to generate in the response. Higher values allow longer responses.",
    top_p: "Nucleus sampling: only consider tokens with cumulative probability up to this value. Lower values make output more focused.",
    top_k: "Only sample from the top K most likely tokens. Lower values make output more focused. (Anthropic only)",
    frequency_penalty: "Reduces repetition by penalizing tokens based on how often they appear. Positive values decrease repetition. (OpenAI only)",
    presence_penalty: "Reduces repetition by penalizing tokens that have appeared at all. Positive values encourage new topics. (OpenAI only)",
  };

  return (
    <DeveloperFeature inline>
      <div className={styles.configDropdown}>
        <button
          type="button"
          className={`${styles.recordButton} ${isRecording ? styles.recording : ""}`}
          onClick={handleToggleRecording}
          disabled={recordingLoading}
          title={isRecording ? "Stop Recording" : "Start Recording"}>
          {isRecording ? <Square size={14} /> : <Circle size={14} />}
        </button>
        <button
          type="button"
          className={styles.configButton}
          onClick={() => setIsOpen(!isOpen)}
          title="Agent Model Config">
          <Settings size={16} />
        </button>
      </div>

      {/* Backdrop */}
      <div
        className={`${styles.backdrop} ${isOpen ? styles.backdropVisible : ""}`}
        onClick={() => setIsOpen(false)}
      />

      {/* Sidebar */}
      <div className={`${styles.sidebar} ${isOpen ? styles.sidebarOpen : ""}`} id={menuId}>
        <div className={styles.sidebarHeader}>
          <span className={styles.sidebarTitle}>Agent Model Configuration</span>
          <button
            type="button"
            className={styles.closeButton}
            onClick={() => setIsOpen(false)}
            title="Close">
            <X size={20} />
          </button>
        </div>

        <div className={styles.sidebarContent}>
          {/* Global Provider Section */}
          {config && !loading && (
            <div className={styles.providerSection}>
              <label className={styles.providerLabel}>Global Provider</label>
              <select
                className={styles.providerSelect}
                value={config?.selected_provider || ""}
                onChange={(e) => handleGlobalProviderChange(e.target.value)}
                disabled={changingProvider}>
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
              </select>
            </div>
          )}

          {/* Cost Tier and Presets */}
          {config && !loading && (
            <div className={styles.presetSection}>
              <div className={styles.presetRow}>
                <select
                  className={styles.presetSelect}
                  disabled={applyingCostTier || !config?.cost_tiers?.length}
                  value={config?.selected_cost_tier || ""}
                  onChange={(e) => handleApplyCostTier(e.target.value)}>
                  {config?.cost_tiers?.map((tier) => (
                    <option key={tier.tier} value={tier.tier}>
                      {tier.tier.charAt(0).toUpperCase() + tier.tier.slice(1)}
                    </option>
                  ))}
                </select>
              </div>
              <div className={styles.presetRow}>
                <select
                  className={styles.presetSelect}
                  disabled={savingPreset || !config?.presets?.length}
                  value={config?.selected_param_preset_id || ""}
                  onChange={(e) => e.target.value && handleApplyPreset(parseInt(e.target.value, 10))}>
                  <option value="">
                    {!config?.presets?.length ? "No presets" : "Presets..."}
                  </option>
                  {config?.presets?.map((preset) => (
                    <option key={preset.id} value={preset.id}>
                      {preset.name}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  className={styles.presetButton}
                  onClick={() => setShowSavePreset(!showSavePreset)}
                  title="Save current settings as preset">
                  <Save size={14} />
                </button>
              </div>
              {showSavePreset && (
                <div className={styles.savePresetRow}>
                  <input
                    type="text"
                    className={styles.presetNameInput}
                    placeholder="Preset name"
                    value={presetName}
                    onChange={(e) => setPresetName(e.target.value)}
                    onKeyDown={(e) => e.key === "Enter" && handleSavePreset()}
                  />
                  <button
                    type="button"
                    className={styles.savePresetButton}
                    onClick={handleSavePreset}
                    disabled={savingPreset || !presetName.trim()}>
                    Save
                  </button>
                </div>
              )}
            </div>
          )}

          {loading && <div className={styles.loading}>Loading...</div>}

          {error && <div className={styles.error}>{error}</div>}

          {/* Node List */}
          {config && !loading && (
            <div className={styles.nodeList}>
              {config.nodes.map((node) => (
                <div key={node.node_name} className={styles.nodeItem}>
                  <div className={styles.nodeHeader}>
                    <div className={styles.nodeInfo}>
                      <span className={styles.nodeName}>{node.display_name}</span>
                      <span className={styles.nodeDesc}>{node.description}</span>
                    </div>
                    <button
                      type="button"
                      className={styles.expandButton}
                      onClick={() => toggleNodeExpanded(node.node_name)}
                      title="Toggle settings">
                      {expandedNodes[node.node_name] ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                    </button>
                  </div>
                  <div className={styles.nodeControls}>
                    <select
                      className={styles.modelSelect}
                      value={node.model}
                      onChange={(e) => handleModelChange(node.node_name, e.target.value)}
                      disabled={saving[node.node_name]}>
                      <optgroup label="OpenAI">
                        {getModelsForProvider("openai").map((m) => (
                          <option key={m} value={m}>
                            {m}
                          </option>
                        ))}
                      </optgroup>
                      <optgroup label="Anthropic">
                        {getModelsForProvider("anthropic").map((m) => (
                          <option key={m} value={m}>
                            {m}
                          </option>
                        ))}
                      </optgroup>
                    </select>
                    {!node.is_default && (
                      <button
                        type="button"
                        className={styles.resetButton}
                        onClick={() => handleReset(node.node_name)}
                        disabled={saving[node.node_name]}
                        title="Reset to default">
                        Reset
                      </button>
                    )}
                  </div>
                  {expandedNodes[node.node_name] && (
                    <div className={styles.settingsPanel}>
                      <div className={styles.settingRow}>
                        <label className={styles.settingLabel}>
                          <span className={styles.labelWithTooltip}>
                            Temperature
                            <span className={styles.tooltipIcon} title={paramTooltips.temperature}>
                              <Info size={12} />
                            </span>
                          </span>
                          <span className={styles.settingValue}>
                            {node.temperature != null ? node.temperature.toFixed(2) : "default"}
                          </span>
                        </label>
                        <input
                          type="range"
                          min="0"
                          max="2"
                          step="0.1"
                          value={node.temperature ?? 1}
                          onChange={(e) =>
                            handleNodeUpdate(node.node_name, { temperature: parseFloat(e.target.value) })
                          }
                          disabled={saving[node.node_name]}
                          className={styles.settingSlider}
                        />
                      </div>
                      <div className={styles.settingRow}>
                        <label className={styles.settingLabel}>
                          <span className={styles.labelWithTooltip}>
                            Max Tokens
                            <span className={styles.tooltipIcon} title={paramTooltips.max_tokens}>
                              <Info size={12} />
                            </span>
                          </span>
                          <span className={styles.settingValue}>
                            {node.max_tokens != null ? node.max_tokens : "default"}
                          </span>
                        </label>
                        <input
                          type="number"
                          min="1"
                          max="16384"
                          value={node.max_tokens ?? ""}
                          placeholder="default"
                          onChange={(e) =>
                            handleNodeUpdate(node.node_name, {
                              max_tokens: e.target.value ? parseInt(e.target.value, 10) : null,
                            })
                          }
                          disabled={saving[node.node_name]}
                          className={styles.settingInput}
                        />
                      </div>
                      <div className={styles.settingRow}>
                        <label className={styles.settingLabel}>
                          <span className={styles.labelWithTooltip}>
                            Top P
                            <span className={styles.tooltipIcon} title={paramTooltips.top_p}>
                              <Info size={12} />
                            </span>
                          </span>
                          <span className={styles.settingValue}>
                            {node.top_p != null ? node.top_p.toFixed(2) : "default"}
                          </span>
                        </label>
                        <input
                          type="range"
                          min="0"
                          max="1"
                          step="0.05"
                          value={node.top_p ?? 1}
                          onChange={(e) =>
                            handleNodeUpdate(node.node_name, { top_p: parseFloat(e.target.value) })
                          }
                          disabled={saving[node.node_name]}
                          className={styles.settingSlider}
                        />
                      </div>
                      {node.provider === "anthropic" && (
                        <div className={styles.settingRow}>
                          <label className={styles.settingLabel}>
                            <span className={styles.labelWithTooltip}>
                              Top K
                              <span className={styles.tooltipIcon} title={paramTooltips.top_k}>
                                <Info size={12} />
                              </span>
                            </span>
                            <span className={styles.settingValue}>
                              {node.top_k != null ? node.top_k : "default"}
                            </span>
                          </label>
                          <input
                            type="number"
                            min="1"
                            max="500"
                            value={node.top_k ?? ""}
                            placeholder="default"
                            onChange={(e) =>
                              handleNodeUpdate(node.node_name, {
                                top_k: e.target.value ? parseInt(e.target.value, 10) : null,
                              })
                            }
                            disabled={saving[node.node_name]}
                            className={styles.settingInput}
                          />
                        </div>
                      )}
                      {node.provider === "openai" && (
                        <>
                          <div className={styles.settingRow}>
                            <label className={styles.settingLabel}>
                              <span className={styles.labelWithTooltip}>
                                Frequency Penalty
                                <span className={styles.tooltipIcon} title={paramTooltips.frequency_penalty}>
                                  <Info size={12} />
                                </span>
                              </span>
                              <span className={styles.settingValue}>
                                {node.frequency_penalty != null ? node.frequency_penalty.toFixed(2) : "default"}
                              </span>
                            </label>
                            <input
                              type="range"
                              min="-2"
                              max="2"
                              step="0.1"
                              value={node.frequency_penalty ?? 0}
                              onChange={(e) =>
                                handleNodeUpdate(node.node_name, { frequency_penalty: parseFloat(e.target.value) })
                              }
                              disabled={saving[node.node_name]}
                              className={styles.settingSlider}
                            />
                          </div>
                          <div className={styles.settingRow}>
                            <label className={styles.settingLabel}>
                              <span className={styles.labelWithTooltip}>
                                Presence Penalty
                                <span className={styles.tooltipIcon} title={paramTooltips.presence_penalty}>
                                  <Info size={12} />
                                </span>
                              </span>
                              <span className={styles.settingValue}>
                                {node.presence_penalty != null ? node.presence_penalty.toFixed(2) : "default"}
                              </span>
                            </label>
                            <input
                              type="range"
                              min="-2"
                              max="2"
                              step="0.1"
                              value={node.presence_penalty ?? 0}
                              onChange={(e) =>
                                handleNodeUpdate(node.node_name, { presence_penalty: parseFloat(e.target.value) })
                              }
                              disabled={saving[node.node_name]}
                              className={styles.settingSlider}
                            />
                          </div>
                        </>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </DeveloperFeature>
  );
};

export default AgentConfigMenu;
