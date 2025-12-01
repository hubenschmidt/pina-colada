import type { Industry, SalaryRange, Project } from "../api";

export const SET_INDUSTRIES = "SET_INDUSTRIES";
export const SET_SALARY_RANGES = "SET_SALARY_RANGES";
export const SET_PROJECTS = "SET_PROJECTS";
export const ADD_INDUSTRY = "ADD_INDUSTRY";

export interface LookupsState {
  industries: Industry[];
  salaryRanges: SalaryRange[];
  projects: Project[];
  loading: {
    industries: boolean;
    salaryRanges: boolean;
    projects: boolean;
  };
  loaded: {
    industries: boolean;
    salaryRanges: boolean;
    projects: boolean;
  };
}

export type LookupsAction =
  | { type: typeof SET_INDUSTRIES; payload: Industry[] }
  | { type: typeof SET_SALARY_RANGES; payload: SalaryRange[] }
  | { type: typeof SET_PROJECTS; payload: Project[] }
  | { type: typeof ADD_INDUSTRY; payload: Industry };

export default (initialState: LookupsState) => {
  return (state: LookupsState, action: LookupsAction): LookupsState => {
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

const setIndustries = (state: LookupsState, payload: Industry[]): LookupsState => {
  return {
    ...state,
    industries: payload.sort((a, b) => a.name.localeCompare(b.name)),
    loading: { ...state.loading, industries: false },
    loaded: { ...state.loaded, industries: true },
  };
};

const setSalaryRanges = (state: LookupsState, payload: SalaryRange[]): LookupsState => {
  return {
    ...state,
    salaryRanges: payload,
    loading: { ...state.loading, salaryRanges: false },
    loaded: { ...state.loaded, salaryRanges: true },
  };
};

const setProjects = (state: LookupsState, payload: Project[]): LookupsState => {
  return {
    ...state,
    projects: payload,
    loading: { ...state.loading, projects: false },
    loaded: { ...state.loaded, projects: true },
  };
};

const addIndustry = (state: LookupsState, payload: Industry): LookupsState => {
  const updated = [...state.industries, payload].sort((a, b) => a.name.localeCompare(b.name));
  return {
    ...state,
    industries: updated,
  };
};
