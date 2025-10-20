"use client";

import Link from "next/link";
import Image from "next/image";

const Hero = () => {
  return (
    <section className="relative overflow-hidden">
      <Image
        src="/pc.png"
        alt=""
        fill
        priority
        aria-hidden
        className="object-cover object-center opacity-20 [filter:brightness(0.98)_contrast(1.1)]"
        sizes="100vw"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-0"
      >
        {/* subtle grid for light bg: darker lines, soft mask */}
        <div className="[mask-image:radial-gradient(ellipse_at_center,black,transparent_65%)] absolute inset-0 bg-[linear-gradient(to_right,transparent_95%,rgba(0,0,0,0.06)_96%),linear-gradient(to_bottom,transparent_95%,rgba(0,0,0,0.06)_96%)] bg-[size:24px_24px]" />
        <div className="absolute -top-24 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-lime-300/30 blur-3xl" />
        <div className="absolute -bottom-24 right-12 h-72 w-72 rounded-full bg-yellow-300/20 blur-3xl" />
      </div>
      <div className="relative z-10 mx-auto max-w-6xl px-4 py-24 sm:py-32">
        <h1 className="max-w-3xl text-balance text-4xl font-semibold tracking-tight text-orange-400 sm:text-6xl leading-[1.1] md:leading-[1.05]">
          Elegant, enterprise
          <span className="block text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 pb-2">
            AI & software solutions
          </span>
        </h1>

        <p className="text-xl mt-6 max-w-2xl text-pretty text-zinc-600">
          We design and build robust systems — from greenfield apps to
          integrations that tame complex CRM/ERP landscapes.
        </p>
        <div className="mt-8 flex flex-col sm:flex-row gap-3">
          <Link
            href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
            className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-black shadow-[0_0_0_1px_rgba(0,0,0,0.04)_inset] hover:brightness-95"
          >
            Book a free consult
          </Link>
          <Link
            href="#services"
            className="inline-flex items-center justify-center rounded-full border border-zinc-300 px-6 py-3 text-sm font-semibold text-zinc-800 hover:border-lime-400/60"
          >
            Explore services
          </Link>
        </div>
      </div>
    </section>
  );
};

export default Hero;
