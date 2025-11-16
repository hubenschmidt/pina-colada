"use client";

import { useEffect, useState } from "react";
import { X, ExternalLink, CheckCircle, Filter } from "lucide-react";
import { BaseLead } from "./LeadTrackerConfig";
import {
  LeadPanelConfig,
  LeadWithStatus,
  LeadStatus,
} from "./LeadPanelConfig";
import styles from "./LeadPanel.module.css";

interface LeadPanelProps<T extends BaseLead> {
  isOpen: boolean;
  onClose: () => void;
  onLeadsChange?: () => void;
  config: LeadPanelConfig<T>;
}

function LeadPanel<T extends BaseLead>({
  isOpen,
  onClose,
  onLeadsChange,
  config,
}: LeadPanelProps<T>) {
  const [leads, setLeads] = useState<LeadWithStatus<T>[]>([]);
  const [leadStatuses, setLeadStatuses] = useState<LeadStatus[]>([]);
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>(
    config.defaultStatusFilter || []
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showToast, setShowToast] = useState(false);
  const [toastMessage, setToastMessage] = useState("");

  useEffect(() => {
    loadLeadStatuses();
  }, []);

  useEffect(() => {
    if (isOpen) {
      loadLeads();
    }
  }, [isOpen, selectedStatuses]);

  const loadLeadStatuses = async () => {
    try {
      const statuses = await config.api.getStatuses();
      setLeadStatuses(statuses);
    } catch (err) {
      console.error("Error loading lead statuses:", err);
    }
  };

  const loadLeads = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await config.api.getLeads(
        selectedStatuses.length > 0 ? selectedStatuses : undefined
      );
      setLeads(data);
    } catch (err) {
      console.error("Error loading leads:", err);
      setError(`Failed to load ${config.entityNamePlural}`);
    } finally {
      setLoading(false);
    }
  };

  const showSuccessToast = (message: string) => {
    setToastMessage(message);
    setShowToast(true);
    setTimeout(() => setShowToast(false), 3000);
  };

  const handleStatusFilterChange = (statusName: string) => {
    setSelectedStatuses((prev) => {
      if (prev.includes(statusName)) {
        return prev.filter((s) => s !== statusName);
      } else {
        return [...prev, statusName];
      }
    });
  };

  const handleSelectAllStatuses = () => {
    if (selectedStatuses.length === leadStatuses.length) {
      setSelectedStatuses([]);
    } else {
      setSelectedStatuses(leadStatuses.map((s) => s.name));
    }
  };

  const handleLeadStatusChange = async (
    leadId: string,
    newStatusId: string
  ) => {
    try {
      await config.api.updateLeadStatus(leadId, newStatusId);
      await loadLeads();
      onLeadsChange?.();
      showSuccessToast("Lead status updated");
    } catch (err) {
      console.error("Error updating lead status:", err);
      alert("Failed to update lead status");
    }
  };

  const handleAction = async (
    action: (typeof config.actions)[0],
    lead: LeadWithStatus<T>
  ) => {
    try {
      await action.onClick(lead);
      await loadLeads();
      onLeadsChange?.();
      showSuccessToast(`${action.label} completed`);
    } catch (err) {
      console.error(`Error performing ${action.label}:`, err);
      alert(`Failed to ${action.label.toLowerCase()}`);
    }
  };

  const getStatusBadgeClass = (statusName: string): string => {
    if (config.getStatusBadgeClass) {
      return config.getStatusBadgeClass(statusName);
    }
    // Default styling
    switch (statusName.toLowerCase()) {
      case "qualifying":
        return styles.badgeQualifying;
      case "cold":
        return styles.badgeCold;
      case "warm":
        return styles.badgeWarm;
      case "hot":
        return styles.badgeHot;
      default:
        return styles.badgeQualifying;
    }
  };

  if (!isOpen) return null;

  return (
    <>
      <div className={styles.backdrop} onClick={onClose} />

      <div className={styles.panel}>
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <h2 className={styles.title}>{config.title}</h2>
            <span className={styles.count}>({leads.length})</span>
          </div>
          <button
            onClick={onClose}
            className={styles.closeButton}
            aria-label="Close panel"
          >
            <X size={24} />
          </button>
        </div>

        <div className={styles.filterSection}>
          <div className={styles.filterHeader}>
            <Filter size={16} />
            <span>Filter by Status</span>
          </div>
          <div className={styles.filterOptions}>
            <label className={styles.filterOption}>
              <input
                type="checkbox"
                checked={selectedStatuses.length === leadStatuses.length}
                onChange={handleSelectAllStatuses}
              />
              <span>All</span>
            </label>
            {leadStatuses.map((status) => (
              <label key={status.id} className={styles.filterOption}>
                <input
                  type="checkbox"
                  checked={selectedStatuses.includes(status.name)}
                  onChange={() => handleStatusFilterChange(status.name)}
                />
                <span className={getStatusBadgeClass(status.name)}>
                  {status.name}
                </span>
              </label>
            ))}
          </div>
        </div>

        <div className={styles.content}>
          {loading && (
            <div className={styles.loading}>
              Loading {config.entityNamePlural}...
            </div>
          )}

          {error && <div className={styles.error}>{error}</div>}

          {!loading && !error && leads.length === 0 && (
            <div className={styles.empty}>
              {config.emptyMessage ||
                `No ${config.entityNamePlural} found. Try adjusting your filters.`}
            </div>
          )}

          {!loading && !error && leads.length > 0 && (
            <div className={styles.leadsList}>
              {leads.map((lead) => {
                const companyName = config.getCompanyName(lead);
                const title = config.getTitle(lead);
                const url = config.getUrl ? config.getUrl(lead) : null;

                return (
                  <div key={lead.id} className={styles.leadCard}>
                    <div className={styles.leadHeader}>
                      <div>
                        <h3 className={styles.leadCompany}>{companyName}</h3>
                        <p className={styles.leadTitle}>{title}</p>
                      </div>
                      {lead.lead_status && (
                        <select
                          value={lead.lead_status_id || ""}
                          onChange={(e) =>
                            handleLeadStatusChange(lead.id, e.target.value)
                          }
                          className={`${
                            styles.statusDropdown
                          } ${getStatusBadgeClass(lead.lead_status.name)}`}
                        >
                          {leadStatuses.map((status) => (
                            <option key={status.id} value={status.id}>
                              {status.name}
                            </option>
                          ))}
                        </select>
                      )}
                    </div>

                    {url && (
                      <a
                        href={url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.leadUrl}
                      >
                        <ExternalLink size={14} />
                        <span>View posting</span>
                      </a>
                    )}

                    <div className={styles.leadActions}>
                      {config.actions.map((action, index) => {
                        const Icon = action.icon;
                        const variantClass =
                          action.variant === "primary"
                            ? styles.appliedButton
                            : action.variant === "danger"
                            ? styles.removeButton
                            : action.variant === "success"
                            ? styles.appliedButton
                            : styles.doNotApplyButton;

                        return (
                          <button
                            key={index}
                            onClick={() => handleAction(action, lead)}
                            className={`${styles.actionButton} ${variantClass}`}
                            title={action.label}
                          >
                            <Icon size={16} />
                            {action.showLabel !== false && (
                              <span>{action.label}</span>
                            )}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {showToast && (
        <div className={styles.toast}>
          <CheckCircle size={16} />
          <span>{toastMessage}</span>
        </div>
      )}
    </>
  );
}

export default LeadPanel;
