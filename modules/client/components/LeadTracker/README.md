# LeadTracker - Generic Lead Management Component

## Overview

`LeadTracker` is a generic, configurable component for managing different types of leads (Jobs, Opportunities, Partnerships, etc.) through dependency injection.

## Architecture

### Core Components

1. **LeadTracker** - Generic tracker component that accepts configuration
2. **LeadTrackerConfig** - Type definitions and interfaces
3. **Specific Configs** - Implementation configs (e.g., `JobLeadConfig`, `OpportunityLeadConfig`)

### Type Parameters

```typescript
LeadTrackerConfig<T, TInsert, TUpdate>
```

- `T extends BaseLead` - The lead entity type
- `TInsert` - Insert DTO type (defaults to `Omit<T, "id" | "created_at" | "updated_at">`)
- `TUpdate` - Update DTO type (defaults to `Partial<T>`)

## Usage

### Creating a New Lead Type

1. **Define the entity type** that extends `BaseLead`:

```typescript
interface OpportunityLead extends BaseLead {
  organization_name: string;
  estimated_value: number;
  probability: number;
  expected_close_date: string;
  notes: string;
}
```

2. **Create the form component** implementing `LeadFormProps<T>`:

```typescript
const OpportunityForm = ({ isOpen, onClose, onAdd }: LeadFormProps<OpportunityLead>) => {
  // Form implementation
};
```

3. **Create the edit modal** implementing `LeadEditModalProps<T>`:

```typescript
const OpportunityEditModal = ({ lead, opened, onClose, onUpdate, onDelete }: LeadEditModalProps<OpportunityLead>) => {
  // Modal implementation
};
```

4. **Create the configuration hook**:

```typescript
export const useOpportunityLeadConfig = (): LeadTrackerConfig<OpportunityLead> => {
  const columns: Column<OpportunityLead>[] = useMemo(() => [
    { header: "Organization", accessor: "organization_name", sortable: true },
    { header: "Value", accessor: "estimated_value", render: (opp) => `$${opp.estimated_value}` },
    // ... more columns
  ], []);

  return {
    name: "opportunity",
    entityName: "Opportunity",
    entityNamePlural: "Opportunities",
    columns,
    FormComponent: OpportunityForm,
    EditModalComponent: OpportunityEditModal,
    api: {
      getLeads: getOpportunities,
      createLead: createOpportunity,
      updateLead: updateOpportunity,
      deleteLead: deleteOpportunity,
    },
    defaultSortBy: "created_at",
    defaultSortDirection: "DESC",
  };
};
```

5. **Create the tracker component**:

```typescript
const OpportunityTracker = forwardRef<LeadTrackerRef, {}>((props, ref) => {
  const config = useOpportunityLeadConfig();
  return <LeadTracker config={config} ref={ref} />;
});
```

## Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `name` | string | Yes | - | Internal identifier |
| `entityName` | string | Yes | - | Singular name (e.g., "Job") |
| `entityNamePlural` | string | Yes | - | Plural name (e.g., "Jobs") |
| `columns` | Column<T>[] | Yes | - | Table column definitions |
| `FormComponent` | Component | Yes | - | Form component for adding leads |
| `EditModalComponent` | Component | Yes | - | Modal for editing/deleting leads |
| `api` | LeadAPI<T> | Yes | - | API functions (CRUD) |
| `defaultSortBy` | string | No | "created_at" | Default sort column |
| `defaultSortDirection` | "ASC" \| "DESC" | No | "DESC" | Default sort direction |
| `defaultPageSize` | number | No | 25 | Items per page |
| `searchPlaceholder` | string | No | Auto-generated | Search input placeholder |
| `emptyMessage` | string | No | Auto-generated | Message when no data |
| `enableSearch` | boolean | No | true | Enable/disable search |
| `enableExport` | boolean | No | false | Enable/disable export (future) |

## API Interface

```typescript
interface LeadAPI<T, TInsert, TUpdate> {
  getLeads: (page, limit, sortBy, sortDirection, search?) => Promise<PageData<T>>;
  createLead: (lead: TInsert) => Promise<void>;
  updateLead: (id: string, updates: TUpdate) => Promise<void>;
  deleteLead: (id: string) => Promise<void>;
}
```

## Example: JobTracker

See `components/JobTracker/JobLeadConfig.tsx` for a complete implementation example.

```typescript
// Usage
import JobTracker, { JobTrackerRef } from '@/components/JobTracker/JobTrackerNew';

const MyPage = () => {
  const trackerRef = useRef<JobTrackerRef>(null);

  return (
    <div>
      <button onClick={() => trackerRef.current?.openLeadForm()}>
        Add Job
      </button>
      <JobTracker ref={trackerRef} />
    </div>
  );
};
```

## Ref Methods

```typescript
interface LeadTrackerRef {
  openLeadForm: () => void;  // Open the add form
  refresh: () => void;        // Refresh the data
}
```

## Benefits

1. **DRY** - Single implementation for all lead types
2. **Type Safety** - Full TypeScript support with generics
3. **Flexibility** - Easy to add new lead types
4. **Testability** - Components are decoupled
5. **Maintainability** - Changes to core tracker benefit all implementations
6. **Scalability** - Can support DB-driven configs in the future
