"use client";
import { createContext, useReducer, useContext } from "react";
import navReducer from "../reducers/navReducer";

const initialState = {
  agentOpen: false,
};

export const NavContext = createContext(initialState);
export const useNav = () => useContext(NavContext);

export const NavProvider = (props) => {
  const reducer = navReducer(initialState);
  const [navState, dispatchNav] = useReducer(reducer, initialState);

  return (
    <NavContext.Provider value={{ navState, dispatchNav }}>
      {props.children}
    </NavContext.Provider>
  );
};
