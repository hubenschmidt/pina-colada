export const SET_USER = "SET_USER";
export const SET_TENANT_NAME = "SET_TENANT_NAME";
export const SET_BEARER_TOKEN = "SET_BEARER_TOKEN";
export const SET_AUTHED = "SET_AUTHED";
export const SET_THEME = "SET_THEME";
export const SET_LOADING = "SET_LOADING";
export const SET_SELECTED_PROJECT_ID = "SET_SELECTED_PROJECT_ID";
export const SET_ROLES = "SET_ROLES";
export const CLEAR_USER = "CLEAR_USER";

export default (initialState) => {
  return (state, action) => {
    switch (action.type) {
      case SET_USER:
        return setUser(state, action.payload);
      case SET_TENANT_NAME:
        return setTenantName(state, action.payload);
      case SET_BEARER_TOKEN:
        return setBearerToken(state, action.payload);
      case SET_AUTHED:
        return setAuthed(state, action.payload);
      case SET_THEME:
        return setTheme(state, action.payload);
      case SET_LOADING:
        return setLoading(state, action.payload);
      case SET_SELECTED_PROJECT_ID:
        return setSelectedProjectId(state, action.payload);
      case SET_ROLES:
        return setRoles(state, action.payload);
      case CLEAR_USER:
        return clearUser(state);
      default:
        return state;
    }
  };
};

const setUser = (state, payload) => {
  return {
    ...state,
    user: payload,
  };
};

const setTenantName = (state, payload) => {
  return {
    ...state,
    tenantName: payload,
  };
};

const setBearerToken = (state, payload) => {
  return {
    ...state,
    bearerToken: payload,
  };
};

const setAuthed = (state, payload) => {
  return {
    ...state,
    isAuthed: true,
  };
};

const setTheme = (state, payload) => ({
  ...state,
  theme: payload.theme,
  canEditTenantTheme: payload.canEditTenant,
});

const setLoading = (state, payload) => ({
  ...state,
  isLoading: payload,
});

const setSelectedProjectId = (state, payload) => ({
  ...state,
  selectedProjectId: payload,
});

const setRoles = (state, payload) => ({
  ...state,
  roles: payload,
});

const clearUser = (state) => ({
  ...state,
  user: null,
  tenantName: null,
  bearerToken: null,
  isAuthed: false,
  isLoading: false,
  theme: "light",
  canEditTenantTheme: false,
  selectedProjectId: null,
  roles: [],
});
