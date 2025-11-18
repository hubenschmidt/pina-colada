"use client";

import { useEffect, useState } from "react";
import { X, ExternalLink, CheckCircle, Filter } from "lucide-react";
import { JobWithLeadStatus, LeadStatus } from "../../api";
import * as api from "../../api";
import styles from "./JobPanel.module.css";

interface JobPanelProps {
  isOpen: boolean;
  onClose: () => void;
  onJobsChange?: () => void;
}

const JobPanel = ({ isOpen, onClose, onJobsChange }: JobPanelProps) => {
  const [jobs, setJobs] = useState<JobWithLeadStatus[]>([]);
  const [leadStatuses, setLeadStatuses] = useState<LeadStatus[]>([]);
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showToast, setShowToast] = useState(false);
  const [toastMessage, setToastMessage] = useState("");

  useEffect(() => {
    if (isOpen) {
      loadLeadStatuses();
      loadJobs();
    }
  }, [isOpen, selectedStatuses]);

  const loadLeadStatuses = async () => {
    try {
      const statuses = await api.getStatuses();
      setLeadStatuses(statuses);
    } catch (err) {
      console.error("Error loading lead statuses:", err);
    }
  };

  const loadJobs = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getLeads(
        selectedStatuses.length > 0
          ? (selectedStatuses as Array<"Qualifying" | "Cold" | "Warm" | "Hot">)
          : undefined
      );
      setJobs(data);
    } catch (err) {
      console.error("Error loading job leads:", err);
      setError("Failed to load job leads");
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

  const handleJobStatusChange = async (jobId: string, newStatusId: string) => {
    try {
      await api.updateJob(jobId, { lead_status_id: newStatusId });
      await loadJobs();
      onJobsChange?.();
      showSuccessToast("Job status updated");
    } catch (err) {
      console.error("Error updating job status:", err);
      alert("Failed to update job status");
    }
  };

  const handleMarkAsApplied = async (job: JobWithLeadStatus) => {
    try {
      await api.markLeadAsApplied(job.id);
      await loadJobs();
      onJobsChange?.();
      showSuccessToast("Marked as applied");
    } catch (err) {
      console.error("Error marking as applied:", err);
      alert("Failed to mark as applied");
    }
  };

  const handleMarkAsDoNotApply = async (job: JobWithLeadStatus) => {
    try {
      await api.markLeadAsDoNotApply(job.id);
      await loadJobs();
      onJobsChange?.();
      showSuccessToast("Marked as do not apply");
    } catch (err) {
      console.error("Error marking as do not apply:", err);
      alert("Failed to mark as do not apply");
    }
  };

  const getStatusBadgeClass = (statusName: string): string => {
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
            <h2 className={styles.title}>Job Leads</h2>
            <span className={styles.count}>({jobs.length})</span>
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
          {loading && <div className={styles.loading}>Loading job leads...</div>}

          {error && <div className={styles.error}>{error}</div>}

          {!loading && !error && jobs.length === 0 && (
            <div className={styles.empty}>
              No job leads found. Try adjusting your filters.
            </div>
          )}

          {!loading && !error && jobs.length > 0 && (
            <div className={styles.jobsList}>
              {jobs.map((job) => (
                <div key={job.id} className={styles.jobCard}>
                  <div className={styles.jobHeader}>
                    <div>
                      <h3 className={styles.jobCompany}>{job.company}</h3>
                      <p className={styles.jobTitle}>{job.job_title}</p>
                    </div>
                    {job.lead_status && (
                      <select
                        value={job.lead_status_id || ""}
                        onChange={(e) =>
                          handleJobStatusChange(job.id, e.target.value)
                        }
                        className={`${
                          styles.statusDropdown
                        } ${getStatusBadgeClass(job.lead_status.name)}`}
                      >
                        {leadStatuses.map((status) => (
                          <option key={status.id} value={status.id}>
                            {status.name}
                          </option>
                        ))}
                      </select>
                    )}
                  </div>

                  {job.job_url && (
                    <a
                      href={job.job_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={styles.jobUrl}
                    >
                      <ExternalLink size={14} />
                      <span>View posting</span>
                    </a>
                  )}

                  <div className={styles.jobActions}>
                    <button
                      onClick={() => handleMarkAsApplied(job)}
                      className={`${styles.actionButton} ${styles.appliedButton}`}
                      title="Mark as Applied"
                    >
                      <CheckCircle size={16} />
                      <span>Applied</span>
                    </button>
                    <button
                      onClick={() => handleMarkAsDoNotApply(job)}
                      className={`${styles.actionButton} ${styles.doNotApplyButton}`}
                      title="Mark as Do Not Apply"
                    >
                      <X size={16} />
                      <span>Do Not Apply</span>
                    </button>
                  </div>
                </div>
              ))}
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
};

export default JobPanel;
