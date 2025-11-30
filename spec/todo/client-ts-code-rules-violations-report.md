# TypeScript/JavaScript Code Rules Violations Report
# Generated for modules/client directory

## Summary
- Total violations found: 199
- Files analyzed: 93

### Violations by Type

- **complex_function**: 12
- **guard_clause**: 3
- **long_function**: 8
- **nested_conditional**: 173
- **switch_case**: 3

## Detailed Violations

### Complex Function

#### modules/client/app/RootLayoutClient.tsx

- **Line 10**: KISS - Keep It Simple
  ```typescript
  export const RootLayoutClient = ({
  ```
  - Note: Function has 7 conditionals - consider breaking into smaller functions

#### modules/client/app/accounts/organizations/page.tsx

- **Line 65**: KISS - Keep It Simple
  ```typescript
  const OrganizationsPage = () => {
  ```
  - Note: Function has 7 conditionals - consider breaking into smaller functions

#### modules/client/app/settings/page.tsx

- **Line 23**: KISS - Keep It Simple
  ```typescript
  const SettingsPage = () => {
  ```
  - Note: Function has 7 conditionals - consider breaking into smaller functions

#### modules/client/components/AccountForm/AccountForm.tsx

- **Line 481**: KISS - Keep It Simple
  ```typescript
  const AccountForm = ({
  ```
  - Note: Function has 12 conditionals - consider breaking into smaller functions

- **Line 553**: KISS - Keep It Simple
  ```typescript
  const handleSubmit = async (e: React.FormEvent) => {
  ```
  - Note: Function has 8 conditionals - consider breaking into smaller functions

#### modules/client/components/Chat/Chat.tsx

- **Line 191**: KISS - Keep It Simple
  ```typescript
  export const renderWithLinks = (text: string): React.ReactNode[] => {
  ```
  - Note: Function has 7 conditionals - consider breaking into smaller functions

- **Line 293**: KISS - Keep It Simple
  ```typescript
  const Chat = ({ variant = "embedded", onConnectionChange }: ChatProps) => {
  ```
  - Note: Function has 10 conditionals - consider breaking into smaller functions

#### modules/client/components/LeadTracker/LeadForm.tsx

- **Line 275**: KISS - Keep It Simple
  ```typescript
  const getErrorMessage = (err: any): string => {
  ```
  - Note: Function has 6 conditionals - consider breaking into smaller functions

- **Line 290**: KISS - Keep It Simple
  ```typescript
  const handleSubmit = async (e: React.FormEvent) => {
  ```
  - Note: Function has 11 conditionals - consider breaking into smaller functions

- **Line 402**: KISS - Keep It Simple
  ```typescript
  const renderField = (field: FormFieldConfig<T>) => {
  ```
  - Note: Function has 8 conditionals - consider breaking into smaller functions

#### modules/client/components/NotesSection.tsx

- **Line 15**: KISS - Keep It Simple
  ```typescript
  const NotesSection = ({
  ```
  - Note: Function has 8 conditionals - consider breaking into smaller functions

#### modules/client/hooks/useWs.tsx

- **Line 66**: KISS - Keep It Simple
  ```typescript
  export const useWs = (url: string): UseWebSocketReturn => {
  ```
  - Note: Function has 11 conditionals - consider breaking into smaller functions

### Guard Clause

#### modules/client/app/accounts/contacts/page.tsx

- **Line 122**: Avoid else statements; use guard clauses
  ```typescript
  } else {
  ```

#### modules/client/app/accounts/individuals/page.tsx

- **Line 119**: Avoid else statements; use guard clauses
  ```typescript
  } else {
  ```

#### modules/client/app/accounts/organizations/page.tsx

- **Line 125**: Avoid else statements; use guard clauses
  ```typescript
  } else {
  ```

### Long Function

#### modules/client/api/index.ts

- **Line 12**: KISS - Keep It Simple
  ```typescript
  const getClient = () => axios.create();
  ```
  - Note: Function is 63 lines - consider breaking into smaller functions

- **Line 178**: KISS - Keep It Simple
  ```typescript
  export const markLeadAsDoNotApply = async (
  ```
  - Note: Function is 63 lines - consider breaking into smaller functions

- **Line 289**: KISS - Keep It Simple
  ```typescript
  export const updateTenantPreferences = async (
  ```
  - Note: Function is 68 lines - consider breaking into smaller functions

- **Line 424**: KISS - Keep It Simple
  ```typescript
  export const updateOrganization = async (
  ```
  - Note: Function is 169 lines - consider breaking into smaller functions

- **Line 739**: KISS - Keep It Simple
  ```typescript
  export const updateNote = async (
  ```
  - Note: Function is 78 lines - consider breaking into smaller functions

