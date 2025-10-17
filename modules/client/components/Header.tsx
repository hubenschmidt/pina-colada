"use client";

import Link from "next/link";
import Image from "next/image";

export default function Header() {
  return (
    <header className="sticky top-0 z-40 border-b border-zinc-200 backdrop-blur bg-blue-100">
      <div className="mx-auto max-w-6xl px-4 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2">
          <Image
            src="/logo.png" // lives in /public
            alt="PinaColada.co"
            width={48}
            height={48}
            priority // preloads on homepage
            className="rounded-md shadow-[0_1px_6px_rgba(0,0,0,0.12)]"
          />
          <span className="text-3xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">
            PinaColada.co
          </span>
        </Link>

        <nav className="hidden sm:flex items-center gap-6 text-sm text-zinc-600 font-semibold">
          <Link href="/about" className="hover:text-black">
            About
          </Link>

          <Link href="/#services" className="hover:text-black">
            Software Development
          </Link>
          <Link href="/#ai" className="hover:text-black">
            AI
          </Link>
          <Link href="/#approach" className="hover:text-black">
            Approach
          </Link>
          <Link href="/#portfolio" className="hover:text-black">
            Portfolio
          </Link>
          <Link href="/#contact" className="hover:text-black">
            Contact
          </Link>

          <Link
            href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
            className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-black hover:brightness-95"
          >
            Start a project
          </Link>
        </nav>
      </div>
    </header>
  );
}
