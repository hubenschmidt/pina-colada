"use client";
import Link from "next/link";
import Image from "next/image";
import { Github, Linkedin, LogIn, ChevronsDown } from "lucide-react";
import BandBg from "../BandBg/BandBg";

const Hero = () => {
  return (
    <section className="relative flex flex-col md:flex-row min-h-screen">
      {/* Left Pane - Hero Image (faded) */}
      <div className="relative w-full md:w-1/2 h-[40vh] md:h-screen flex items-center justify-center md:justify-end overflow-hidden bg-zinc-100">
        {/* Background image layer - faded */}
        <div
          className="absolute inset-0 bg-cover bg-center opacity-20"
          style={{ backgroundImage: "url('/pc.jpg')" }}
        />
        {/* BandBg grid for visual continuity */}
        <BandBg />
        {/* Bottom fade to cream - left pane only */}
        <div className="absolute bottom-0 left-0 right-0 h-48 bg-gradient-to-b from-transparent via-amber-50/50 to-amber-50 pointer-events-none z-10" />
        {/* Text content */}
        <div className="relative z-10 animate-fade-in-up text-center md:text-right pr-0 md:pr-12 px-6 py-4">
          <p className="pb-1">
            <span className="bg-amber-50/90 px-4 py-1 rounded inline-block">
              <span className="text-lg md:text-4xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">elegant software</span>
            </span>
          </p>
          <p className="pb-1">
            <span className="bg-amber-50/90 px-4 py-1 rounded inline-block">
              <span className="text-xl md:text-5xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">refreshing solutions</span>
            </span>
          </p>
          <p>
            <span className="bg-amber-50/90 px-4 py-1 rounded inline-block">
              <span className="text-2xl md:text-6xl font-semibold tracking-tight text-orange-400">business simplified</span><span className="text-2xl md:text-6xl font-semibold tracking-tight text-yellow-400">.</span>
            </span>
          </p>
        </div>
      </div>

      {/* Right Pane - Content */}
      <div className="relative w-full md:w-1/2 bg-zinc-800 flex items-center justify-center md:justify-start py-12 md:py-0">
        <BandBg dark />
        <div className="relative z-10 animate-fade-in-up-delay-1 flex flex-col items-center md:items-start gap-5 pl-0 md:pl-12">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-3">
            <Image src="/loading-icon.png" alt="PinaColada.co" width={64} height={64} priority className="shrink-0" />
            <span className="text-3xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">
              PinaColada
            </span>
          </Link>



          {/* Nav Links */}
          <nav className="flex flex-col gap-2 text-md font-semibold">
            <Link href="/#services" className="text-zinc-300 hover:text-white transition-colors duration-200">
              Engineering
            </Link>
            <Link href="/#approach" className="text-zinc-300 hover:text-white transition-colors duration-200">
              Our Approach
            </Link>
            <Link href="/#contact" className="text-zinc-300 hover:text-white transition-colors duration-200">
              Free Consultation
            </Link>
          </nav>

          {/* Social Icons - aligned with logo icon */}
          <div className="flex items-center gap-3">
            <Link
              href="https://github.com/hubenschmidt"
              target="_blank"
              className="text-zinc-300 hover:text-white transition-colors duration-200">
              <Github size={18} />
            </Link>
            <Link
              href="https://www.linkedin.com/in/williamhubenschmidt"
              target="_blank"
              className="text-zinc-300 hover:text-white transition-colors duration-200">
              <Linkedin size={18} />
            </Link>
            <Link
              href="/login"
              className="flex items-center gap-2 text-zinc-300 hover:text-white transition-colors duration-200 text-sm font-semibold">
              <LogIn size={18} />
            </Link>
          </div>
        </div>
      </div>

      {/* Scroll indicator arrow */}
      <Link
        href="/#services"
        className="absolute bottom-8 left-1/2 -translate-x-1/2 z-20 animate-bounce">
        <ChevronsDown
          size={48}
          className="text-yellow-400 drop-shadow-[0_0_8px_rgba(250,204,21,0.5)]"
        />
      </Link>
    </section>
  );
};

export default Hero;
