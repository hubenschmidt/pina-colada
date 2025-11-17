"use client";

import { useState, useEffect } from "react";
import { Plus, X } from "lucide-react";
import { CreatedJob } from "../../types/types";
import * as api from "../../api";

interface JobFormProps {
  isOpen: boolean;
  onClose: () => void;
  onAdd: () => Promise<void>;
}

export function JobForm({ isOpen, onClose, onAdd }: JobFormProps) {
  const [formData, setFormData] = useState({
    company: "",
    job_title: "",
    date: new Date().toISOString().split("T")[0],
    status: "Applied" as CreatedJob["status"],
    job_url: "",
    notes: "",
    resume: "",
    salary_range: "",
    source: "manual" as const,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});
  const [useLatestResume, setUseLatestResume] = useState(true);

  // Reset form when opened
  useEffect(() => {
    if (isOpen) {
      const loadResumeDate = async () => {
        try {
          const resumeDate = await api.getRecentResumeDate();
          setFormData({
            company: "",
            job_title: "",
            date: new Date().toISOString().split("T")[0],
            status: "Applied",
            job_url: "",
            notes: "",
            resume: resumeDate || "",
            salary_range: "",
            source: "manual",
          });
        } catch (error) {
          console.error("Failed to fetch recent resume date:", error);
          setFormData({
            company: "",
            job_title: "",
            date: new Date().toISOString().split("T")[0],
            status: "Applied",
            job_url: "",
            notes: "",
            resume: "",
            salary_range: "",
            source: "manual",
          });
        }
      };
      loadResumeDate();
      setUseLatestResume(true);
      setErrors({});
    }
  }, [isOpen]);

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

    if (!validate()) return;

    setIsSubmitting(true);
    try {
      await api.createJob({
        ...formData,
        job_url: formData.job_url || null,
        notes: formData.notes || null,
        resume: formData.resume || null,
        salary_range: formData.salary_range || null,
        lead_status_id: null,
      });
      await onAdd();
      onClose();
    } catch (error: any) {
      console.error("Failed to add job:", error);
      alert(error?.message || "Failed to add job. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="border border-zinc-300 rounded-lg p-6 bg-blue-50 shadow-lg">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-xl font-semibold text-blue-900">Add New Job</h3>
        <button
          onClick={onClose}
          className="p-2 text-zinc-600 hover:bg-zinc-200 rounded"
        >
          <X size={20} />
        </button>
      </div>

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
              placeholder="e.g. Acme Corp"
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
              placeholder="e.g. Software Engineer"
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
              type="text"
              value={formData.job_url}
              onChange={(e) => handleChange("job_url", e.target.value)}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
              placeholder="https://..."
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
              placeholder="e.g. $100k - $150k"
            />
          </div>

          {/* Resume */}
          <div>
            <label className="block text-sm font-medium text-zinc-700 mb-1">
              Resume Date
            </label>
            <input
              type="date"
              value={formData.resume ? formData.resume.split("T")[0] : ""}
              onChange={(e) => handleChange("resume", e.target.value)}
              className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
            />
            <label className="flex items-center gap-2 mt-2 text-sm text-zinc-600 cursor-pointer">
              <input
                type="checkbox"
                checked={useLatestResume}
                onChange={async (e) => {
                  setUseLatestResume(e.target.checked);
                  if (!e.target.checked) {
                    handleChange("resume", "");
                    return;
                  }
                  try {
                    const resumeDate = await api.getRecentResumeDate();
                    handleChange("resume", resumeDate || "");
                  } catch (error) {
                    console.error("Failed to fetch recent resume date:", error);
                  }
                }}
                className="w-4 h-4 text-lime-500 border-zinc-300 rounded focus:ring-lime-500"
              />
              <span>Use latest resume on file?</span>
            </label>
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
              placeholder="Additional notes..."
              rows={3}
            />
          </div>
        </div>

        <div className="flex gap-3 mt-6">
          <button
            type="submit"
            disabled={isSubmitting}
            className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 rounded-lg hover:brightness-95 font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus size={18} />
            {isSubmitting ? "Adding..." : "Add Job"}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-6 py-3 bg-zinc-200 text-zinc-700 rounded-lg hover:bg-zinc-300 font-semibold"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
