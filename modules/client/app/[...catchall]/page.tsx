"use client";

import { useEffect } from "react";
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
      <p className="text-zinc-600">Redirecting...</p>
    </div>
  );
};

export default CatchAllPage;
