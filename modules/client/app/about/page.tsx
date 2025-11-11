import Image from "next/image";
import BandBg from "../../components/BandBg";

export const metadata = { title: "About — PinaColada.co" };

const AboutPage = () => {
  return (
    <section className="relative overflow-hidden">
      <main className="mx-auto max-w-6xl px-4 py-16">
        <BandBg className="z-0" />

        <div className="mb-10">
          <h1 className="text-3xl font-semibold tracking-tight text-blue-800">
            About
          </h1>
        </div>

        <div className="grid gap-8 md:grid-cols-[240px_1fr] items-start">
          {/* Portrait */}
          <div className="flex items-center justify-center">
            <div className="relative h-60 w-60 overflow-hidden rounded-2xl border border-zinc-200 shadow-sm mt-1">
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
          <div className="relative z-10 rounded-2xl bg-white p-6 shadow-md ring-1 ring-zinc-200 mt-1">
            <p className="text-blue-800 leading-relaxed">
              I’m William Hubenschmidt, a principal software engineer and
              consultant focused on enterprise-grade systems, integrations, and
              agentic AI. I help SMBs and mid-market teams ship dependable
              software—fast—by balancing architecture discipline with pragmatic
              execution.
            </p>
            <p className="mt-4 text-blue-800 leading-relaxed">
              Recent work includes multitenant iPaaS (Helios), AI-assisted music
              discovery (TuneCrook), and modernization projects across CRM/ERP
              ecosystems. If you’re planning a new build or taming legacy
              complexity, I’d love to chat.
            </p>

            <div className="mt-6">
              <a
                href="mailto:whubenschmidt@gmail.com?subject=Project%20Inquiry"
                className="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 hover:text-blue-500"
              >
                Get in touch
              </a>
            </div>
          </div>
        </div>
      </main>
    </section>
  );
};

export default AboutPage
