# Agent Model Configuration Sidebar

## Overview
Converted the Agent Model Configuration gear icon dropdown to a full-height sidebar that slides in from the right side of the chat page.

## Changes Made

### 1. Sidebar UI
- [x] Replace dropdown with right-side sliding sidebar
- [x] Full height (100vh)
- [x] Width: 450px (max 90vw on mobile)
- [x] Overlay with backdrop (darken background)
- [x] Smooth slide-in/out animation (transform: translateX, 0.3s ease)
- [x] Close via X button or backdrop click

### 2. Global Provider Selection
- [x] New section at top of sidebar: "Global Provider"
- [x] Dropdown: OpenAI | Anthropic
- [x] Immediate update: changing provider instantly updates all agent nodes to that provider's first model
- [x] Calls `updateAgentNodeConfig` for each node in parallel

### 3. Default Preset Fix
- [x] Changed frontend fallback from `"standard"` to `"premium"`
- [x] Line 337: `value={config?.selected_cost_tier || "premium"}`

## Files Modified

| File | Changes |
|------|---------|
| `components/Chat/AgentConfigMenu.jsx` | Converted to sidebar, added Global Provider, X icon import, provider change handler |
| `components/Chat/AgentConfigMenu.module.css` | Added sidebar, backdrop, header, close button, provider section styles |

## CSS Classes Added

- `.sidebar` - Fixed position container, slides from right
- `.sidebarOpen` - Transform to show sidebar
- `.backdrop` / `.backdropVisible` - Dark overlay behind sidebar
- `.sidebarHeader` - Header with title and close button
- `.sidebarTitle` - Title text styling
- `.closeButton` - X button styling
- `.sidebarContent` - Scrollable content area
- `.providerSection` - Global provider dropdown section
- `.providerLabel` / `.providerSelect` - Provider dropdown styling

## Testing

1. Click gear icon to open sidebar
2. Verify sidebar slides in from right with backdrop
3. Test Global Provider dropdown - should update all nodes immediately
4. Test cost tier dropdown - should default to "Premium"
5. Click backdrop or X to close sidebar
6. Verify scroll works for long node lists
