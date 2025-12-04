"use client";
import { createContext, useReducer, useContext } from "react";
import userReducer from "../reducers/userReducer";

const initialState = {
  user: null,
  tenantName: null,
  bearerToken: null,
  isAuthed: false,
  isLoading: true,
  theme: "light",
  canEditTenantTheme: false,
};

export const UserContext = createContext({
  userState: initialState,
  dispatchUser: () => {},
});

export const useUserContext = () => useContext(UserContext);

export const UserProvider = ({ children }) => {
  const reducer = userReducer(initialState);
  const [userState, dispatchUser] = useReducer(reducer, initialState);

  return (
    <UserContext.Provider value={{ userState, dispatchUser }}>
      {children}
    </UserContext.Provider>
  );
};
