"use client";

import { useEffect } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { useUserContext } from "../../context/userContext";

const CatchAllPage = () => {
  const router = useRouter();
  const { userState } = useUserContext();

  useEffect(() => {
    if (userState.isAuthed) {
      router.replace("/chat");
      return;
    }

    router.replace("/");
  }, [userState.isAuthed, router]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <Image src="/loading-icon.png" alt="Redirecting" width={200} height={200} priority />
    </div>
  );
};

export default CatchAllPage;
