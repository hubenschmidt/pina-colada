"use client";
import Link from "next/link";
import { useCallback, useEffect } from "react";

export default function Home() {
  const scrollWithOffset = useCallback((hash: string, offset = 64) => {
    const el = document.querySelector(hash);
    if (!el) return;
    const y =
      (el as HTMLElement).getBoundingClientRect().top +
      window.pageYOffset -
      offset;
    window.scrollTo({ top: y, behavior: "smooth" });
    // Update the hash without jumping
    history.pushState(null, "", hash);
  }, []);

  // Handle direct loads with a hash and back/forward nav
  useEffect(() => {
    const handle = () => {
      if (location.hash) {
        // Timeout lets the layout settle before measuring
        setTimeout(() => scrollWithOffset(location.hash), 0);
      }
    };
    handle();
    window.addEventListener("hashchange", handle);
    return () => window.removeEventListener("hashchange", handle);
  }, [scrollWithOffset]);
  return (
    <div className="min-h-screen bg-black text-zinc-200 selection:bg-lime-400/30">
      {/* Top Nav */}
      <header className="sticky top-0 z-40 backdrop-blur supports-[backdrop-filter]:bg-black/40">
        <div className="mx-auto max-w-6xl px-4 py-4 flex items-center justify-between">
          <Link href="#" className="flex items-center gap-2">
            <span className="text-2xl leading-none">🍍</span>
            <span className="font-semibold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-lime-400 via-yellow-300 to-lime-400">
              PinaColada.co
            </span>
          </Link>
          <nav className="hidden sm:flex items-center gap-6 text-sm text-zinc-300">
            <Link
              href="#services"
              className="hover:text-white"
              onClick={(e) => {
                e.preventDefault();
                scrollWithOffset("#services", 64);
              }}
            >
              Software Development
            </Link>
            <Link
              href="#ai"
              className="hover:text-white"
              onClick={(e) => {
                e.preventDefault();
                scrollWithOffset("#ai", 64);
              }}
            >
              AI
            </Link>
            <Link
              href="#approach"
              className="hover:text-white"
              onClick={(e) => {
                e.preventDefault();
                scrollWithOffset("#approach", 64);
              }}
            >
              Approach
            </Link>
            <Link
              href="#portfolio"
              className="hover:text-white"
              onClick={(e) => {
                e.preventDefault();
                scrollWithOffset("#approach", 64);
              }}
            >
              Portfolio
            </Link>
            <Link
              href="#contact"
              className="hover:text-white"
              onClick={(e) => {
                e.preventDefault();
                scrollWithOffset("#contact", 64);
              }}
            >
              Contact
            </Link>
            <Link
              href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
              className="inline-flex h-9 items-center rounded-full border border-zinc-800 px-4 text-sm font-medium hover:border-lime-400/60 hover:text-white"
            >
              Start a project
            </Link>
          </nav>
        </div>
      </header>

      {/* Hero */}
      <section className="relative overflow-hidden">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-30"
        >
          {/* subtle grid */}
          <div className="[mask-image:radial-gradient(ellipse_at_center,black,transparent_65%)] absolute inset-0 bg-[linear-gradient(to_right,transparent_95%,rgba(255,255,255,0.06)_96%),linear-gradient(to_bottom,transparent_95%,rgba(255,255,255,0.06)_96%)] bg-[size:24px_24px]"></div>
          {/* lime/yellow glow */}
          <div className="absolute -top-24 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-lime-400/20 blur-3xl"></div>
          <div className="absolute -bottom-24 right-12 h-72 w-72 rounded-full bg-yellow-300/10 blur-3xl"></div>
        </div>
        <div className="mx-auto max-w-6xl px-4 py-24 sm:py-32">
          <h1 className="max-w-3xl text-balance text-4xl font-semibold tracking-tight sm:text-6xl">
            Elegant, enterprise‑grade
            <span className="block text-transparent bg-clip-text bg-gradient-to-r from-lime-400 via-yellow-300 to-lime-400">
              AI & software solutions consulting
            </span>
          </h1>
          <p className="mt-6 max-w-2xl text-pretty text-zinc-400">
            We design and build robust systems for small and medium‑sized
            businesses— from greenfield apps to integrations that tame complex
            CRM/ERP landscapes.
          </p>
          <div className="mt-8 flex flex-col sm:flex-row gap-3">
            <Link
              href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
              className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-black shadow-[0_0_0_1px_rgba(255,255,255,0.06)_inset] hover:brightness-95"
            >
              Book a free consult
            </Link>
            <Link
              href="#services"
              className="inline-flex items-center justify-center rounded-full border border-zinc-800 px-6 py-3 text-sm font-medium text-zinc-200 hover:border-lime-400/60"
              onClick={(e) => {
                e.preventDefault();
                scrollWithOffset("#services", 64);
              }}
            >
              Explore services
            </Link>
          </div>
          <div className="mt-10 flex flex-wrap items-center gap-3 text-xs text-zinc-400">
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              Next.js
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              React
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              Node
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              Postgres
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              Azure DevOps
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              Docker
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              LangGraph
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              OpenAI Agents SDK
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              MCP
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              RAG
            </span>
            <span className="rounded-full border border-zinc-800 px-3 py-1">
              Azure
            </span>
          </div>
        </div>
      </section>

      {/* Services */}
      <section id="services" className="mx-auto max-w-6xl px-4 py-20">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight">Services</h2>
          <span className="h-px w-24 bg-gradient-to-r from-lime-400/60 via-yellow-300/60 to-transparent" />
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[
            {
              title: "Custom Software Development",
              points: [
                "Greenfield builds",
                "Event‑driven & real‑time",
                "Secure by design",
              ],
            },
            {
              title: "Full‑Stack Development",
              points: ["Next.js/React UIs", "Node & APIs", "Testing & CI/CD"],
            },
            {
              title: "Solutions Consulting (SMBs)",
              points: [
                "Roadmaps & audits",
                "Cost‑effective modernization",
                "Buy vs. build",
              ],
            },
            {
              title: "CRM & ERP Systems",
              points: [
                "Plex, Dynamics, Salesforce",
                "Customization & extensions",
                "Data hygiene",
              ],
            },
            {
              title: "Systems Integrations",
              points: ["ETL & streaming", "Microservices", "3rd‑party APIs"],
            },
            {
              title: "Web Development",
              points: [
                "E-commerce integration",
                "Marketing sites",
                "Dashboards & portals",
                "Accessibility",
              ],
            },
          ].map((s) => (
            <div
              key={s.title}
              className="group rounded-2xl border border-zinc-900 bg-zinc-950 p-5 transition-colors hover:border-lime-400/40"
            >
              <div className="mb-3 flex items-start justify-between">
                <h3 className="text-base font-medium text-white">{s.title}</h3>
                <span className="mt-1 h-2 w-2 rounded-full bg-gradient-to-r from-lime-400 to-yellow-300 opacity-70 group-hover:opacity-100" />
              </div>
              <ul className="space-y-1 text-sm text-zinc-400">
                {s.points.map((p) => (
                  <li key={p} className="flex items-center gap-2">
                    <span className="h-1.5 w-1.5 rounded-full bg-lime-400/70" />{" "}
                    {p}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </section>

      {/* Agentic AI */}
      <section id="ai" className="mx-auto max-w-6xl px-4 pb-10">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight">
            Agentic AI Development
          </h2>
          <span className="h-px w-24 bg-gradient-to-r from-lime-400/60 via-yellow-300/60 to-transparent" />
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-zinc-900 bg-zinc-950 p-6">
            <div className="text-lg font-medium text-white">
              Agent Workflows
            </div>
            <p className="mt-2 text-sm text-zinc-400">
              Design and implement graph-based, tool-using agents with
              LangGraph, OpenAI Agents SDK, and MCP. Multimodal where it moves
              the needle.
            </p>
          </div>
          <div className="rounded-2xl border border-zinc-900 bg-zinc-950 p-6">
            <div className="text-lg font-medium text-white">
              RAG & Knowledge
            </div>
            <p className="mt-2 text-sm text-zinc-400">
              Retrieval-augmented pipelines with embeddings, chunking, and
              evaluators for accuracy, latency, and cost.
            </p>
          </div>
          <div className="rounded-2xl border border-zinc-900 bg-zinc-950 p-6">
            <div className="text-lg font-medium text-white">
              Production & Governance
            </div>
            <p className="mt-2 text-sm text-zinc-400">
              CI/CD on Azure DevOps, observability, guardrails, SOC2-ready auth,
              and secure data paths.
            </p>
          </div>
        </div>
      </section>

      {/* Approach */}
      <section id="approach" className="mx-auto max-w-6xl px-4 pb-20">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight">Approach</h2>
          <span className="h-px w-24 bg-gradient-to-r from-yellow-300/60 via-lime-400/60 to-transparent" />
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          {[
            {
              k: "1",
              h: "Discover",
              p: "Quickly map goals, constraints, and ROI. We favor lean specs and high‑signal prototypes.",
            },
            {
              k: "2",
              h: "Build",
              p: "Ship in vertical slices with strong CI/CD, observability, and performance budgets.",
            },
            {
              k: "3",
              h: "Scale",
              p: "Harden for production: security, infra, and support so your team can move fast.",
            },
          ].map((a) => (
            <div
              key={a.k}
              className="rounded-2xl border border-zinc-900 bg-zinc-950 p-6"
            >
              <div className="mb-2 text-sm text-zinc-500">Step {a.k}</div>
              <div className="text-lg font-medium text-white">{a.h}</div>
              <p className="mt-2 text-sm text-zinc-400">{a.p}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Portfolio */}
      <section id="portfolio" className="mx-auto max-w-6xl px-4 pb-20">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight">Portfolio</h2>
          <span className="h-px w-24 bg-gradient-to-r from-yellow-300/60 via-lime-400/60 to-transparent" />
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-zinc-900 bg-zinc-950 p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-white">
                Helios — B2B Integration Platform
              </h3>
              <span className="h-2 w-2 rounded-full bg-lime-400/80" />
            </div>
            <p className="mt-2 text-sm text-zinc-400">
              Multitenant iPaaS adopted by enterprise manufacturers.
              Domain-driven microservices, React/Node, MSSQL, Docker, Azure
              DevOps. SOC2-ready auth (Okta/JWT/MFA) and sub-2s IoT edge syncs.
            </p>
            <div className="mt-3 text-xs text-zinc-500">
              React • Node • MSSQL • Azure • Docker • IoT
            </div>
          </div>
          <div className="rounded-2xl border border-zinc-900 bg-zinc-950 p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-white">
                TuneCrook — DJ Music Discovery
              </h3>
              <span className="h-2 w-2 rounded-full bg-yellow-300/80" />
            </div>
            <p className="mt-2 text-sm text-zinc-400">
              Agentic AI curates tracks from Discogs and YouTube with RAG. Built
              with React, Node, Postgres; deployed with Azure DevOps.
            </p>
            <div className="mt-3 text-xs text-zinc-500">
              RAG • Agents • React • Node • Postgres
            </div>
          </div>
        </div>
      </section>

      {/* CTA / Contact */}
      <section id="contact" className="relative">
        <div className="mx-auto max-w-6xl px-4 pb-28">
          <div className="rounded-3xl border border-zinc-900 bg-zinc-950 p-8 md:p-10">
            <div className="flex flex-col items-start justify-between gap-6 md:flex-row md:items-center">
              <div>
                <h3 className="text-xl font-semibold text-white">
                  Have a project in mind?
                </h3>
                <p className="mt-1 text-sm text-zinc-400">
                  Let's align on goals and craft a pragmatic plan.
                </p>
              </div>
              <div className="flex flex-col sm:flex-row gap-3">
                <Link
                  href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
                  className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-black hover:brightness-95"
                >
                  Email us
                </Link>
                <Link
                  href="https://cal.com/your-handle/intro"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center justify-center rounded-full border border-zinc-800 px-6 py-3 text-sm font-medium text-zinc-200 hover:border-lime-400/60"
                >
                  Book a 20‑min intro
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-zinc-900">
        <div className="mx-auto max-w-6xl px-4 py-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-zinc-400">
          <div className="flex items-center gap-2">
            <span>© {new Date().getFullYear()} PinaColada.co</span>
          </div>
          <div className="flex items-center gap-4">
            <Link
              href="https://github.com/hubenschmidt"
              target="_blank"
              className="hover:text-white"
            >
              GitHub
            </Link>
            <Link
              href="https://www.linkedin.com/company/pinacoladaco"
              target="_blank"
              className="hover:text-white"
            >
              LinkedIn
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
