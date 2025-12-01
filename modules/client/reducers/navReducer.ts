export const SET_AGENT_OPEN = "SET_AGENT_OPEN";
export const SET_SIDEBAR_COLLAPSED = "SET_SIDEBAR_COLLAPSED";
export const SET_SIDEBAR_SECTION_EXPANDED = "SET_SIDEBAR_SECTION_EXPANDED";

export interface NavState {
  agentOpen: boolean;
  sidebarCollapsed: boolean;
  sidebarSections: {
    accounts: boolean;
    leads: boolean;
    reports: boolean;
  };
}

export default (initialState: NavState) => {
  return (state: NavState, action: { type: string; payload?: any }) => {
    switch (action.type) {
      case SET_AGENT_OPEN:
        return setAgentOpen(state, action.payload);
      case SET_SIDEBAR_COLLAPSED:
        return setSidebarCollapsed(state, action.payload);
      case SET_SIDEBAR_SECTION_EXPANDED:
        return setSidebarSectionExpanded(state, action.payload);
      default:
        return state;
    }
  };
};

const setAgentOpen = (state: NavState, payload: boolean) => {
  return {
    ...state,
    agentOpen: payload,
  };
};

const setSidebarCollapsed = (state: NavState, payload: boolean) => {
  return {
    ...state,
    sidebarCollapsed: payload,
  };
};

const setSidebarSectionExpanded = (
  state: NavState,
  payload: { section: keyof NavState["sidebarSections"]; expanded: boolean }
) => {
  return {
    ...state,
    sidebarSections: {
      ...state.sidebarSections,
      [payload.section]: payload.expanded,
    },
  };
};
