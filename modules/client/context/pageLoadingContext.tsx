"use client";
import { createContext, useReducer, useContext, ReactNode } from "react";
import pageLoadingReducer, {
  PageLoadingState,
} from "../reducers/pageLoadingReducer";

const initialState: PageLoadingState = {
  isPageLoading: true,
  redirect: true,
};

export const PageLoadingContext = createContext<{
  pageLoadingState: PageLoadingState;
  dispatchPageLoading: React.Dispatch<{ type: string; payload?: any }>;
}>({
  pageLoadingState: initialState,
  dispatchPageLoading: () => {},
});

export const usePageLoading = () => useContext(PageLoadingContext);

export const PageLoadingProvider = ({ children }: { children: ReactNode }) => {
  const reducer = pageLoadingReducer(initialState);
  const [pageLoadingState, dispatchPageLoading] = useReducer(
    reducer,
    initialState
  );

  return (
    <PageLoadingContext.Provider
      value={{ pageLoadingState, dispatchPageLoading }}
    >
      {children}
    </PageLoadingContext.Provider>
  );
};
