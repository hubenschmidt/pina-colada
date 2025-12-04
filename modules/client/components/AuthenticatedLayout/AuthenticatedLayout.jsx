"use client";

import Image from "next/image";
import { useUserContext } from "../../context/userContext";
import { Sidebar } from "../Sidebar/Sidebar";

export const AuthenticatedLayout = ({ children }) => {
  const { userState } = useUserContext();
  const { isLoading, isAuthed } = userState;

  if (isLoading || !isAuthed) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image
          src="/loading-icon.png"
          alt="Loading"
          width={200}
          height={200}
          unoptimized
        />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 p-6">{children}</main>
    </div>
  );
};
