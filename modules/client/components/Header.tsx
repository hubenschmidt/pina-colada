"use client";

import Link from "next/link";
import Image from "next/image";

const Header = () => {
  return (
    <header className="sticky top-0 z-40 border-b border-zinc-200 bg-blue-50">
      <div className="mx-auto max-w-6xl px-4 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2">
          <Image
            src="/logo.png"
            alt="PinaColada.co"
            width={48}
            height={48}
            priority
          />
          <span className="text-3xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">
            PinaColada
          </span>
        </Link>

        <nav className="hidden sm:flex items-center gap-6 text-sm text-zinc-600 font-semibold">
          <Link
            href="/#agent"
            className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
          >
            Chat with agent
          </Link>
          <Link href="/about" className="text-blue-700 hover:text-blue-500">
            About
          </Link>
          <Link href="/#services" className="text-blue-700 hover:text-blue-500">
            Software Development
          </Link>
          <Link href="/#ai" className="text-blue-700 hover:text-blue-500">
            AI
          </Link>
          <Link href="/#approach" className="text-blue-700 hover:text-blue-500">
            Approach
          </Link>
          <Link
            href="/#portfolio"
            className="text-blue-700 hover:text-blue-500"
          >
            Portfolio
          </Link>
          <Link href="/#contact" className="text-blue-700 hover:text-blue-500">
            Contact
          </Link>
          <Link
            href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
            className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
          >
            Start a project
          </Link>
        </nav>
      </div>
    </header>
  );
};

export default Header;
