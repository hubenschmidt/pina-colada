export const SET_PAGE_LOADING = "SET_PAGE_LOADING";
export const SET_REDIRECT = "SET_REDIRECT";

export interface PageLoadingState {
  isPageLoading: boolean;
  redirect: boolean;
}

export default (initialState: PageLoadingState) => {
  return (
    state: PageLoadingState,
    action: { type: string; payload?: any }
  ): PageLoadingState => {
    switch (action.type) {
      case SET_PAGE_LOADING:
        return setPageLoading(state, action.payload);
      case SET_REDIRECT:
        return setRedirect(state, action.payload);
      default:
        return state;
    }
  };
};

const setPageLoading = (
  state: PageLoadingState,
  payload: boolean
): PageLoadingState => {
  return {
    ...state,
    isPageLoading: payload,
  };
};

const setRedirect = (
  state: PageLoadingState,
  payload: boolean
): PageLoadingState => {
  return {
    ...state,
    redirect: payload,
  };
};
