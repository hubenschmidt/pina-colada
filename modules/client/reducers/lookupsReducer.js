export const SET_INDUSTRIES = "SET_INDUSTRIES";
export const SET_SALARY_RANGES = "SET_SALARY_RANGES";
export const SET_PROJECTS = "SET_PROJECTS";
export const SET_REVENUE_RANGES = "SET_REVENUE_RANGES";
export const SET_EMPLOYEE_COUNT_RANGES = "SET_EMPLOYEE_COUNT_RANGES";
export const SET_FUNDING_STAGES = "SET_FUNDING_STAGES";
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
      case SET_REVENUE_RANGES:
        return setRevenueRanges(state, action.payload);
      case SET_EMPLOYEE_COUNT_RANGES:
        return setEmployeeCountRanges(state, action.payload);
      case SET_FUNDING_STAGES:
        return setFundingStages(state, action.payload);
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

const setRevenueRanges = (state, payload) => {
  return {
    ...state,
    revenueRanges: payload,
    loading: { ...state.loading, revenueRanges: false },
    loaded: { ...state.loaded, revenueRanges: true },
  };
};

const setEmployeeCountRanges = (state, payload) => {
  return {
    ...state,
    employeeCountRanges: payload,
    loading: { ...state.loading, employeeCountRanges: false },
    loaded: { ...state.loaded, employeeCountRanges: true },
  };
};

const setFundingStages = (state, payload) => {
  return {
    ...state,
    fundingStages: payload,
    loading: { ...state.loading, fundingStages: false },
    loaded: { ...state.loaded, fundingStages: true },
  };
};
