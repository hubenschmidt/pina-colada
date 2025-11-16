"use client";

import Image from "next/image";
import { useUser } from "@auth0/nextjs-auth0/client";
import Header from "../components/Header";

export const RootLayoutClient = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { user, isLoading } = useUser();

  if (isLoading && !user) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image src="/icon.png" alt="Loading" width={200} height={200} />
      </div>
    );
  }

  return (
    <>
      <Header />
      {children}
    </>
  );
};
