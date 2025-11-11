const BandBg = ({ className = "" }: { className?: string }) => {
  return (
    <div
      aria-hidden
      className={`pointer-events-none absolute inset-0 ${className}`}
    >
      {/* subtle grid */}
      <div className="absolute inset-0 opacity-[0.08] bg-[linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:24px_24px] text-blue-900" />
      {/* vignette (very light) */}
      <div className="absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]" />
      {/* lime glow */}
      <div className="absolute -top-24 right-0 h-64 w-64 rounded-full bg-lime-300/30 blur-3xl" />
    </div>
  );
};

export default BandBg
