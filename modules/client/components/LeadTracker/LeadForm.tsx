"use client";

import { useState, useEffect } from "react";
import { Plus, X } from "lucide-react";
import { BaseLead } from "./LeadTrackerConfig";
import { LeadFormConfig, FormFieldConfig } from "./LeadFormConfig";

interface LeadFormProps<T extends BaseLead> {
  isOpen: boolean;
  onClose: () => void;
  onAdd: (lead: Omit<T, "id" | "created_at" | "updated_at">) => Promise<void>;
  config: LeadFormConfig<T>;
}

const LeadForm = <T extends BaseLead>({
  isOpen,
  onClose,
  onAdd,
  config,
}: LeadFormProps<T>) => {
  const [formData, setFormData] = useState<any>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  // Initialize form data with defaults
  useEffect(() => {
    if (isOpen) {
      const initialData: any = {};
      config.fields.forEach((field) => {
        if (field.defaultValue !== undefined) {
          initialData[field.name] = field.defaultValue;
        } else if (field.type === "date") {
          initialData[field.name] = new Date().toISOString().split("T")[0];
        } else if (field.type === "checkbox") {
          initialData[field.name] = false;
        } else {
          initialData[field.name] = "";
        }
      });
      setFormData(initialData);
      setErrors({});

      // Run onInit for fields that have it
      config.fields.forEach(async (field) => {
        if (field.onInit) {
          try {
            const value = await field.onInit();
            setFormData((prev: any) => ({ ...prev, [field.name]: value }));
          } catch (error) {
            console.error(`Failed to initialize field ${String(field.name)}:`, error);
          }
        }
      });
    }
  }, [isOpen, config.fields]);

  const handleFieldChange = (fieldName: string, value: any) => {
    const field = config.fields.find((f) => f.name === fieldName);

    let processedValue = value;
    if (field?.onChange) {
      processedValue = field.onChange(value, formData);
    }

    setFormData((prev: any) => ({ ...prev, [fieldName]: processedValue }));

    // Clear error for this field
    if (errors[fieldName]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[fieldName];
        return newErrors;
      });
    }
  };

  const validate = (): boolean => {
    const newErrors: { [key: string]: string } = {};

    // Field-level validation
    config.fields.forEach((field) => {
      const value = formData[field.name];

      // Required check
      if (field.required && (!value || value === "")) {
        newErrors[String(field.name)] = `${field.label} is required`;
      }

      // Custom validation
      if (field.validation && value) {
        const error = field.validation(value);
        if (error) {
          newErrors[String(field.name)] = error;
        }
      }
    });

    // Form-level validation
    if (config.onValidate) {
      const formErrors = config.onValidate(formData);
      if (formErrors) {
        Object.assign(newErrors, formErrors);
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    setIsSubmitting(true);
    try {
      let submitData = { ...formData };

      // Pre-process data if needed
      if (config.onBeforeSubmit) {
        submitData = config.onBeforeSubmit(submitData);
      }

      await onAdd(submitData);

      // Reset form
      const resetData: any = {};
      config.fields.forEach((field) => {
        if (field.defaultValue !== undefined) {
          resetData[field.name] = field.defaultValue;
        } else if (field.type === "date") {
          resetData[field.name] = new Date().toISOString().split("T")[0];
        } else if (field.type === "checkbox") {
          resetData[field.name] = false;
        } else {
          resetData[field.name] = "";
        }
      });
      setFormData(resetData);
      setErrors({});
      onClose();
    } catch (error: any) {
      console.error("Failed to add lead:", error);
      const errorMessage =
        error?.message || error?.error || "Failed to add. Please try again.";
      alert(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  const renderField = (field: FormFieldConfig<T>) => {
    if (field.hidden) return null;

    const value = formData[field.name];
    const error = errors[String(field.name)];
    const inputClasses = `w-full px-3 py-2 border ${
      error ? "border-red-500" : "border-zinc-300"
    } rounded focus:outline-none focus:ring-2 focus:ring-lime-500`;

    const fieldWrapper = (content: React.ReactNode) => (
      <div className={field.gridColumn || ""} key={String(field.name)}>
        <label className="block text-sm font-medium text-zinc-700 mb-1">
          {field.label}{" "}
          {field.required && <span className="text-red-500">*</span>}
        </label>
        {content}
        {error && <p className="text-red-500 text-xs mt-1">{error}</p>}
      </div>
    );

    if (field.type === "custom" && field.renderCustom) {
      return fieldWrapper(
        field.renderCustom({
          value,
          onChange: (v) => handleFieldChange(String(field.name), v),
          field,
        })
      );
    }

    if (field.type === "textarea") {
      return fieldWrapper(
        <textarea
          value={value || ""}
          onChange={(e) => handleFieldChange(String(field.name), e.target.value)}
          className={inputClasses}
          placeholder={field.placeholder}
          rows={field.rows || 3}
          required={field.required}
          disabled={field.disabled}
        />
      );
    }

    if (field.type === "select") {
      return fieldWrapper(
        <select
          value={value || ""}
          onChange={(e) => handleFieldChange(String(field.name), e.target.value)}
          className={inputClasses}
          required={field.required}
          disabled={field.disabled}
        >
          {field.options?.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      );
    }

    if (field.type === "checkbox") {
      return fieldWrapper(
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={value || false}
            onChange={(e) =>
              handleFieldChange(String(field.name), e.target.checked)
            }
            className="w-4 h-4 text-lime-500 border-zinc-300 rounded focus:ring-lime-500"
            disabled={field.disabled}
          />
          <span className="text-sm text-zinc-600">{field.placeholder}</span>
        </label>
      );
    }

    return fieldWrapper(
      <input
        type={field.type}
        value={value || ""}
        onChange={(e) => handleFieldChange(String(field.name), e.target.value)}
        className={inputClasses}
        placeholder={field.placeholder}
        required={field.required}
        min={field.min}
        max={field.max}
        step={field.step}
        pattern={field.pattern}
        disabled={field.disabled}
      />
    );
  };

  if (!isOpen) return null;

  return (
    <div className="border border-zinc-300 rounded-lg p-6 bg-blue-50 shadow-lg">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-xl font-semibold text-blue-900">{config.title}</h3>
        <button
          onClick={onClose}
          className="p-2 text-zinc-600 hover:bg-zinc-200 rounded"
        >
          <X size={20} />
        </button>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {config.fields.map((field) => renderField(field))}
        </div>

        <div className="flex gap-3 mt-6">
          <button
            type="submit"
            disabled={isSubmitting}
            className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-lime-500 to-yellow-400 text-blue-900 rounded-lg hover:brightness-95 font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Plus size={18} />
            {isSubmitting
              ? "Adding..."
              : config.submitButtonText || "Add"}
          </button>
          <button
            type="button"
            onClick={onClose}
            className="px-6 py-3 bg-zinc-200 text-zinc-700 rounded-lg hover:bg-zinc-300 font-semibold"
          >
            {config.cancelButtonText || "Cancel"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default LeadForm;
