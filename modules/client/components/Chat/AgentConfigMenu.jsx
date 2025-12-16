import React, { useState, useEffect, useId } from "react";
import { Settings } from "lucide-react";
import {
  getAgentConfig,
  getAvailableModels,
  updateAgentNodeConfig,
  resetAgentNodeConfig,
} from "../../api";
import { useUserContext } from "../../context/userContext";
import DeveloperFeature from "../DeveloperFeature/DeveloperFeature";
import styles from "./AgentConfigMenu.module.css";

const AgentConfigMenu = () => {
  const { userState } = useUserContext();
  const [isOpen, setIsOpen] = useState(false);
  const [config, setConfig] = useState(null);
  const [availableModels, setAvailableModels] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState({});
  const menuId = useId();

  useEffect(() => {
    if (!isOpen || !userState.roles?.includes("developer")) return;

    setLoading(true);
    setError(null);

    Promise.all([getAgentConfig(), getAvailableModels()])
      .then(([configRes, modelsRes]) => {
        setConfig(configRes);
        setAvailableModels(modelsRes);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [isOpen, userState.roles]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      const menu = document.getElementById(menuId);
      if (menu && !menu.contains(event.target)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      return () => document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [isOpen, menuId]);

  const handleModelChange = async (nodeName, model) => {
    setSaving((prev) => ({ ...prev, [nodeName]: true }));

    try {
      const updated = await updateAgentNodeConfig(nodeName, model);
      setConfig((prev) => ({
        ...prev,
        nodes: prev.nodes.map((n) =>
          n.node_name === nodeName
            ? { ...n, model: updated.model, provider: updated.provider, is_default: updated.is_default }
            : n
        ),
      }));
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving((prev) => ({ ...prev, [nodeName]: false }));
    }
  };

  const handleReset = async (nodeName) => {
    setSaving((prev) => ({ ...prev, [nodeName]: true }));

    try {
      const updated = await resetAgentNodeConfig(nodeName);
      setConfig((prev) => ({
        ...prev,
        nodes: prev.nodes.map((n) =>
          n.node_name === nodeName
            ? { ...n, model: updated.model, provider: updated.provider, is_default: updated.is_default }
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
    if (!availableModels) return [];
    return provider === "anthropic" ? availableModels.anthropic : availableModels.openai;
  };

  return (
    <DeveloperFeature inline>
      <div className={styles.configDropdown} id={menuId}>
        <button
          type="button"
          className={styles.configButton}
          onClick={() => setIsOpen(!isOpen)}
          title="Agent Model Config">
          <Settings size={16} />
        </button>

        {isOpen && (
          <div className={styles.configMenu}>
            <div className={styles.menuHeader}>Agent Model Configuration</div>

            {loading && <div className={styles.loading}>Loading...</div>}

            {error && <div className={styles.error}>{error}</div>}

            {config && !loading && (
              <div className={styles.nodeList}>
                {config.nodes.map((node) => (
                  <div key={node.node_name} className={styles.nodeItem}>
                    <div className={styles.nodeInfo}>
                      <span className={styles.nodeName}>{node.display_name}</span>
                      <span className={styles.nodeDesc}>{node.description}</span>
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
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </DeveloperFeature>
  );
};

export default AgentConfigMenu;
