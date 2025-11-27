"use client";

import Image from "next/image";
import { useUser } from "@auth0/nextjs-auth0/client";
import { useUserContext } from "../context/userContext";
import Header from "./Header";
import { Sidebar } from "./Sidebar/Sidebar";

const AuthenticatedPageLayout = ({
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
    <div className="flex h-screen overflow-hidden">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        <Header />
        <main className="flex-1 overflow-auto p-6">{children}</main>
      </div>
    </div>
  );
};

export default AuthenticatedPageLayout;
