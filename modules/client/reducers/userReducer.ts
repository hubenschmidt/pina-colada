export const SET_USER = "SET_USER";
export const SET_TENANT_NAME = "SET_TENANT_NAME";
export const SET_BEARER_TOKEN = "SET_BEARER_TOKEN";
export const SET_AUTHED = "SET_AUTHED";
export const CLEAR_USER = "CLEAR_USER";

export interface UserState {
  user: any | null;
  tenantName: string | null;
  bearerToken: string | null;
  isAuthed: boolean;
}

export default (initialState: UserState) => {
  return (
    state: UserState,
    action: { type: string; payload?: any }
  ): UserState => {
    switch (action.type) {
      case SET_USER:
        return setUser(state, action.payload);
      case SET_TENANT_NAME:
        return setTenantName(state, action.payload);
      case SET_BEARER_TOKEN:
        return setBearerToken(state, action.payload);
      case SET_AUTHED:
        return setAuthed(state, action.payload);
      case CLEAR_USER:
        return clearUser(state);
      default:
        return state;
    }
  };
};

const setUser = (state: UserState, payload: any): UserState => {
  return {
    ...state,
    user: payload,
  };
};

const setTenantName = (state: UserState, payload: string): UserState => {
  return {
    ...state,
    tenantName: payload,
  };
};

const setBearerToken = (state: UserState, payload: string): UserState => {
  return {
    ...state,
    bearerToken: payload,
  };
};

const setAuthed = (state: UserState, payload: string): UserState => {
  return {
    ...state,
    isAuthed: true,
  };
};

const clearUser = (state: UserState): UserState => {
  return {
    ...state,
    user: null,
    tenantName: null,
    bearerToken: null,
    isAuthed: false,
  };
};
