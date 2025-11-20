"use client";

import Image from "next/image";
import { useUser } from "@auth0/nextjs-auth0/client";
import { useUserContext } from "../../context/userContext";
import { Sidebar } from "../Sidebar/Sidebar";

export const AuthenticatedLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { isLoading } = useUser();
  const { userState } = useUserContext();

  if (isLoading || !userState.isAuthed) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image src="/icon.png" alt="Loading" width={200} height={200} />
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
