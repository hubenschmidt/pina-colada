

export const SET_INDUSTRIES = "SET_INDUSTRIES";
export const SET_SALARY_RANGES = "SET_SALARY_RANGES";
export const SET_PROJECTS = "SET_PROJECTS";
export const ADD_INDUSTRY = "ADD_INDUSTRY";























export default (initialState) => {
  return (state, action) => {
    switch (action.type) {
      case SET_INDUSTRIES:
        return setIndustries(state, action.payload);
      case SET_SALARY_RANGES:
        return setSalaryRanges(state, action.payload);
      case SET_PROJECTS:
        return setProjects(state, action.payload);
      case ADD_INDUSTRY:
        return addIndustry(state, action.payload);
      default:
        return state;
    }
  };
};

const setIndustries = (state, payload) => {
  return {
    ...state,
    industries: payload.sort((a, b) => a.name.localeCompare(b.name)),
    loading: { ...state.loading, industries: false },
    loaded: { ...state.loaded, industries: true },
  };
};

const setSalaryRanges = (state, payload) => {
  return {
    ...state,
    salaryRanges: payload,
    loading: { ...state.loading, salaryRanges: false },
    loaded: { ...state.loaded, salaryRanges: true },
  };
};

const setProjects = (state, payload) => {
  return {
    ...state,
    projects: payload,
    loading: { ...state.loading, projects: false },
    loaded: { ...state.loaded, projects: true },
  };
};

const addIndustry = (state, payload) => {
  const updated = [...state.industries, payload].sort((a, b) => a.name.localeCompare(b.name));
  return {
    ...state,
    industries: updated,
  };
};
