"use client";

import Image from "next/image";
import { useUserContext } from "../../context/userContext";
import Header from "../Header/Header";
import { Sidebar } from "../Sidebar/Sidebar";

const AuthenticatedPageLayout = ({
  children


}) => {
  const { userState } = useUserContext();
  const { isLoading, isAuthed } = userState;

  if (isLoading || !isAuthed) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image src="/loading-icon.png" alt="Loading" width={200} height={200} unoptimized />
      </div>);

  }

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <Header />
        <main className="flex-1 overflow-auto p-6">{children}</main>
      </div>
    </div>);

};

export default AuthenticatedPageLayout;