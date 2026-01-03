"use client";
import Link from "next/link";
import Image from "next/image";
import { Github, Linkedin, LogIn, ChevronsDown } from "lucide-react";

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
        {/* Text content */}
        <div
          className="relative z-10 animate-fade-in-up text-center md:text-right pr-0 md:pr-12 px-6 py-4 backdrop-blur-xs rounded-lg"
          style={{ maskImage: "radial-gradient(ellipse at center, black 50%, transparent 100%)", WebkitMaskImage: "radial-gradient(ellipse at center, black 50%, transparent 100%)" }}>
          <p className="text-xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">
            elegant software
          </p>
          <p className="text-2xl font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500">
            refreshing solutions
          </p>
          <p className="text-3xl font-semibold tracking-tight text-orange-400">
            business simplified.
          </p>
        </div>
      </div>

      {/* Right Pane - Content */}
      <div className="w-full md:w-1/2 bg-zinc-800 flex items-center justify-center md:justify-start py-12 md:py-0">
        <div className="animate-fade-in-up-delay-1 flex flex-col items-center md:items-start gap-5 pl-0 md:pl-12">
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
              href="https://www.linkedin.com/company/pinacoladaco"
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
          className="text-yellow-500 drop-shadow-[0_0_8px_rgba(132,204,22,0.5)]"
        />
      </Link>
    </section>
  );
};

export default Hero;
