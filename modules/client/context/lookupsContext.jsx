"use client";

import { createContext, useReducer, useContext } from "react";
import lookupsReducer from "../reducers/lookupsReducer";

const initialState = {
  industries: [],
  salaryRanges: [],
  projects: [],
  revenueRanges: [],
  employeeCountRanges: [],
  fundingStages: [],
  loading: {
    industries: false,
    salaryRanges: false,
    projects: false,
    revenueRanges: false,
    employeeCountRanges: false,
    fundingStages: false,
  },
  loaded: {
    industries: false,
    salaryRanges: false,
    projects: false,
    revenueRanges: false,
    employeeCountRanges: false,
    fundingStages: false,
  },
};

export const LookupsContext = createContext({
  lookupsState: initialState,
  dispatchLookups: () => {},
});

export const useLookupsContext = () => useContext(LookupsContext);

export const LookupsProvider = ({ children }) => {
  const reducer = lookupsReducer(initialState);
  const [lookupsState, dispatchLookups] = useReducer(reducer, initialState);

  return (
    <LookupsContext.Provider value={{ lookupsState, dispatchLookups }}>
      {children}
    </LookupsContext.Provider>
  );
};
