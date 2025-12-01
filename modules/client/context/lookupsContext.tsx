"use client";

import { createContext, useReducer, useContext } from "react";
import lookupsReducer, { LookupsState, LookupsAction } from "../reducers/lookupsReducer";

const initialState: LookupsState = {
  industries: [],
  salaryRanges: [],
  projects: [],
  loading: {
    industries: false,
    salaryRanges: false,
    projects: false,
  },
  loaded: {
    industries: false,
    salaryRanges: false,
    projects: false,
  },
};

export const LookupsContext = createContext<{
  lookupsState: LookupsState;
  dispatchLookups: React.Dispatch<LookupsAction>;
}>({
  lookupsState: initialState,
  dispatchLookups: () => {},
});

export const useLookupsContext = () => useContext(LookupsContext);

export const LookupsProvider = ({ children }: { children: React.ReactNode }) => {
  const reducer = lookupsReducer(initialState);
  const [lookupsState, dispatchLookups] = useReducer(reducer, initialState);

  return (
    <LookupsContext.Provider value={{ lookupsState, dispatchLookups }}>
      {children}
    </LookupsContext.Provider>
  );
};
