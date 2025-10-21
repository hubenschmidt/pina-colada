import { type PropsWithChildren, type HTMLAttributes } from "react";
import Link from "next/link";

import cx from "../util/concat-classes";

type CardProps = PropsWithChildren<HTMLAttributes<HTMLDivElement>>;

export function Card({ className, children, ...rest }: CardProps) {
  return (
    <div
      {...rest}
      className={cx(
        // base
        "group rounded-2xl border border-zinc-200/60 bg-white/70 backdrop-blur-[1px] shadow-[0_1px_0_0_rgba(0,0,0,0.04)_inset]",
        // hover
        "transition-all hover:border-lime-400/50 hover:shadow-[0_8px_24px_-12px_rgba(0,0,0,0.15)] hover:-translate-y-[1px]",
        // user classes last so they override
        className
      )}
    >
      {children}
    </div>
  );
}

export function CardLink({
  href,
  children,
  className,
  newTab = true,
}: PropsWithChildren<{ href: string; className?: string; newTab?: boolean }>) {
  return (
    <Link
      href={href}
      target={newTab ? "_blank" : undefined}
      rel={newTab ? "noopener noreferrer" : undefined}
      className="block focus:outline-none focus-visible:ring-2 focus-visible:ring-lime-400/60 rounded-2xl"
      aria-label={typeof children === "string" ? children : undefined}
    >
      <Card
        className={cx(
          "p-6 hover:bg-zinc-50 hover:-translate-y-[1px] transition-colors",
          className
        )}
      >
        {children}
      </Card>
    </Link>
  );
}

export function SectionTitle({
  kicker,
  right,
  className,
  ...rest
}: {
  kicker?: string;
  right?: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      {...rest}
      className={cx("mb-8 flex items-end justify-between gap-4", className)}
    >
      <div>
        {kicker && (
          <div className="text-xs uppercase tracking-[0.18em] text-blue-500">
            {kicker}
          </div>
        )}
      </div>
      <div className="hidden md:flex items-center gap-3">
        {right}
        <span className="h-px w-28 bg-gradient-to-r from-lime-400/70 via-yellow-400/70 to-transparent" />
      </div>
    </div>
  );
}
