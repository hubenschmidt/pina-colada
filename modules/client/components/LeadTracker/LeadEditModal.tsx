"use client";

import { useState, useEffect } from "react";
import { Modal } from "@mantine/core";
import { BaseLead } from "./LeadTrackerConfig";
import { LeadFormConfig, FormFieldConfig } from "./LeadFormConfig";

interface LeadEditModalProps<T extends BaseLead> {
  lead: T | null;
  opened: boolean;
  onClose: () => void;
  onUpdate: (id: string, updates: Partial<T>) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  config: LeadFormConfig<T>;
}

const LeadEditModal = <T extends BaseLead>({
  lead,
  opened,
  onClose,
  onUpdate,
  onDelete,
  config,
}: LeadEditModalProps<T>) => {
  const [formData, setFormData] = useState<Partial<T>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  useEffect(() => {
    if (lead) {
      const data: any = {};
      config.fields.forEach((field) => {
        const value = (lead as any)[field.name];
        if (field.type === "date" && value) {
          data[field.name] = new Date(value).toISOString().split("T")[0];
        } else {
          data[field.name] = value ?? "";
        }
      });
      setFormData(data);
      setErrors({});
      setIsDeleting(false);
    }
  }, [lead, config.fields]);

  const handleFieldChange = (fieldName: string, value: any) => {
    const field = config.fields.find((f) => f.name === fieldName);

    let processedValue = value;
    if (field?.onChange) {
      processedValue = field.onChange(value, formData);
    }

    setFormData((prev) => ({ ...prev, [fieldName]: processedValue }));

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

    config.fields.forEach((field) => {
      const value = (formData as any)[field.name];

      if (field.required && (!value || value === "")) {
        newErrors[String(field.name)] = `${field.label} is required`;
      }

      if (field.validation && value) {
        const error = field.validation(value);
        if (error) {
          newErrors[String(field.name)] = error;
        }
      }
    });

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

    if (!lead || !validate()) {
      return;
    }

    // Clean up data: convert empty strings to null for optional fields
    const cleanedData: any = { ...formData };
    config.fields.forEach((field) => {
      if (!field.required && cleanedData[field.name] === "") {
        cleanedData[field.name] = null;
      }
    });

    setIsSubmitting(true);
    try {
      let submitData = cleanedData;
      if (config.onBeforeSubmit) {
        submitData = config.onBeforeSubmit(submitData);
      }

      console.log("[LeadEditModal] Submitting update:", { leadId: lead.id, submitData });
      await onUpdate(lead.id, submitData);
      console.log("[LeadEditModal] Update successful");
      onClose();
    } catch (error) {
      console.error("[LeadEditModal] Failed to update:", error);
      alert("Failed to update. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!lead) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await onDelete(lead.id);
      onClose();
    } catch (error) {
      console.error("Failed to delete:", error);
      alert("Failed to delete. Please try again.");
      setIsDeleting(false);
    }
  };

  const renderField = (field: FormFieldConfig<T>) => {
    if (field.hidden) return null;

    const value = (formData as any)[field.name];
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

  if (!lead) return null;

  return (
    <Modal.Root opened={opened} onClose={onClose} size="lg">
      <Modal.Overlay backgroundOpacity={0.5} />
      <Modal.Content>
        <Modal.Header>
          <Modal.Title>{config.title.replace("Add New", "Edit")}</Modal.Title>
          <Modal.CloseButton />
        </Modal.Header>
        <Modal.Body>
          <form onSubmit={handleSubmit}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {config.fields.map((field) => renderField(field))}
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
                  {config.cancelButtonText || "Cancel"}
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
};

export default LeadEditModal;
