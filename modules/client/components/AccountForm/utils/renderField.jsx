"use client";

import { ExternalLink } from "lucide-react";

export const inputClasses =
  "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

export const renderField = (field, value, onChange, customRender) => {
  const handleChange = (e) => {
    const newValue = field.onChange ? field.onChange(e.target.value) : e.target.value;
    onChange(field.name, newValue);
  };

  if (field.type === "custom" && customRender) {
    return customRender;
  }

  if (field.type === "textarea") {
    return (
      <textarea
        value={value || ""}
        onChange={handleChange}
        className={inputClasses}
        rows={field.rows || 3}
        placeholder={field.placeholder}
      />
    );
  }

  if (field.type === "select" && field.options) {
    return (
      <select value={value || ""} onChange={handleChange} className={inputClasses}>
        <option value="">Select...</option>
        {field.options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    );
  }

  const isUrlField = field.name.endsWith("_url") || field.name === "website";
  const hasUrl = isUrlField && value && String(value).trim();

  if (hasUrl) {
    const url = String(value).startsWith("http") ? String(value) : `https://${value}`;
    return (
      <div>
        <input
          type={field.type === "tel" ? "tel" : field.type}
          value={value || ""}
          onChange={handleChange}
          className={inputClasses}
          placeholder={field.placeholder}
          required={field.required}
          min={field.min}
          max={field.max}
        />

        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 mt-1 text-sm text-blue-600 dark:text-blue-400 hover:underline">
          {String(value)}
          <ExternalLink size={14} />
        </a>
      </div>
    );
  }

  return (
    <input
      type={field.type === "tel" ? "tel" : field.type}
      value={value || ""}
      onChange={handleChange}
      className={inputClasses}
      placeholder={field.placeholder}
      required={field.required}
      min={field.min}
      max={field.max}
    />
  );
};
