"use client";
import { useEffect } from "react";
import Link from "next/link";
import SectionFrame from "../components/SectionFrame";

export default function Home() {
  // Clear the hash when user scrolls back to (near) top
  useEffect(() => {
    let ticking = false;

    const onScroll = () => {
      if (ticking) return;
      ticking = true;

      requestAnimationFrame(() => {
        const nearTop = window.scrollY <= 2; // small tolerance
        if (nearTop && window.location.hash) {
          // remove only the hash, keep path + query
          history.replaceState(
            null,
            "",
            window.location.pathname + window.location.search
          );
        }
        ticking = false;
      });
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <div className="min-h-screen text-zinc-800 selection:bg-lime-300/40">
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 opacity-40"
        >
          {/* subtle grid for light bg: darker lines, soft mask */}
          <div className="[mask-image:radial-gradient(ellipse_at_center,black,transparent_65%)] absolute inset-0 bg-[linear-gradient(to_right,transparent_95%,rgba(0,0,0,0.06)_96%),linear-gradient(to_bottom,transparent_95%,rgba(0,0,0,0.06)_96%)] bg-[size:24px_24px]" />
          {/* lime/yellow glow */}
          <div className="absolute -top-24 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full bg-lime-300/30 blur-3xl" />
          <div className="absolute -bottom-24 right-12 h-72 w-72 rounded-full bg-yellow-300/20 blur-3xl" />
        </div>
        <div className="mx-auto max-w-6xl px-4 py-24 sm:py-32">
          <h1 className="max-w-3xl text-balance text-4xl font-semibold tracking-tight text-orange-400 sm:text-6xl leading-[1.1] md:leading-[1.05]">
            Elegant, enterprise-grade
            <span className="block text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 leading-[1.05] pb-2">
              AI & software solutions consulting
            </span>
          </h1>

          <p className="mt-6 max-w-2xl text-pretty text-zinc-600">
            We design and build robust systems for small and medium-sized
            businesses— from greenfield apps to integrations that tame complex
            CRM/ERP landscapes.
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
              className="inline-flex items-center justify-center rounded-full border border-zinc-300 px-6 py-3 text-sm font-medium text-zinc-800 hover:border-lime-400/60"
            >
              Explore services
            </Link>
          </div>
        </div>
      </section>
      {/* Services — zinc-50 */}
      <SectionFrame id="services" bandBg="bg-blue-200">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
            Software Development
          </h2>
          <span className="h-px w-24 bg-gradient-to-r from-lime-400/60 via-yellow-400/60 to-transparent" />
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[
            {
              title: "Custom Software Development",
              points: [
                "Greenfield builds",
                "Event-driven & real-time",
                "Secure by design",
              ],
            },
            {
              title: "Full-Stack Development",
              points: ["Next.js/React UIs", "Node & APIs", "Testing & CI/CD"],
            },
            {
              title: "Solutions Consulting (SMBs)",
              points: [
                "Roadmaps & audits",
                "Cost-effective modernization",
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
              points: ["ETL & streaming", "Microservices", "3rd-party APIs"],
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
              className="group rounded-2xl border border-zinc-200 p-5 transition-colors hover:border-lime-400/40"
            >
              <div className="mb-3 flex items-start justify-between">
                <h3 className="text-base font-medium text-zinc-900">
                  {s.title}
                </h3>
                <span className="mt-1 h-2 w-2 rounded-full bg-gradient-to-r from-lime-400 to-yellow-400 opacity-70 group-hover:opacity-100" />
              </div>
              <ul className="space-y-1 text-sm text-zinc-600">
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

        {/* Agentic AI — zinc-100 */}
        <div id="ai" className="mx-auto max-w-6xl px-4 py-20">
          <div className="mb-10 flex items-center justify-between">
            <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
              AI Development
            </h2>
            <span className="h-px w-24 bg-gradient-to-r from-lime-400/60 via-yellow-400/60 to-transparent" />
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            {[
              {
                h: "Agent Workflows",
                p: "Design and implement graph-based, tool-using agents with LangGraph, OpenAI Agents SDK, and MCP. Multimodal where it moves the needle.",
              },
              {
                h: "RAG & Knowledge",
                p: "Retrieval-augmented pipelines with embeddings, chunking, and evaluators for accuracy, latency, and cost.",
              },
              {
                h: "Production & Governance",
                p: "CI/CD on Azure DevOps, observability, guardrails, SOC2-ready auth, and secure data paths.",
              },
            ].map((a) => (
              <div key={a.h} className="rounded-2xl border border-zinc-200 p-6">
                <div className="text-lg font-medium text-zinc-900">{a.h}</div>
                <p className="mt-2 text-sm text-zinc-600">{a.p}</p>
              </div>
            ))}
          </div>
        </div>
      </SectionFrame>
      {/* Approach — zinc-200 */}
      <SectionFrame id="approach" bandBg="bg-blue-200">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
            Approach
          </h2>
          <span className="h-px w-24 bg-gradient-to-r from-yellow-400/60 via-lime-400/60 to-transparent" />
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          {[
            {
              k: "1",
              h: "Discover",
              p: "Quickly map goals, constraints, and ROI. We favor lean specs and high-signal prototypes.",
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
            <div key={a.k} className="rounded-2xl border border-zinc-200 p-6">
              <div className="mb-2 text-sm text-zinc-500">Step {a.k}</div>
              <div className="text-lg font-medium text-zinc-900">{a.h}</div>
              <p className="mt-2 text-sm text-zinc-600">{a.p}</p>
            </div>
          ))}
        </div>
      </SectionFrame>
      {/* Portfolio — zinc-300 */}
      <SectionFrame id="portfolio" bandBg="bg-blue-300">
        <div className="mb-10 flex items-center justify-between">
          <h2 className="text-2xl font-semibold tracking-tight text-zinc-900">
            Portfolio
          </h2>
          <span className="h-px w-24 bg-gradient-to-r from-yellow-400/60 via-lime-400/60 to-transparent" />
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {/* Helios */}
          <div className="rounded-2xl border border-zinc-200 p-6 transition hover:border-zinc-300 hover:bg-zinc-50">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-zinc-900">
                <a
                  href="https://www.cumulus-erp.com/helios-ipaas/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:underline focus-visible:underline underline-offset-4 decoration-2"
                  aria-label="Open Helios — B2B Integration Platform"
                >
                  Helios — B2B Integration Platform
                </a>
              </h3>
              <span className="h-2 w-2 rounded-full bg-lime-400/80" />
            </div>
            <p className="mt-2 text-sm text-zinc-600">
              Multitenant iPaaS adopted by enterprise manufacturers.
              Domain-driven microservices, React/Node, MSSQL, Docker, Azure
              DevOps. SOC2-ready auth (Okta/JWT/MFA) and sub-2s IoT edge syncs.
            </p>
            <div className="mt-3 text-xs text-zinc-500">
              React • Node • MSSQL • Azure • Docker • IoT
            </div>
          </div>

          {/* TuneCrook */}
          <div className="rounded-2xl border border-zinc-200 p-6 transition hover:border-zinc-300 hover:bg-zinc-50">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-zinc-900">
                <a
                  href="https://www.tunecrook.com/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:underline focus-visible:underline underline-offset-4 decoration-2"
                  aria-label="Open TuneCrook — DJ Music Discovery"
                >
                  TuneCrook — DJ Music Discovery
                </a>
              </h3>
              <span className="h-2 w-2 rounded-full bg-yellow-300/80" />
            </div>
            <p className="mt-2 text-sm text-zinc-600">
              Agentic AI curates tracks from Discogs and YouTube with RAG. Built
              with React, Node, Postgres; deployed with Azure DevOps.
            </p>
            <div className="mt-3 text-xs text-zinc-500">
              RAG • Agents • React • Node • Postgres
            </div>
          </div>
        </div>
      </SectionFrame>
      {/* CTA / Contact */}
      <section id="contact" className="relative bg-blue-800 text-blue-50">
        <div className="mx-auto max-w-6xl px-4 pt-20 pb-28">
          <div className="rounded-3xl border border-blue-600/60 bg-blue-700/40 p-8 md:p-10 shadow-sm">
            <div className="flex flex-col items-start justify-between gap-6 md:flex-row md:items-center">
              <div>
                <h3 className="text-xl font-semibold text-white">
                  Have a project in mind?
                </h3>
                <p className="mt-1 text-sm text-blue-200">
                  Let's connect and talk about goals for your technical domain.
                </p>
              </div>
              <div className="flex flex-col sm:flex-row gap-3">
                <Link
                  href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
                  className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-black hover:brightness-95"
                >
                  Email William
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>
      {/* Footer */}
      <footer className="border-t border-blue-700 bg-blue-800">
        <div className="mx-auto max-w-6xl px-4 py-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-blue-200">
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
