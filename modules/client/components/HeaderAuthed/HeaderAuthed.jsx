"use client";

import Link from "next/link";
import { useUserContext } from "../../context/userContext";

const HeaderAuthed = () => {
  const { userState } = useUserContext();

  if (!userState.isAuthed) return null;

  return (
    <div className="flex items-center gap-2">
      <Link
        href="/chat"
        className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-4 py-1.5 text-xs font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
      >
        Chat
      </Link>
    </div>
  );
};

export default HeaderAuthed;
