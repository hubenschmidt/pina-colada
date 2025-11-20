export const SET_AGENT_OPEN = "SET_AGENT_OPEN";
export const SET_SIDEBAR_COLLAPSED = "SET_SIDEBAR_COLLAPSED";

export interface NavState {
  agentOpen: boolean;
  sidebarCollapsed: boolean;
}

export default (initialState: NavState) => {
  return (state, action) => {
    switch (action.type) {
      case SET_AGENT_OPEN:
        return setAgentOpen(state, action.payload);
      case SET_SIDEBAR_COLLAPSED:
        return setSidebarCollapsed(state, action.payload);
      default:
        return state;
    }
  };
};

const setAgentOpen = (state, payload) => {
  return {
    ...state,
    agentOpen: payload,
  };
};

const setSidebarCollapsed = (state, payload) => {
  return {
    ...state,
    sidebarCollapsed: payload,
  };
};
