"use client";

import { useState, useEffect } from "react";
import ContactSection from "../ContactSection/ContactSection";
import RelationshipsSection from "../RelationshipsSection/RelationshipsSection";
import NotesSection from "../NotesSection/NotesSection";
import TasksSection from "../TasksSection/TasksSection";
import CommentsSection from "../CommentsSection/CommentsSection";
import DocumentsSection from "../DocumentsSection/DocumentsSection";
import FormActions from "../FormActions/FormActions";
import Timestamps from "../Timestamps/Timestamps";
import {
  createIndividualContact,
  deleteIndividualContact,
  updateIndividualContact,
  createOrganizationContact,
  deleteOrganizationContact,
  updateOrganizationContact,
  searchContacts,
  searchAccounts,
  createNote,
  createTask,
  getIndustries,
  linkDocumentToEntity,
} from "../../api";
import { useProjectContext } from "../../context/projectContext";

import { emptyPendingContact } from "./types/AccountFormTypes";
import { useAccountFormConfig } from "./hooks/useAccountFormConfig";
import { usePendingChanges } from "../../hooks/usePendingChanges";
import IndustrySelector from "./IndustrySelector";
import EmployeeCountRangeSelector from "./EmployeeCountRangeSelector";
import FundingStageSelector from "./FundingStageSelector";
import RevenueRangeSelector from "./RevenueRangeSelector";
import ProjectSelector from "./ProjectSelector";
import { renderField } from "./utils/renderField";

