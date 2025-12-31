const BandBg = ({ className = "" }) => {
  return (
    <div aria-hidden className={`pointer-events-none absolute inset-0 ${className}`}>
      {/* subtle grid */}
      <div className="absolute inset-0 opacity-[0.8] bg-[linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:24px_24px] text-zinc-900" />
    </div>
  );
};

export default BandBg;
