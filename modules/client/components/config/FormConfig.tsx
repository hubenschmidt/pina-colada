import { CreatedJob } from "../../types/types";
import { LeadFormConfig } from "../LeadTracker/LeadFormConfig";
import { getRecentResumeDate } from "../../api";

type LeadType = "job";

const JOB_STATUS_OPTIONS: CreatedJob["status"][] = [
  "lead",
  "applied",
  "interviewing",
  "rejected",
  "offer",
  "accepted",
  "do_not_apply",
];

const getJobFormConfig = (): LeadFormConfig<CreatedJob> => ({
  title: "Add New",
  submitButtonText: "Add",
  fields: [
    {
      name: "company",
      label: "Company",
      type: "text",
      required: true,
      placeholder: "e.g., Google",
    },
    {
      name: "job_title",
      label: "Job Title",
      type: "text",
      required: true,
      placeholder: "e.g., Software Engineer",
    },
    {
      name: "date",
      label: "Date",
      type: "date",
      required: true,
      defaultValue: new Date().toISOString().split("T")[0],
    },
    {
      name: "resume",
      label: "Resume Date",
      type: "custom",
      onInit: async () => {
        try {
          const resumeDate = await getRecentResumeDate();
          return resumeDate || "";
        } catch (error) {
          console.error("Failed to fetch recent resume date:", error);
          return "";
        }
      },
      renderCustom: ({ value, onChange }) => (
        <div>
          <input
            type="date"
            value={value || ""}
            onChange={(e) => onChange(e.target.value)}
            className="w-full px-3 py-2 border border-zinc-300 rounded focus:outline-none focus:ring-2 focus:ring-lime-500"
          />
          <label className="flex items-center gap-2 mt-2 text-sm text-zinc-600 cursor-pointer">
            <input
              type="checkbox"
              checked={!!value}
              onChange={async (e) => {
                if (e.target.checked) {
                  try {
                    const resumeDate = await getRecentResumeDate();
                    onChange(resumeDate || "");
                  } catch (error) {
                    console.error("Failed to fetch recent resume date:", error);
                  }
                } else {
                  onChange("");
                }
              }}
              className="w-4 h-4 text-lime-500 border-zinc-300 rounded focus:ring-lime-500"
            />
            <span>Use latest resume on file?</span>
          </label>
        </div>
      ),
    },
    {
      name: "salary_range",
      label: "Salary Range",
      type: "text",
      placeholder: "e.g., $120k - $150k",
    },
    {
      name: "job_url",
      label: "Job URL",
      type: "url",
      placeholder: "https://...",
      gridColumn: "md:col-span-2",
    },
    {
      name: "status",
      label: "Status",
      type: "select",
      defaultValue: "applied",
      options: JOB_STATUS_OPTIONS.map((status) => ({
        label:
          status.charAt(0).toUpperCase() + status.slice(1).replace("_", " "),
        value: status,
      })),
    },
    {
      name: "notes",
      label: "Notes",
      type: "textarea",
      placeholder: "Additional notes about this application...",
      gridColumn: "md:col-span-2",
      rows: 3,
    },
  ],
  onBeforeSubmit: (formData) => {
    // Add default values for fields not in the form
    return {
      ...formData,
      source: "manual",
    };
  },
  onValidate: (formData) => {
    if (!formData.company || !formData.job_title) {
      return {
        company: formData.company ? "" : "Company is required",
        job_title: formData.job_title ? "" : "Job Title is required",
      };
    }
    return null;
  },
});

export const useFormConfig = (type: LeadType): LeadFormConfig<any> => {
  if (type === "job") return getJobFormConfig();
  throw new Error(`Unknown lead type: ${type}`);
};
