const BandBg = ({ className = "", dark = false }) => {
  return (
    <div aria-hidden className={`pointer-events-none absolute inset-0 ${className}`}>
      {/* subtle grid */}
      <div className={`absolute inset-0 bg-[linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:24px_24px] ${dark ? "text-zinc-300 opacity-[0.15]" : "text-zinc-900 opacity-[0.06]"}`} />
      {/* vignette (very light) */}
      <div className="absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]" />
    </div>
  );
};

export default BandBg;
