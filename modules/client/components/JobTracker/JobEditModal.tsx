"use client";

import { useState, useEffect } from "react";
import { Modal } from "@mantine/core";
import { CreatedJob } from "../../types/types";
import * as api from "../../api";

interface JobEditModalProps {
  job: CreatedJob | null;
  opened: boolean;
  onClose: () => void;
  onUpdate: () => Promise<void>;
  onDelete: () => Promise<void>;
}

export function JobEditModal({
  job,
  opened,
  onClose,
  onUpdate,
  onDelete,
}: JobEditModalProps) {
  const [formData, setFormData] = useState({
    company: "",
    job_title: "",
    date: "",
    status: "Applied" as CreatedJob["status"],
    job_url: "",
    notes: "",
    resume: "",
    salary_range: "",
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  // Populate form when job changes
  useEffect(() => {
    if (job) {
      setFormData({
        company: job.company,
        job_title: job.job_title,
        date: job.date ? new Date(job.date).toISOString().split("T")[0] : "",
        status: job.status,
        job_url: job.job_url || "",
        notes: job.notes || "",
        resume: job.resume || "",
        salary_range: job.salary_range || "",
      });
      setErrors({});
      setIsDeleting(false);
    }
  }, [job]);

  const handleChange = (
    field: keyof typeof formData,
    value: string
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  const validate = (): boolean => {
    const newErrors: { [key: string]: string } = {};

    if (!formData.company.trim()) {
      newErrors.company = "Company is required";
    }
    if (!formData.job_title.trim()) {
      newErrors.job_title = "Job title is required";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!job || !validate()) return;

    setIsSubmitting(true);
    try {
      await api.updateJob(job.id, {
        ...formData,
        job_url: formData.job_url || null,
        notes: formData.notes || null,
        resume: formData.resume || null,
        salary_range: formData.salary_range || null,
      });
      await onUpdate();
      onClose();
    } catch (error) {
      console.error("Failed to update job:", error);
      alert("Failed to update job. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!job) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await api.deleteJob(job.id);
      await onDelete();
      onClose();
    } catch (error) {
      console.error("Failed to delete job:", error);
      alert("Failed to delete job. Please try again.");
      setIsDeleting(false);
    }
  };

  if (!job) return null;

  return (
    <Modal.Root opened={opened} onClose={onClose} size="lg">
      <Modal.Overlay backgroundOpacity={0.5} />
      <Modal.Content>
        <Modal.Header>
          <Modal.Title>Edit Job</Modal.Title>
          <Modal.CloseButton />
        </Modal.Header>
        <Modal.Body>
          <form onSubmit={handleSubmit}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Company */}
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Company <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.company}
                  onChange={(e) => handleChange("company", e.target.value)}
                  className={`w-full px-3 py-2 border ${
                    errors.company ? "border-red-500" : "border-zinc-300"
                  } rounded focus:outline-none focus:ring-2 focus:ring-lime-500`}
                />
                {errors.company && (
                  <p className="text-red-500 text-xs mt-1">{errors.company}</p>
                )}
              </div>

              {/* Job Title */}
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Job Title <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.job_title}
                  onChange={(e) => handleChange("job_title", e.target.value)}
                  className={`w-full px-3 py-2 border ${
                    errors.job_title ? "border-red-500" : "border-zinc-300"
                  } rounded focus:outline-none focus:ring-2 focus:ring-lime-500`}
                />
                {errors.job_title && (
                  <p className="text-red-500 text-xs mt-1">{errors.job_title}</p>
                )}
              </div>

              {/* Date */}
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Date
                </label>
                <input
                  type="date"
                  value={formData.date}
                  onChange={(e) => handleChange("date", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              {/* Status */}
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Status
                </label>
                <select
                  value={formData.status}
                  onChange={(e) =>
                    handleChange("status", e.target.value as CreatedJob["status"])
                  }
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                >
                  <option value="Lead">Lead</option>
                  <option value="Applied">Applied</option>
                  <option value="Interviewing">Interviewing</option>
                  <option value="Rejected">Rejected</option>
                  <option value="Offer">Offer</option>
                  <option value="Accepted">Accepted</option>
                  <option value="Do Not Apply">Do Not Apply</option>
                </select>
              </div>

              {/* Job URL */}
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Job URL
                </label>
                <input
                  type="url"
                  value={formData.job_url}
                  onChange={(e) => handleChange("job_url", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              {/* Salary Range */}
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Salary Range
                </label>
                <input
                  type="text"
                  value={formData.salary_range}
                  onChange={(e) => handleChange("salary_range", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              {/* Resume */}
              <div>
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Resume Version
                </label>
                <input
                  type="text"
                  value={formData.resume}
                  onChange={(e) => handleChange("resume", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                />
              </div>

              {/* Notes */}
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-zinc-700 mb-1">
                  Notes
                </label>
                <textarea
                  value={formData.notes}
                  onChange={(e) => handleChange("notes", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
                  rows={3}
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6 justify-between">
              <button
                type="button"
                onClick={handleDelete}
                className={
                  isDeleting
                    ? "px-6 py-3 rounded-lg font-semibold bg-red-600 text-white hover:bg-red-700"
                    : "px-6 py-3 rounded-lg font-semibold bg-red-100 text-red-700 hover:bg-red-200"
                }
              >
                {isDeleting ? "Click again to confirm delete" : "Delete"}
              </button>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={onClose}
                  className="px-6 py-3 bg-zinc-200 text-zinc-700 rounded-lg hover:bg-zinc-300 font-semibold"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-6 py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 rounded-lg hover:brightness-95 font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isSubmitting ? "Saving..." : "Save Changes"}
                </button>
              </div>
            </div>
          </form>
        </Modal.Body>
      </Modal.Content>
    </Modal.Root>
  );
}
