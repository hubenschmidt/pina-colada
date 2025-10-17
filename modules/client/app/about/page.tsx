// app/about/page.tsx
import Image from "next/image";

export const metadata = { title: "About — PinaColada.co" };

export default function AboutPage() {
  return (
    <main className="mx-auto max-w-6xl px-4 py-16">
      <div className="mb-10">
        <h1 className="text-3xl font-semibold tracking-tight text-zinc-900">
          About
        </h1>
      </div>

      <div className="grid gap-8 md:grid-cols-[240px_1fr] items-start">
        {/* Portrait */}
        <div className="flex items-center justify-center">
          <div className="relative h-60 w-60 overflow-hidden rounded-2xl border border-zinc-200 shadow-sm">
            <Image
              src="/wh.jpg"
              alt="William Hubenschmidt"
              width={240}
              height={240}
              className="object-cover"
              priority
            />
          </div>
        </div>

        {/* Blurb */}
        <div>
          <p className="text-zinc-700 leading-relaxed">
            I’m William Hubenschmidt, a principal software engineer and
            consultant focused on enterprise-grade systems, integrations, and
            agentic AI. I help SMBs and mid-market teams ship dependable
            software—fast—by balancing architecture discipline with pragmatic
            execution.
          </p>
          <p className="mt-4 text-zinc-700 leading-relaxed">
            Recent work includes multitenant iPaaS (Helios), AI-assisted music
            discovery (TuneCrook), and modernization projects across CRM/ERP
            ecosystems. If you’re planning a new build or taming legacy
            complexity, I’d love to chat.
          </p>

          <div className="mt-6">
            <a
              href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
              className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-black hover:brightness-95"
            >
              Get in touch
            </a>
          </div>
        </div>
      </div>
    </main>
  );
}
