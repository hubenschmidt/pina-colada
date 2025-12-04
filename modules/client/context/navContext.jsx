"use client";
import { createContext, useReducer, useContext } from "react";
import navReducer from "../reducers/navReducer";

const initialState = {
  agentOpen: false,
  sidebarCollapsed: false,
  sidebarSections: {
    accounts: true,
    leads: true,
    reports: true
  }
};

export const NavContext = createContext(


  {
    navState: initialState,
    dispatchNav: () => {}
  });

export const useNavContext = () => useContext(NavContext);

export const NavProvider = ({ children }) => {
  const reducer = navReducer(initialState);
  const [navState, dispatchNav] = useReducer(reducer, initialState);

  return (
    <NavContext.Provider value={{ navState, dispatchNav }}>
      {children}
    </NavContext.Provider>);

};