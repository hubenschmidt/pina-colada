"use client";
import { createContext, useReducer, useContext, ReactNode } from "react";
import userReducer, { UserState } from "../reducers/userReducer";

const initialState: UserState = {
  user: null,
  tenantName: null,
  bearerToken: null,
  isAuthed: false,
  theme: "light",
  canEditTenantTheme: false,
};

export const UserContext = createContext<{
  userState: UserState;
  dispatchUser: React.Dispatch<{ type: string; payload?: any }>;
}>({
  userState: initialState,
  dispatchUser: () => {},
});

export const useUserContext = () => useContext(UserContext);

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const reducer = userReducer(initialState);
  const [userState, dispatchUser] = useReducer(reducer, initialState);

  return (
    <UserContext.Provider value={{ userState, dispatchUser }}>
      {children}
    </UserContext.Provider>
  );
};
