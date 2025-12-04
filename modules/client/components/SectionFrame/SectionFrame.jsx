// components/SectionFrame.tsx










import cx from "../../lib/concat-classes";

const SectionFrame = ({
  id,
  bandBg = "",
  className,
  children,
  roundedClassName = "rounded-2xl",
  containerClassName = "mx-auto max-w-6xl px-4"
}) => {
  return (
    <section id={id} className={cx("py-15", bandBg)}>
      <div
        className={cx(
          containerClassName,
          roundedClassName,
          "bg-white shadow-sm",
          "p-5 sm:p-6 md:p-8",
          className
        )}>

        {children}
      </div>
    </section>);

};

export default SectionFrame;