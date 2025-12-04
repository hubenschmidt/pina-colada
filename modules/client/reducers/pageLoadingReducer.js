export const SET_PAGE_LOADING = "SET_PAGE_LOADING";
export const SET_REDIRECT = "SET_REDIRECT";






export default (initialState) => {
  return (
    state,
    action
  ) => {
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
  state,
  payload
) => {
  return {
    ...state,
    isPageLoading: payload,
  };
};

const setRedirect = (
  state,
  payload
) => {
  return {
    ...state,
    redirect: payload,
  };
};
