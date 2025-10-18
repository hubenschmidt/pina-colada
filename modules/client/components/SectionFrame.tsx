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

import cx from "../util/concat-classes";

const SectionFrame = ({
  id,
  bandBg = "",
  className,
  children,
  roundedClassName = "rounded-2xl",
  containerClassName = "mx-auto max-w-6xl px-4",
}: SectionFrameProps) => {
  return (
    <section id={id} className={cx("py-20", bandBg)}>
      <div
        className={cx(
          containerClassName,
          roundedClassName,
          "bg-white shadow-sm",
          "p-5 sm:p-6 md:p-8",
          className
        )}
      >
        {children}
      </div>
    </section>
  );
};

export default SectionFrame;
