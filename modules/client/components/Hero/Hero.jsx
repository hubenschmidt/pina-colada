"use client";
import Link from "next/link";

const Hero = () => {
  return (
    <section className="relative overflow-hidden">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <div
        className="absolute inset-0 bg-cover bg-center bg-fixed opacity-20 [filter:brightness(0.98)_contrast(1.1)]"
        style={{ backgroundImage: "url('/pc.jpg')" }}
        aria-hidden="true"
      />

      <div className="relative z-10 mx-auto max-w-6xl px-4 py-24 sm:py-32">
        <h1 className="animate-fade-in-up max-w-3xl text-balance text-4xl font-semibold tracking-tight text-orange-400 sm:text-6xl leading-[1.1] md:leading-[1.05]">
          Elegant, enterprise
          <span className="block text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 pb-2 animate-gradient">
            AI & software solutions
          </span>
        </h1>
        <p className="animate-fade-in-up-delay-1 text-xl mt-6 max-w-2xl text-pretty text-zinc-900">
          We design and build robust systems — from greenfield apps to integrations that tame
          complex CRM/ERP landscapes.
        </p>
        <div className="animate-fade-in-up-delay-2 mt-8 flex flex-col sm:flex-row gap-3">
          <Link
            href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
            className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-zinc-900 shadow-[0_0_0_1px_rgba(0,0,0,0.04)_inset] hover:scale-105 hover:shadow-lg transition-all duration-200">
            Book a free consult
          </Link>
          <Link
            href="/#services"
            className="inline-flex items-center justify-center rounded-full border border-zinc-300 px-6 py-3 text-sm font-semibold text-zinc-900 hover:border-lime-400/60 hover:scale-105 transition-all duration-200">
            Explore services
          </Link>
        </div>
      </div>
    </section>
  );
};

export default Hero;
