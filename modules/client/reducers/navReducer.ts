export const SET_AGENT_OPEN = "SET_AGENT_OPEN";

export interface NavState {
  agentOpen: boolean;
}

export default (initialState: NavState) => {
  return (state, action) => {
    switch (action.type) {
      case SET_AGENT_OPEN:
        return setAgentOpen(state, action.payload);
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
