"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

export default function Header() {
  const onHome = usePathname() === "/";

  return (
    <header className="sticky top-0 z-40 border-b border-zinc-200 backdrop-blur supports-[backdrop-filter]:bg-white/100">
      <div className="mx-auto max-w-6xl px-4 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2">
          <span className="text-2xl leading-none">🍍</span>
          <span className="font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">
            PinaColada.co
          </span>
        </Link>

        <nav className="hidden sm:flex items-center gap-6 text-sm text-zinc-600">
          <Link href="/about" className="hover:text-black">
            About
          </Link>

          {onHome ? (
            <>
              <a href="#services" className="hover:text-black">
                Software Development
              </a>
              <a href="#ai" className="hover:text-black">
                AI
              </a>
              <a href="#approach" className="hover:text-black">
                Approach
              </a>
              <a href="#portfolio" className="hover:text-black">
                Portfolio
              </a>
              <a href="#contact" className="hover:text-black">
                Contact
              </a>
            </>
          ) : (
            <>
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
            </>
          )}

          <Link
            href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
            className="inline-flex h-9 items-center rounded-full border border-zinc-300 px-4 text-sm font-medium hover:border-lime-400/60 hover:text-black"
          >
            Start a project
          </Link>
        </nav>
      </div>
    </header>
  );
}
