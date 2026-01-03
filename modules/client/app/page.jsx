"use client";
import React, { useEffect } from "react";
import Link from "next/link";
import { CheckCircle2, ChevronRight } from "lucide-react";
import Hero from "../components/Hero/Hero";
import Header from "../components/Header/Header";
import BandBg from "../components/BandBg/BandBg";
import useScrollReveal from "../hooks/useScrollReveal";

const Home = () => {
  const servicesReveal = useScrollReveal();
  const aiReveal = useScrollReveal();
  const approachReveal = useScrollReveal();
  const contactReveal = useScrollReveal();

  useEffect(() => {
    let ticking = false;

    const onScroll = () => {
      if (ticking) return;
      ticking = true;

      requestAnimationFrame(() => {
        const nearTop = window.scrollY <= 2;
        if (nearTop && window.location.hash) {
          history.replaceState(null, "", window.location.pathname + window.location.search);
        }
        ticking = false;
      });
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <>
      <Header />
      <div className="min-h-screen text-zinc-200">
        <Hero />

        {/* Software Development */}
        <section
          ref={servicesReveal.ref}
          id="services"
          className={`relative bg-white transition-all duration-700 ${servicesReveal.isVisible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          {/* Top fade from white */}
          <div className="absolute top-0 left-0 right-0 h-32 bg-gradient-to-b from-white to-transparent pointer-events-none z-10" />
          <BandBg />
          <div className="relative mx-auto max-w-6xl px-4 py-24">
            <p className="mb-6 text-lg text-pretty text-zinc-600">
              We design and build reliable technical solutions — from greenfield apps to high-performance integrations
            </p>
            <p className="mb-6 text-lg text-pretty text-zinc-600">
              to unlock value and unleash efficiency in complex business, manufacturing and consumer landscapes.
            </p>
            <p className="mb-6 text-lg text-pretty text-yellow-500">
              Software and agentic AI that just works for you.
            </p>
            <h2 className="mb-12 text-3xl font-semibold tracking-tight text-zinc-900">
              Software Development
            </h2>
            <div className="grid gap-x-12 gap-y-10 sm:grid-cols-2 lg:grid-cols-3">
              {[
                {
                  title: "Custom Software",
                  points: ["Greenfield builds", "Event-driven & real-time", "Secure by design"],
                },
                {
                  title: "Full-Stack",
                  points: ["Next.js/React UIs", "Node & APIs", "Testing & CI/CD"],
                },
                {
                  title: "Solutions Consulting (SMBs)",
                  points: ["Roadmaps & audits", "Cost-effective modernization", "Buy vs. build"],
                },
                {
                  title: "CRM & ERP Systems",
                  points: ["Plex, Salesforce", "Customization & extensions", "Data hygiene"],
                },
                {
                  title: "Systems Integrations",
                  points: ["ETL & streaming", "Microservices", "3rd-party APIs"],
                },
                {
                  title: "Web",
                  points: ["E-commerce integration", "Marketing sites", "Dashboards & portals"],
                },
              ].map((s) => (
                <div key={s.title} className="rounded-lg p-4 -m-4 hover:shadow-lg hover:-translate-y-1 transition-all duration-200">
                  <h3 className="mb-3 text-lg font-semibold text-zinc-900">{s.title}</h3>
                  <ul className="space-y-2 text-sm text-zinc-600">
                    {s.points.map((p) => (
                      <li key={p} className="flex items-center gap-2">
                        <CheckCircle2 className="size-4 text-lime-500 animate-pulse-subtle" />
                        <span>{p}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* AI Development */}
        <section
          ref={aiReveal.ref}
          id="ai"
          className={`relative border-t border-zinc-200 bg-zinc-50 transition-all duration-700 ${aiReveal.isVisible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <BandBg />
          <div className="relative mx-auto max-w-6xl px-4 py-24">
            <h2 className="mb-12 text-3xl font-semibold tracking-tight text-zinc-900">
              AI Development
            </h2>
            <div className="grid gap-x-12 gap-y-10 md:grid-cols-3">
              {[
                {
                  title: "Agentic AI",
                  desc: "Graph-based agents with LangGraph, OpenAI Agents SDK, MCP, and custom tooling.",
                },
                {
                  title: "RAG & Knowledge",
                  desc: "Retrieval-augmented pipelines with evaluators for accuracy, latency, and cost.",
                },
                {
                  title: "Production & Governance",
                  desc: "CI/CD on Azure DevOps, observability, guardrails, and secure data paths.",
                },
              ].map((s) => (
                <div key={s.title} className="rounded-lg p-4 -m-4 hover:shadow-lg hover:-translate-y-1 transition-all duration-200">
                  <h3 className="mb-3 text-lg font-semibold text-zinc-900">{s.title}</h3>
                  <p className="text-sm text-zinc-600">{s.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Approach */}
        <section
          ref={approachReveal.ref}
          id="approach"
          className={`relative border-t border-zinc-200 bg-white transition-all duration-700 ${approachReveal.isVisible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <BandBg />
          <div className="relative mx-auto max-w-6xl px-4 py-24">
            <h2 className="mb-12 text-3xl font-semibold tracking-tight text-zinc-900">
              Our Approach
            </h2>
            <div className="flex flex-col gap-8 md:flex-row md:items-start md:gap-4">
              {[
                {
                  num: "1",
                  title: "Discover",
                  desc: "Clearly map goals, constraints, and ROI. We favor lean specs and rapid prototypes.",
                },
                {
                  num: "2",
                  title: "Build & Ship",
                  desc: "Continuously integrate feedback to respond quickly to your customer needs.",
                },
                {
                  num: "3",
                  title: "Scale",
                  desc: "Cloud solutions for security, infra, and support to optimize your software.",
                },
              ].map((s, i) => (
                <React.Fragment key={s.title}>
                  <div className="flex-1 rounded-lg p-4 -m-4 hover:shadow-lg hover:-translate-y-1 transition-all duration-200">
                    <div className="mb-3 flex items-center gap-3">
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-zinc-900 text-sm font-bold text-white">
                        {s.num}
                      </span>
                      <h3 className="text-lg font-semibold text-zinc-900">{s.title}</h3>
                    </div>
                    <p className="text-sm text-zinc-600">{s.desc}</p>
                  </div>
                  {i < 2 && (
                    <div className="hidden md:flex items-center justify-center">
                      <ChevronRight className="size-5 text-zinc-300" />
                    </div>
                  )}
                </React.Fragment>
              ))}
            </div>
          </div>
        </section>

        {/* Contact */}
        <section
          ref={contactReveal.ref}
          id="contact"
          className={`border-t border-zinc-700 bg-zinc-900 text-white transition-all duration-700 ${contactReveal.isVisible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}>
          <div className="mx-auto max-w-6xl px-4 py-24">
            <div className="flex flex-col items-start justify-between gap-6 md:flex-row md:items-center">
              <div>
                <h3 className="text-2xl font-semibold">Have a project in mind?</h3>
                <p className="mt-2 text-zinc-400">
                  Let's connect and talk about goals for your technical domain.
                </p>
              </div>
              <Link
                href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
                className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 px-6 py-3 text-sm font-semibold text-zinc-900 hover:scale-105 hover:shadow-lg transition-all duration-200">
                Email William
              </Link>
            </div>
          </div>
        </section>

        {/* Footer */}
        <footer className="border-t border-zinc-800 bg-zinc-900">
          <div className="mx-auto max-w-6xl px-4 py-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-zinc-400">
            <span>© {new Date().getFullYear()} PinaColada.co</span>
            <div className="flex items-center gap-4">
              <Link
                href="https://github.com/hubenschmidt"
                target="_blank"
                className="hover:text-white transition-colors duration-200">
                GitHub
              </Link>
              <Link
                href="https://www.linkedin.com/company/pinacoladaco"
                target="_blank"
                className="hover:text-white transition-colors duration-200">
                LinkedIn
              </Link>
            </div>
          </div>
        </footer>
      </div>
    </>
  );
};

export default Home;