- **Line 911**: KISS - Keep It Simple
  ```typescript
  export const createOrganizationSignal = async (
  ```
  - Note: Function is 139 lines - consider breaking into smaller functions

#### modules/client/components/Chat/Chat.tsx

- **Line 46**: KISS - Keep It Simple
  ```typescript
  const parseUTM = (u: URL) =>
  ```
  - Note: Function is 128 lines - consider breaking into smaller functions

#### modules/client/components/LeadTracker/hooks/useLeadTrackerConfig.tsx

- **Line 18**: KISS - Keep It Simple
  ```typescript
  const getJobLeadConfig = (): LeadTrackerConfig<
  ```
  - Note: Function is 174 lines - consider breaking into smaller functions

### Nested Conditional

#### modules/client/app/RootLayoutClient.tsx

- **Line 26**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (pathname === "/") {
  ```

- **Line 27**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.tenantName) {
  ```

- **Line 36**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (pathname === "/tenant/select" && userState.tenantName) {
  ```

- **Line 47**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isProtectedRoute && !userState.tenantName) {
  ```

#### modules/client/app/[...catchall]/page.tsx

- **Line 12**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.isAuthed) {
  ```

#### modules/client/app/accounts/contacts/page.tsx

- **Line 100**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (searchQuery.trim()) {
  ```

- **Line 114**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (sortBy === "updated_at") {
  ```

- **Line 118**: Avoid nested conditionals; use guard clauses
  ```typescript
  } else if (sortBy === "account") {
  ```

#### modules/client/app/accounts/individuals/page.tsx

- **Line 97**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (searchQuery.trim()) {
  ```

- **Line 111**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (sortBy === "updated_at") {
  ```

- **Line 115**: Avoid nested conditionals; use guard clauses
  ```typescript
  } else if (sortBy === "industry") {
  ```

#### modules/client/app/accounts/organizations/page.tsx

- **Line 96**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (searchQuery.trim()) {
  ```

- **Line 109**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (sortBy === "updated_at") {
  ```

- **Line 113**: Avoid nested conditionals; use guard clauses
  ```typescript
  } else if (sortBy === "industries") {
  ```

- **Line 117**: Avoid nested conditionals; use guard clauses
  ```typescript
  } else if (sortBy === "funding_stage") {
  ```

- **Line 121**: Avoid nested conditionals; use guard clauses
  ```typescript
  } else if (sortBy === "employee_count_range") {
  ```

#### modules/client/app/api/user/tenant/route.ts

- **Line 8**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!session?.user) {
  ```

- **Line 26**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!response.ok) {
  ```

- **Line 27**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (response.status === 404) {
  ```

#### modules/client/app/login/page.tsx

- **Line 17**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isLoading && user) {
  ```

- **Line 22**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (tenant && tenant.id) {
  ```

#### modules/client/app/not-found.tsx

- **Line 12**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.isAuthed) {
  ```

#### modules/client/app/page.tsx

- **Line 22**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (ticking) return;
  ```

- **Line 27**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (nearTop && window.location.hash) {
  ```

#### modules/client/app/reports/custom/page.tsx

- **Line 32**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!confirm("Are you sure you want to delete this report?")) {
  ```

#### modules/client/app/settings/page.tsx

- **Line 31**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!userState.isAuthed) return;
  ```

- **Line 37**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.canEditTenantTheme) {
  ```

- **Line 77**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userTheme === null) {
  ```

#### modules/client/app/tenant/select/page.tsx

- **Line 42**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!userLoading && !user) {
  ```

- **Line 51**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!userState.user) {
  ```

- **Line 58**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!userState.bearerToken) {
  ```

- **Line 83**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!tenantName.trim()) {
  ```

#### modules/client/components/AccountForm/AccountForm.tsx

- **Line 127**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (dropdown && !dropdown.contains(event.target as Node)) {
  ```

- **Line 220**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (e.key === "Enter") {
  ```

- **Line 225**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (e.key === "Escape") {
  ```

- **Line 506**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode && account) {
  ```

- **Line 544**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (errors[field]) {
  ```

- **Line 558**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (validationErrors) {
  ```

- **Line 569**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (selectedIndustries.length > 0) {
  ```

- **Line 579**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode && account && onUpdate) {
  ```

- **Line 596**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!onAdd) {
  ```

- **Line 605**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isOrganization) {
  ```

- **Line 617**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isOrganization) {
  ```

- **Line 635**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isOrganization && rel.type === "individual") {
  ```

- **Line 643**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isOrganization && rel.type === "organization") {
  ```

- **Line 681**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isDeleting) {
  ```

- **Line 697**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isEditMode) {
  ```

- **Line 706**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isOrganization) {
  ```

- **Line 734**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isEditMode) {
  ```

- **Line 748**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isEditMode) {
  ```

- **Line 762**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isOrganization) {
  ```

- **Line 821**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode) {
  ```

- **Line 829**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode) {
  ```

#### modules/client/components/AccountForm/hooks/useAccountFormConfig.ts

- **Line 227**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.name || String(formData.name).trim() === "") {
  ```

- **Line 258**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.first_name || String(formData.first_name).trim() === "") {
  ```

- **Line 261**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.last_name || String(formData.last_name).trim() === "") {
  ```

#### modules/client/components/AuthStateManager.tsx

- **Line 19**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isLoading) return;
  ```

- **Line 20**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!auth0User) return;
  ```

- **Line 21**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.isAuthed) return;
  ```

- **Line 34**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (tenant) {
  ```

#### modules/client/components/Chat/Chat.tsx

- **Line 96**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (perm?.state === "granted") {
  ```

- **Line 98**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (pos) {
  ```

- **Line 178**: Avoid nested conditionals; use guard clauses
  ```typescript
  if ("getBattery" in navigator) {
  ```

- **Line 208**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isError) {
  ```

- **Line 223**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!line) {
  ```

- **Line 341**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (
  ```

- **Line 349**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (toolsDropdownOpen) {
  ```

- **Line 359**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (
  ```

- **Line 367**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (demoDropdownOpen) {
  ```

- **Line 390**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (e.key === "Enter" && !e.shiftKey) {
  ```

#### modules/client/components/ContactForm.tsx

- **Line 46**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field === "phone") {
  ```

- **Line 59**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isEditMode && savedContact && pendingNotes.length > 0) {
  ```

- **Line 74**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isDeleting) {
  ```

#### modules/client/components/ContactSection.tsx

- **Line 83**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (debounceTimer) {
  ```

- **Line 87**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!individualSearch || query.length < 2) {
  ```

- **Line 129**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!newContact.first_name?.trim() || !newContact.last_name?.trim()) {
  ```

- **Line 153**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (display.nameFields && display.nameFields.length > 0) {
  ```

- **Line 160**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (contact.first_name || contact.last_name) {
  ```

- **Line 177**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (editingIndex !== null && editingContact && onUpdate) {
  ```

- **Line 196**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditing && editingContact) {
  ```

#### modules/client/components/DataTable/DataTable.tsx

- **Line 117**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (typeof col.accessor === 'string') {
  ```

- **Line 183**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (val)
  ```

#### modules/client/components/FormActions.tsx

- **Line 33**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isSubmitting) {
  ```

#### modules/client/components/LeadTracker/LeadForm.tsx

- **Line 53**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "date" && value) {
  ```

- **Line 61**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (data.account_type !== "Individual" || !data.account) {
  ```

- **Line 66**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (parts.length === 2) {
  ```

- **Line 83**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (fieldName === "individual_first_name") {
  ```

- **Line 87**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (fieldName === "individual_last_name") {
  ```

- **Line 99**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.required && (!value || value === "")) {
  ```

- **Line 104**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!field.validation || !value) {
  ```

- **Line 109**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (error) {
  ```

- **Line 119**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!config.onValidate) {
  ```

- **Line 124**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (formErrors) {
  ```

- **Line 150**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode && lead) {
  ```

- **Line 171**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.defaultValue !== undefined) {
  ```

- **Line 175**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "date") {
  ```

- **Line 179**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "checkbox") {
  ```

- **Line 192**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.onInit) {
  ```

- **Line 215**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field?.onChange) {
  ```

- **Line 223**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (fieldName === "account_type") {
  ```

- **Line 232**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (
  ```

- **Line 250**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (errors[fieldName]) {
  ```

- **Line 280**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (err?.errorData && typeof err.errorData === "object") {
  ```

- **Line 293**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!validate()) {
  ```

- **Line 302**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (submitData.account_type === "Individual" && !isEditMode) {
  ```

- **Line 305**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (lastName || firstName) {
  ```

- **Line 316**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (validContacts.length > 0) {
  ```

- **Line 325**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (config.onBeforeSubmit) {
  ```

- **Line 329**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode && lead && onUpdate) {
  ```

- **Line 336**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!onAdd) {
  ```

- **Line 345**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (createdLead && pendingNotes.length > 0) {
  ```

- **Line 355**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.defaultValue !== undefined) {
  ```

- **Line 359**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "date") {
  ```

- **Line 363**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "checkbox") {
  ```

- **Line 387**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isDeleting) {
  ```

- **Line 416**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isAccountField && accountType === "Individual") {
  ```

- **Line 453**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isAccountTypeField && accountType === "Individual") {
  ```

- **Line 525**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "custom" && field.renderCustom) {
  ```

- **Line 538**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "textarea") {
  ```

- **Line 554**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "select") {
  ```

- **Line 575**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (field.type === "checkbox") {
  ```

- **Line 598**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (hasUrl) {
  ```

- **Line 643**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode && lead) {
  ```

- **Line 702**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isContactSection) {
  ```

- **Line 745**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (Array.isArray(rendered)) {
  ```

#### modules/client/components/LeadTracker/LeadTracker.tsx

- **Line 44**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (showFullLoading) {
  ```

- **Line 47**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (showBar) {
  ```

- **Line 76**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (data !== null && !loading) {
  ```

- **Line 85**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (config.detailPagePath) {
  ```

- **Line 169**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (config.newPagePath) {
  ```

#### modules/client/components/LeadTracker/hooks/useLeadFormConfig.tsx

- **Line 36**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!searchQuery || searchQuery.length < 2) {
  ```

- **Line 62**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (dropdown && !dropdown.contains(event.target as Node)) {
  ```

- **Line 93**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isOrganization(item)) {
  ```

- **Line 175**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (dropdown && !dropdown.contains(event.target as Node)) {
  ```

- **Line 193**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (newIndustry.trim()) {
  ```

- **Line 269**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (e.key === "Enter") {
  ```

- **Line 274**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (e.key === "Escape") {
  ```

- **Line 417**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!e.target.checked) {
  ```

- **Line 513**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isEditMode) {
  ```

- **Line 531**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (cleaned[key] !== undefined) {
  ```

- **Line 545**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!cleaned.industry || (Array.isArray(cleaned.industry) && cleaned.industry.le
  ```

- **Line 550**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!cleaned.contacts || (Array.isArray(cleaned.contacts) && cleaned.contacts.le
  ```

- **Line 570**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (accountType === "Individual") {
  ```

- **Line 572**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.individual_first_name || formData.individual_first_name.trim() ===
  ```

- **Line 575**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.individual_last_name || formData.individual_last_name.trim() === "
  ```

- **Line 582**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.account || formData.account.trim() === "") {
  ```

- **Line 586**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!formData.job_title) {
  ```

#### modules/client/components/LeadTracker/hooks/useLeadTrackerConfig.tsx

- **Line 38**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!job.account || job.account.trim() === "") {
  ```

- **Line 84**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!job.notes) return <span className="text-zinc-400">—</span>;
  ```

#### modules/client/components/NotesSection.tsx

- **Line 32**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!isCreateMode && entityId) {
  ```

- **Line 53**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (isCreateMode && onPendingNotesChange) {
  ```

- **Line 99**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (onPendingNotesChange && pendingNotes) {
  ```

#### modules/client/components/RelationshipsSection.tsx

- **Line 53**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (debounceTimer) {
  ```

- **Line 57**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!onSearch || query.length < 2) {
  ```

- **Line 220**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (readOnly) {
  ```

#### modules/client/components/Reports/ReportBuilder.tsx

- **Line 86**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!entity) {
  ```

- **Line 131**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!entity || selectedColumns.length === 0) {
  ```

- **Line 148**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!entity || selectedColumns.length === 0) {
  ```

- **Line 172**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!saveName.trim()) {
  ```

- **Line 175**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (!onSave) {
  ```

- **Line 376**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (col === "id" && entityLink) {
  ```

#### modules/client/components/SearchHeader.tsx

- **Line 28**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (debounceTimer) {
  ```

- **Line 41**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (debounceTimer) {
  ```

#### modules/client/components/ThemeApplier.tsx

- **Line 15**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.theme === "dark") {
  ```

- **Line 19**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (userState.theme === "light") {
  ```

#### modules/client/hooks/useWs.tsx

- **Line 94**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (wsRef.current === socket) {
  ```

- **Line 102**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (typeof event.data !== "string") return;
  ```

- **Line 116**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (parsed === null || typeof parsed !== "object") return;
  ```

- **Line 120**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (obj.on_chat_model_start === true) {
  ```

- **Line 128**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (obj.on_token_usage && typeof obj.on_token_usage === "object") {
  ```

- **Line 148**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (typeof chunk === "string" && chunk.length > 0) {
  ```

- **Line 154**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (obj.on_chat_model_end === true) {
  ```

- **Line 167**: Avoid nested conditionals; use guard clauses
  ```typescript
  if (uiKeys.length === 0) return;
  ```

### Switch Case

#### modules/client/reducers/navReducer.ts

- **Line 11**: Avoid switch/case; use guard clauses or dict dispatch
  ```typescript
  switch (action.type) {
  ```

#### modules/client/reducers/pageLoadingReducer.ts

- **Line 14**: Avoid switch/case; use guard clauses or dict dispatch
  ```typescript
  switch (action.type) {
  ```

#### modules/client/reducers/userReducer.ts

- **Line 22**: Avoid switch/case; use guard clauses or dict dispatch
  ```typescript
  switch (action.type) {
  ```
