// components/SectionFrame.tsx
type SectionFrameProps = {
  id?: string;
  bandBg?: string; // e.g. "bg-zinc-100"
  className?: string; // extra classes for the inner framed box
  children: React.ReactNode;
  borderClassName?: string; // e.g. "border border-orange-300"
  roundedClassName?: string; // e.g. "rounded-2xl"
  containerClassName?: string; // default container classes
};

// Tiny local helper to concat classes safely
function cx(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export default function SectionFrame({
  id,
  bandBg = "",
  className,
  children,
  borderClassName = "border-2 border-yellow-400",
  roundedClassName = "rounded-3xl",
  containerClassName = "mx-auto max-w-6xl px-4",
}: SectionFrameProps) {
  return (
    <section id={id} className={cx("py-20", bandBg)}>
      <div
        className={cx(
          containerClassName,
          roundedClassName,
          borderClassName,
          "bg-blue-100 shadow-sm",
          "p-5 sm:p-6 md:p-8",
          className
        )}
      >
        {children}
      </div>
    </section>
  );
}
