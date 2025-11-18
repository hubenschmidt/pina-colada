"use client";
import { createContext, useReducer, useContext } from "react";
import navReducer, { NavState } from "../reducers/navReducer";

const initialState: NavState = {
  agentOpen: false,
  sidebarCollapsed: true,
};

export const NavContext = createContext<{
  navState: NavState;
  dispatchNav: React.Dispatch<{ type: string; payload?: any }>;
}>({
  navState: initialState,
  dispatchNav: () => {},
});

export const useNavContext = () => useContext(NavContext);

export const NavProvider = ({children}: {children: React.ReactNode}) => {
  const reducer = navReducer(initialState);
  const [navState, dispatchNav] = useReducer(reducer, initialState);

  return (
    <NavContext.Provider value={{ navState, dispatchNav }}>
      {children}
    </NavContext.Provider>
  );
};
