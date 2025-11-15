/**
 * Generic lead configurations
 *
 * Use these factories to get type-specific configurations:
 * - useFormConfig(type): Form field definitions
 * - useLeadConfig(type): Tracker configuration (columns, API, components)
 * - usePanelConfig(type): Panel configuration (filtering, actions)
 *
 * Currently supported types: "job"
 * Future: "opportunity", "partnership"
 */

export { useFormConfig } from "./FormConfig";
export { useLeadConfig } from "./LeadConfig";
export { usePanelConfig } from "./PanelConfig";
