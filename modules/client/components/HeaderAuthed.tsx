"use client";

import Link from "next/link";
import { useUserContext } from "../context/userContext";

const HeaderAuthed = () => {
  const { userState } = useUserContext();

  if (!userState.isAuthed) return null;

  return (
    <>
      <Link
        href="/chat"
        className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
      >
        Chat
      </Link>
    </>
  );
};

export default HeaderAuthed;
