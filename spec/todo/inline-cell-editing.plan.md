# Inline Cell Editing for DataTable

## Goal
Allow inline editing of cells (starting with Status) directly from any DataTable, extensible to other columns.

## Files to Modify

### 1. `components/DataTable/DataTable.jsx`
- Add `onCellUpdate` prop: `(row, field, value) => Promise<void>`
- Modify `getCellContent` to pass context to render: `col.render(row, { onCellUpdate })`

```jsx
const getCellContent = (col, row) => {
  const context = { onCellUpdate: onCellUpdate ? (field, value) => onCellUpdate(row, field, value) : null };
  if (col.render) return col.render(row, context);
  // ... rest unchanged
};
```

### 2. `components/StatusSelect/StatusSelect.jsx` (new file)
Generic status dropdown component:
```jsx
const StatusSelect = ({ value, options, colors, onUpdate, field = "status" }) => {
  const [localValue, setLocalValue] = useState(value);
  const [loading, setLoading] = useState(false);

  const handleChange = async (newValue) => {
    if (!newValue || newValue === localValue || !onUpdate) return;
    setLoading(true);
    setLocalValue(newValue);
    try {
      await onUpdate(field, newValue);
    } catch {
      setLocalValue(value);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Select
      size="xs"
      value={localValue}
      data={options}
      onChange={handleChange}
      disabled={loading || !onUpdate}
      onClick={(e) => e.stopPropagation()}
      allowDeselect={false}
      classNames={{ input: colors[localValue] || "" }}
    />
  );
};
```

### 3. `components/LeadTracker/LeadTracker.jsx`
- Add `handleCellUpdate` function that calls `config.api.updateLead`
- Pass to DataTable as `onCellUpdate`

```jsx
const handleCellUpdate = async (row, field, value) => {
  await config.api.updateLead(row.id, { [field]: value });
};

<DataTable ... onCellUpdate={handleCellUpdate} />
```

### 4. `components/LeadTracker/hooks/useLeadTrackerConfig.jsx`
- Export status options/colors constants
- Update status column renders for all 3 types (job, opportunity, partnership):

```jsx
render: (row, { onCellUpdate }) => (
  <StatusSelect
    value={row.status}
    options={JOB_STATUS_OPTIONS}
    colors={JOB_STATUS_COLORS}
    onUpdate={onCellUpdate}
  />
)
```
