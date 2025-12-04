"use client";
import React, { useEffect } from "react";
import Link from "next/link";
import SectionFrame from "../components/SectionFrame/SectionFrame";
import { CheckCircle2, ChevronRight } from "lucide-react";
import { Card, SectionTitle, CardLink } from "../components/ui/ui";
import Hero from "../components/Hero/Hero";
import BandBg from "../components/BandBg/BandBg";
import Chat from "../components/Chat/Chat";
import { useNavContext } from "../context/navContext";
import Header from "../components/Header/Header";

const Home = () => {
  const { navState } = useNavContext();
  const { agentOpen } = navState;

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

  useEffect(() => {
    document.getElementById("agent")?.scrollIntoView({ behavior: "smooth" });
  }, [agentOpen]);

  return (
    <>
      <Header />
      <div className="min-h-screen text-zinc-800 selection:bg-lime-300/40">
        <Hero />
        <section className="relative overflow-hidden">
          <BandBg />

          {/* Keep Chat mounted, but hide when not open */}
          <div style={{ display: agentOpen ? "block" : "none" }}>
            <SectionFrame id="agent" bandBg="bg-blue-200">
              <Chat />
            </SectionFrame>
          </div>

          <SectionFrame id="services" bandBg="bg-blue-200">
            <SectionTitle kicker="Software and AI Development" />
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {[
              {
                title: "Custom Software Development",
                points: [
                "Greenfield builds",
                "Event-driven & real-time",
                "Secure by design"]

              },
              {
                title: "Full-Stack Development",
                points: [
                "Next.js/React UIs",
                "Node & APIs",
                "Testing & CI/CD"]

              },
              {
                title: "Solutions Consulting (SMBs)",
                points: [
                "Roadmaps & audits",
                "Cost-effective modernization",
                "Buy vs. build"]

              },
              {
                title: "CRM & ERP Systems",
                points: [
                "Plex, Salesforce",
                "Customization & extensions",
                "Data hygiene"]

              },
              {
                title: "Systems Integrations",
                points: [
                "ETL & streaming",
                "Microservices",
                "3rd-party APIs"]

              },
              {
                title: "Web Development",
                points: [
                "E-commerce integration",
                "Marketing sites",
                "Dashboards & portals"]

              }].
              map((s) =>
              <Card key={s.title} className="p-5 transition-shadow hover:shadow-md">
                  <div className="mb-3 flex items-start justify-between">
                    <h3 className="text-xl font-semibold text-blue-800">{s.title}</h3>
                    <span className="mt-1 h-2 w-2 rounded-full bg-gradient-to-r from-lime-400 to-yellow-400 opacity-80" />
                  </div>
                  <ul className="space-y-1.5 text-sm text-blue-800">
                    {s.points.map((p) =>
                  <li key={p} className="flex items-center gap-2">
                        <CheckCircle2 className="size-4 text-lime-500/80" />
                        <span>{p}</span>
                      </li>
                  )}
                  </ul>
                </Card>
              )}
            </div>

            <div id="ai" className="mx-auto max-w-6xl px-4 py-20">
              <div className="mb-10 flex items-center justify-between">
                <h2 className="text-2xl font-semibold tracking-tight text-blue-800">
                  AI Development
                </h2>
                <span className="h-px w-24 bg-gradient-to-r from-lime-400/60 via-yellow-400/60 to-transparent" />
              </div>
              <div className="grid gap-4 md:grid-cols-3">
                {[
                {
                  title: "Agentic AI",
                  points: [
                  "Graph-based agents with LangGraph, OpenAI Agents SDK, MCP, and custom tooling."]

                },
                {
                  title: "RAG & Knowledge",
                  points: [
                  "Retrieval-augmented pipelines with evaluators for accuracy, latency, and cost."]

                },
                {
                  title: "Production & Governance",
                  points: [
                  "CI/CD on Azure DevOps, observability, guardrails, and secure data paths."]

                }].
                map((s) =>
                <Card key={s.title} className="p-5 transition-shadow hover:shadow-md">
                    <div className="mb-3 flex items-start justify-between">
                      <h3 className="text-xl font-semibold text-blue-800">{s.title}</h3>
                      <span className="mt-1 h-2 w-2 rounded-full bg-gradient-to-r from-lime-400 to-yellow-400 opacity-80" />
                    </div>
                    <ul className="space-y-1.5 text-sm text-blue-800">
                      {s.points.map((p) =>
                    <li key={p} className="flex items-center gap-2">
                          <span>{p}</span>
                        </li>
                    )}
                    </ul>
                  </Card>
                )}
              </div>
            </div>
          </SectionFrame>

          <SectionFrame id="approach" bandBg="bg-blue-200">
            <SectionTitle kicker="Our Approach" />
            <div className="mb-10 flex items-center justify-between">
              <h2 className="text-2xl font-semibold tracking-tight text-blue-800">
                Continuous Integration and Delivery
              </h2>
            </div>
            <div className="flex flex-col gap-4 md:flex-row md:items-stretch">
              {[
              {
                k: "1",
                title: "Discover",
                points: [
                "Clearly map goals, constraints, and ROI. We favor lean specs and rapid prototypes."]

              },
              {
                k: "2",
                title: "Build and Ship",
                points: [
                "Continuously integrate feedback to respond quickly to your customer needs."]

              },
              {
                k: "3",
                title: "Scale",
                points: [
                "Cloud solutions for security, infra, and support to optimize your software."]

              }].
              map((s, i) =>
              <React.Fragment key={s.title}>
                  <Card className="flex-1 p-5 transition-shadow hover:shadow-md">
                    <div className="mb-3 flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <span className="flex h-7 w-7 items-center justify-center rounded-full bg-gradient-to-r from-lime-400 to-yellow-400 text-sm font-bold text-blue-900">
                          {s.k}
                        </span>
                        <h3 className="text-xl font-semibold text-blue-800">{s.title}</h3>
                      </div>
                    </div>
                    <ul className="space-y-1.5 text-sm text-blue-800">
                      {s.points.map((p) =>
                    <li key={p} className="flex items-center gap-2">
                          <span>{p}</span>
                        </li>
                    )}
                    </ul>
                  </Card>
                  {i < 2 &&
                <div className="hidden md:flex items-center justify-center px-2">
                      <ChevronRight className="size-6 text-blue-400" />
                    </div>
                }
                </React.Fragment>
              )}
            </div>
          </SectionFrame>

          <SectionFrame id="portfolio" bandBg="bg-blue-300">
            <SectionTitle kicker="Portfolio" />
            {[
            {
              k: "helios",
              href: "https://www.cumulus-erp.com/helios-ipaas/",
              title: "Helios — B2B Integration Platform",
              description: "iPaaS adopted by enterprise manufacturers."
            },
            {
              k: "langgraph-react",
              href: "https://github.com/hubenschmidt/langgraph-react",
              title: "langgraph-react",
              description:
              "enterprise-ready LangGraph + React starter rolled with Langfuse self-hosted LLM observability tracing"
            },
            {
              k: "tunecrook",
              href: "https://www.tunecrook.com/",
              title: "TuneCrook — DJ Music Discovery",
              description: "Agentic AI curates your Discogs collection."
            }].
            map((p) =>
            <CardLink key={p.k} href={p.href}>
                <div className="flex items-center justify-between">
                  <h3 className="text-2xl font-medium text-blue-800">
                    {p.title}
                  </h3>
                  <span
                  className={`h-2 w-2 rounded-full bg-gradient-to-r from-lime-400 to-yellow-400`} />

                </div>
                <p className="mt-2 text-sm text-blue-800">
                  {p.description}
                </p>
              </CardLink>
            )}
          </SectionFrame>
        </section>
        {/* Contact */}
        <section id="contact" className="relative bg-blue-800 text-blue-50">
          <div className="mx-auto max-w-6xl px-4 pt-20 pb-28">
            <div className="rounded-3xl border border-blue-600/60 bg-blue-700/40 p-8 md:p-10 shadow-sm">
              <div className="flex flex-col items-start justify-between gap-6 md:flex-row md:items-center">
                <div>
                  <h3 className="text-xl font-semibold text-white">
                    Have a project in mind?
                  </h3>
                  <p className="mt-1 text-sm text-blue-200">
                    Let's connect and talk about goals for your technical
                    domain.
                  </p>
                </div>
                <div className="flex flex-col sm:flex-row gap-3">
                  <Link
                    href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
                    className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500">

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
                className="hover:text-white">

                GitHub
              </Link>
              <Link
                href="https://www.linkedin.com/company/pinacoladaco"
                target="_blank"
                className="hover:text-white">

                LinkedIn
              </Link>
            </div>
          </div>
        </footer>
      </div>
    </>);

};

export default Home;