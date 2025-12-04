"use client";
import { createContext, useReducer, useContext } from "react";
import pageLoadingReducer from

"../reducers/pageLoadingReducer";

const initialState = {
  isPageLoading: true,
  redirect: true
};

export const PageLoadingContext = createContext(


  {
    pageLoadingState: initialState,
    dispatchPageLoading: () => {}
  });

export const usePageLoading = () => useContext(PageLoadingContext);

export const PageLoadingProvider = ({ children }) => {
  const reducer = pageLoadingReducer(initialState);
  const [pageLoadingState, dispatchPageLoading] = useReducer(
    reducer,
    initialState
  );

  return (
    <PageLoadingContext.Provider
      value={{ pageLoadingState, dispatchPageLoading }}>

      {children}
    </PageLoadingContext.Provider>);

};