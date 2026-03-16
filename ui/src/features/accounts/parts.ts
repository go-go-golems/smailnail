// CSS part names for theming via data-part attributes
// Consumers can style any part with: [data-widget="account-setup"][data-part="..."]

export const WIDGET = "account-setup";

export const parts = {
  root: "root",
  emptyState: "empty-state",
  emptyHeadline: "empty-headline",
  emptyBody: "empty-body",
  form: "form",
  formHeader: "form-header",
  formBody: "form-body",
  formActions: "form-actions",
  fieldGroup: "field-group",
  advancedToggle: "advanced-toggle",
  advancedPanel: "advanced-panel",
  testProgress: "test-progress",
  testStage: "test-stage",
  resultPanel: "result-panel",
  resultHeadline: "result-headline",
  resultChecklist: "result-checklist",
  resultWarning: "result-warning",
  resultError: "result-error",
  resultActions: "result-actions",
  sampleInfo: "sample-info",
} as const;