const AccountForm = ({ type, onClose, onAdd, account, onUpdate, onDelete }) => {
  const config = useAccountFormConfig(type);
  const { projectState } = useProjectContext();
  const isEditMode = !!account;
  const isOrganization = type === "organization";

  const [formData, setFormData] = useState({});
  const [contacts, setContacts] = useState([]);
  const [pendingContacts, setPendingContacts] = useState([]);
  const [pendingDeletions, setPendingDeletions] = useState([]);
  const [pendingNotes, setPendingNotes] = useState([]);
  const [pendingTasks, setPendingTasks] = useState([]);
  const [pendingComments, setPendingComments] = useState([]);
  const [pendingDocumentIds, setPendingDocumentIds] = useState([]);
  const [relationships, setRelationships] = useState([]);
  const [pendingRelationships, setPendingRelationships] = useState([]);
  const [selectedIndustries, setSelectedIndustries] = useState([]);
  const [selectedProjectIds, setSelectedProjectIds] = useState([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState({});

  useEffect(() => {
    if (isEditMode && account) {
      setFormData({ ...account });
      setContacts(account.contacts || []);
      setSelectedIndustries(account.industries || []);
      const projectIds =
        account.projects?.map((p) => p.id) || account.project_ids || [];
      setSelectedProjectIds(projectIds);
      setRelationships(account.relationships || []);
      setPendingContacts([]);
      setPendingDeletions([]);
      setPendingNotes([]);
      setPendingTasks([]);
      setPendingComments([]);
      setPendingDocumentIds([]);
      setPendingRelationships([]);
      setErrors({});
      setIsDeleting(false);
      return;
    }

    const initialData = {};
    config.fields.forEach((field) => {
      initialData[field.name] = field.defaultValue ?? "";
    });
    setFormData(initialData);
    setContacts([]);
    setRelationships([]);
    setSelectedIndustries([]);
    setSelectedProjectIds(
      projectState.selectedProject ? [projectState.selectedProject.id] : [],
    );
    setPendingContacts([]);
    setPendingDeletions([]);
    setPendingNotes([]);
    setPendingTasks([]);
    setPendingComments([]);
    setPendingRelationships([]);
    setErrors({});
    setIsDeleting(false);
  }, [isEditMode, account, config.fields, projectState.selectedProject?.id]);

  const formDataHasChanges = usePendingChanges({
    original: account,
    current: formData,
    pendingDeletions,
  });

  const originalProjectIds = account?.projects?.map((p) => p.id).sort() || [];
  const currentProjectIds = [...selectedProjectIds].sort();
  const projectsChanged =
    JSON.stringify(originalProjectIds) !== JSON.stringify(currentProjectIds);
  const hasPendingChanges = formDataHasChanges || projectsChanged;

  const handleChange = (field, value) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => {
        const next = { ...prev };
        delete next[field];
        return next;
      });
    }
  };

  const resolveIndustryIds = async () => {
    if (selectedIndustries.length === 0) return [];
    const allIndustries = await getIndustries();
    return selectedIndustries
      .map((name) => allIndustries.find((ind) => ind.name === name)?.id)
      .filter((id) => id !== undefined);
  };

  const processPendingContactDeletions = async (accountId) => {
    const deleteContact = isOrganization
      ? deleteOrganizationContact
      : deleteIndividualContact;
    for (const contact of pendingDeletions) {
      try {
        await deleteContact(accountId, contact.id);
      } catch (err) {
        console.error("Failed to delete contact:", err);
      }
    }
  };

  const buildContactData = (pending) => ({
    individual_id: pending.individual_id || undefined,
    first_name: pending.first_name?.trim() || undefined,
    last_name: pending.last_name?.trim() || undefined,
    email: pending.email?.trim() || undefined,
    phone: pending.phone?.trim() || undefined,
    title: pending.title?.trim() || undefined,
    notes: pending.notes?.trim() || undefined,
    is_primary: false,
  });

  const createPendingContacts = async (entityId) => {
    const createContact = isOrganization
      ? createOrganizationContact
      : createIndividualContact;
    for (const pending of pendingContacts) {
      try {
        await createContact(entityId, buildContactData(pending));
      } catch (err) {
        console.error("Failed to create contact:", err);
      }
    }
  };

  const createRelationshipContacts = async (entityId) => {
    for (const rel of pendingRelationships) {
      try {
        const shouldCreateOrgContact =
          isOrganization && rel.type === "individual";
        const shouldCreateIndContact =
          !isOrganization && rel.type === "organization";

        if (shouldCreateOrgContact) {
          await createOrganizationContact(entityId, {
            individual_id: rel.id,
            first_name: rel.name.split(" ")[0] || "",
            last_name: rel.name.split(" ").slice(1).join(" ") || "",
            is_primary: false,
          });
        }
        if (shouldCreateIndContact) {
          await createIndividualContact(entityId, {
            organization_id: rel.id,
            is_primary: false,
          });
        }
      } catch (err) {
        console.error("Failed to create relationship:", err);
      }
    }
  };

  const createPendingNotes = async (entityId) => {
    const entityType = isOrganization ? "organization" : "individual";
    for (const noteContent of pendingNotes) {
      try {
        await createNote(entityType, entityId, noteContent);
      } catch (err) {
        console.error("Failed to create note:", err);
      }
    }
  };

  const createPendingTasks = async (entityId) => {
    const entityType = isOrganization ? "Organization" : "Individual";
    for (const taskData of pendingTasks) {
      try {
        await createTask({
          ...taskData,
          taskable_type: entityType,
          taskable_id: entityId,
        });
      } catch (err) {
        console.error("Failed to create task:", err);
      }
    }
  };

  const linkPendingDocuments = async (entityId) => {
    const entityType = isOrganization ? "Organization" : "Individual";
    for (const documentId of pendingDocumentIds) {
      try {
        await linkDocumentToEntity(documentId, entityType, entityId);
      } catch (err) {
        console.error("Failed to link document:", err);
      }
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setErrors({});

    const validationErrors = config.onValidate?.(formData);
    if (validationErrors) {
      setErrors(validationErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      const submitData = config.onBeforeSubmit?.(formData) ?? formData;
      const industryIds = await resolveIndustryIds();
      const dataWithIndustriesAndProjects = {
        ...submitData,
        industry_ids: industryIds.length > 0 ? industryIds : undefined,
        project_ids: selectedProjectIds.length > 0 ? selectedProjectIds : [],
      };

      if (isEditMode && account && onUpdate) {
        await onUpdate(account.id, dataWithIndustriesAndProjects);
        await processPendingContactDeletions(account.id);
        onClose();
        return;
      }

      if (!onAdd) {
        onClose();
        return;
      }

      const created = await onAdd(dataWithIndustriesAndProjects);
      await createPendingContacts(created.id);
      await createRelationshipContacts(created.id);
      await createPendingNotes(created.id);
      await createPendingTasks(created.id);
      await linkPendingDocuments(created.id);
      onClose();
    } catch (err) {
      const message =
        err instanceof Error
          ? err.message
          : `Failed to ${isEditMode ? "update" : "create"}`;
      setErrors({ _form: message });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!account || !onDelete) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await onDelete(account.id);
      onClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to delete";
      setErrors({ _form: message });
      setIsDeleting(false);
    }
  };

  const handleAddContact = async (contact) => {
    if (!isEditMode) {
      setPendingContacts([...pendingContacts, contact]);
      return;
    }

    if (!account?.id) return;
    setErrors({});

    const createContact = isOrganization
      ? createOrganizationContact
      : createIndividualContact;
    const contactData = {
      individual_id: isOrganization
        ? contact.individual_id || undefined
        : undefined,
      first_name: contact.first_name?.trim() || undefined,
      last_name: contact.last_name?.trim() || undefined,
      email: contact.email?.trim() || undefined,
      phone: contact.phone?.trim() || undefined,
      is_primary: contacts.length === 0,
    };

    try {
      const newContactData = await createContact(account.id, contactData);
      setContacts([...contacts, newContactData]);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to add contact";
      setErrors({ _form: message });
    }
  };

  const handleRemoveContact = (index) => {
    if (!isEditMode) {
      setPendingContacts(pendingContacts.filter((_, i) => i !== index));
      return;
    }

    const contact = contacts[index];
    if (!contact) return;

    setPendingDeletions([...pendingDeletions, contact]);
    setContacts(contacts.filter((_, i) => i !== index));
  };

  const handleUpdateContact = async (index, updatedContact) => {
    if (!isEditMode) {
      setPendingContacts(
        pendingContacts.map((c, i) => (i === index ? updatedContact : c)),
      );
      return;
    }

    const contact = contacts[index];
    if (!account?.id || !contact) return;

    const updateContact = isOrganization
      ? updateOrganizationContact
      : updateIndividualContact;
    const updateData = {
      email: updatedContact.email?.trim() || undefined,
      phone: updatedContact.phone?.trim() || undefined,
    };

    try {
      const updated = await updateContact(account.id, contact.id, updateData);
      setContacts(contacts.map((c, i) => (i === index ? updated : c)));
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to update contact";
      setErrors({ _form: message });
    }
  };

  const handleSetPrimaryContact = async (index) => {
    if (!isEditMode || !account?.id) return;

    const contact = contacts[index];
    if (!contact) return;

    try {
      const updateContact = isOrganization
        ? updateOrganizationContact
        : updateIndividualContact;
      await updateContact(account.id, contact.id, { is_primary: true });

      setContacts(
        contacts.map((c, i) => ({
          ...c,
          is_primary: i === index,
        })),
      );
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to set primary contact";
      setErrors({ _form: message });
    }
  };

  const handleSearchContacts = async (query) => {
    const results = await searchContacts(query);
    return results.map((result) => ({
      individual_id: result.individual_id,
      first_name: result.first_name,
      last_name: result.last_name,
      title: result.title,
      email: result.email,
      phone: result.phone,
      account_name: result.account_name,
    }));
  };

  const handleSearchRelationships = async (query) => {
    const results = await searchAccounts(query);
    return results
      .filter((r) => r.type !== "unknown")
      .map((r) => ({
        id: r.id,
        name: r.name,
        type: r.type,
      }));
  };

  const handleAddRelationship = (relationship) => {
    if (isEditMode) {
      setRelationships([...relationships, relationship]);
      return;
    }
    setPendingRelationships([...pendingRelationships, relationship]);
  };

  const handleRemoveRelationship = (index) => {
    if (isEditMode) {
      setRelationships(relationships.filter((_, i) => i !== index));
      return;
    }
    setPendingRelationships(pendingRelationships.filter((_, i) => i !== index));
  };

  const displayRelationships = isEditMode
    ? relationships
    : pendingRelationships;
  const relationshipSearchType = isOrganization ? "individual" : "organization";

  const contactFields = [
    { name: "first_name", label: "First Name", placeholder: "e.g., John" },
    { name: "last_name", label: "Last Name", placeholder: "e.g., Doe" },
    { name: "title", label: "Title", placeholder: "e.g., VP of Sales" },
    {
      name: "email",
      label: "Email",
      type: "email",
      placeholder: "email@example.com",
    },
    {
      name: "phone",
      label: "Phone",
      type: "tel",
      placeholder: "+1-555-123-4567",
    },
  ];

  const title = isEditMode ? config.editTitle : config.title;

  const displayContacts = isEditMode
    ? contacts.map((c) => ({
        id: c.id,
        first_name: c.first_name || "",
        last_name: c.last_name || "",
        email: c.email || "",
        phone: c.phone || "",
        title: c.title,
        is_primary: c.is_primary,
      }))
    : pendingContacts;

  const hasPairedFields = config.fields.some(
    (f) => f.gridColumn === "md:col-span-1",
  );

  const getCustomRenderer = (fieldName) => {
    if (fieldName === "industry") {
      return (
        <IndustrySelector
          value={selectedIndustries}
          onChange={setSelectedIndustries}
        />
      );
    }
    if (fieldName === "employee_count_range_id") {
      return (
        <EmployeeCountRangeSelector
          value={formData.employee_count_range_id}
          onChange={(val) =>
            handleChange("employee_count_range_id", val?.toString() ?? "")
          }
        />
      );
    }
    if (fieldName === "funding_stage_id") {
      return (
        <FundingStageSelector
          value={formData.funding_stage_id}
          onChange={(val) =>
            handleChange("funding_stage_id", val?.toString() ?? "")
          }
        />
      );
    }
    if (fieldName === "revenue_range_id") {
      return (
        <RevenueRangeSelector
          value={formData.revenue_range_id}
          onChange={(val) =>
            handleChange("revenue_range_id", val?.toString() ?? "")
          }
        />
      );
    }
    return undefined;
  };

  const renderFormFields = () => {
    if (hasPairedFields) {
      return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {config.fields.map((field) => (
            <div
              key={field.name}
              className={
                field.gridColumn === "md:col-span-1" ? "" : "md:col-span-2"
              }
            >
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                {field.label}
                {field.required && <span className="text-red-500"> *</span>}
              </label>
              {renderField(
                field,
                formData[field.name],
                handleChange,
                getCustomRenderer(field.name),
              )}
              {errors[field.name] && (
                <p className="mt-1 text-sm text-red-500">
                  {errors[field.name]}
                </p>
              )}
            </div>
          ))}
        </div>
      );
    }

    return config.fields.map((field) => (
      <div key={field.name}>
        <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
          {field.label}
          {field.required && <span className="text-red-500"> *</span>}
        </label>
        {renderField(
          field,
          formData[field.name],
          handleChange,
          getCustomRenderer(field.name),
        )}
        {errors[field.name] && (
          <p className="mt-1 text-sm text-red-500">{errors[field.name]}</p>
        )}
      </div>
    ));
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-6">
        {title}
      </h1>

      <form onSubmit={handleSubmit} className="space-y-4">
        {renderFormFields()}

        {/* Project Assignment */}
        <div className="mt-4">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Projects
          </label>
          <ProjectSelector
            value={selectedProjectIds}
            onChange={setSelectedProjectIds}
            projects={projectState.projects}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <ContactSection
            contacts={displayContacts}
            onAdd={handleAddContact}
            onRemove={handleRemoveContact}
            onUpdate={handleUpdateContact}
            onSetPrimary={handleSetPrimaryContact}
            fields={contactFields}
            emptyContact={emptyPendingContact}
            display={{ primaryLabel: "Primary" }}
            showFormByDefault={false}
            individualSearch={{
              enabled: true,
              onSearch: handleSearchContacts,
            }}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <RelationshipsSection
            relationships={displayRelationships}
            onAdd={handleAddRelationship}
            onRemove={handleRemoveRelationship}
            searchType={relationshipSearchType}
            onSearch={handleSearchRelationships}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <NotesSection
            entityType={isOrganization ? "organization" : "individual"}
            entityId={isEditMode ? (account?.id ?? null) : null}
            pendingNotes={!isEditMode ? pendingNotes : undefined}
            onPendingNotesChange={!isEditMode ? setPendingNotes : undefined}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <TasksSection
            entityType={isOrganization ? "Organization" : "Individual"}
            entityId={isEditMode ? (account?.id ?? null) : null}
            pendingTasks={!isEditMode ? pendingTasks : undefined}
            onPendingTasksChange={!isEditMode ? setPendingTasks : undefined}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <DocumentsSection
            entityType={isOrganization ? "Organization" : "Individual"}
            entityId={isEditMode ? (account?.id ?? null) : null}
            pendingDocumentIds={!isEditMode ? pendingDocumentIds : undefined}
            onPendingDocumentIdsChange={
              !isEditMode ? setPendingDocumentIds : undefined
            }
          />
        </div>

        <div className="border-t border-zinc-200 dark:border-zinc-700 pt-4 mt-4">
          <FormActions
            isEditMode={isEditMode}
            isSubmitting={isSubmitting}
            isDeleting={isDeleting}
            hasPendingChanges={hasPendingChanges}
            onClose={onClose}
            onDelete={onDelete ? handleDelete : undefined}
            variant="compact"
          />
        </div>

        {isEditMode && account && (
          <Timestamps
            createdAt={account.created_at}
            updatedAt={account.updated_at}
          />
        )}

        {errors._form && (
          <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
            {errors._form}
          </div>
        )}
      </form>

      {/* Comments Section - below the form */}
      <div className="border-t border-zinc-200 dark:border-zinc-700 pt-6 mt-6">
        <CommentsSection
          entityType={isOrganization ? "Organization" : "Individual"}
          entityId={isEditMode ? (account?.id ?? null) : null}
          pendingComments={!isEditMode ? pendingComments : undefined}
          onPendingCommentsChange={!isEditMode ? setPendingComments : undefined}
        />
      </div>
    </div>
  );
};

export default AccountForm;
