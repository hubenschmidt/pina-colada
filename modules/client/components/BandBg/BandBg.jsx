"use client";
import { useMemo } from "react";

const BandBg = ({ className = "", dark = false, sweep = false }) => {
  const gridSize = 24;

  // Fibonacci sequence for subtle pattern
  const fibCells = useMemo(() => {
    if (!sweep) return [];

    // Fibonacci + fill rows to cover full height
    const fibCols = [1, 2, 3, 5, 8, 13, 21, 34];
    const fibRows = [1, 2, 3, 5, 8, 13, 21, 34, 40, 45, 50, 55];
    const cells = [];

    // Place cells at Fibonacci column positions across all rows
    fibRows.forEach((fRow) => {
      fibCols.forEach((fCol) => {
        cells.push({ x: fCol * gridSize, y: fRow * gridSize });
      });
    });

    return cells;
  }, [sweep]);

  // Generate sporadic cells that increase toward bottom-right
  const sweepCells = useMemo(() => {
    if (!sweep) return [];

    const cells = [];
    const cols = 40;
    const rows = 35;

    // Seeded random for consistent SSR/CSR
    const seededRandom = (seed) => {
      const x = Math.sin(seed * 9999) * 10000;
      return x - Math.floor(x);
    };

    for (let row = 0; row < rows; row++) {
      for (let col = 0; col < cols; col++) {
        // Probability increases toward bottom-right
        const xFactor = col / cols; // 0 to 1, higher = more right
        const yFactor = row / rows; // 0 to 1, higher = more bottom

        // Sweep from bottom-right: combine factors with right-side bias
        const probability = (xFactor * 0.7 + yFactor * 0.3) * 0.15;

        const seed = row * cols + col + 42;
        if (seededRandom(seed) < probability) {
          cells.push({
            x: col * gridSize,
            y: row * gridSize,
            opacity: 0.08 + seededRandom(seed + 1) * 0.12,
          });
        }
      }
    }
    return cells;
  }, [sweep]);

  return (
    <div aria-hidden className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`}>
      {/* subtle grid */}
      <div className={`absolute inset-0 bg-[linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:24px_24px] ${dark ? "text-zinc-300 opacity-[0.15]" : "text-zinc-900 opacity-[0.06]"}`} />

      {/* Fibonacci grid - very subtle */}
      {sweep && fibCells.map((cell, i) => (
        <div
          key={`fib-${i}`}
          className="absolute bg-zinc-500"
          style={{
            left: cell.x,
            top: cell.y,
            width: 23,
            height: 23,
            opacity: 0.08,
          }}
        />
      ))}

      {/* Sweep effect: sporadic cells increasing toward bottom-right */}
      {sweep && sweepCells.map((cell, i) => (
        <div
          key={i}
          className="absolute bg-zinc-600"
          style={{
            left: cell.x,
            top: cell.y,
            width: 23,
            height: 23,
            opacity: cell.opacity + 0.3,
          }}
        />
      ))}

      {/* vignette (very light) */}
      <div className="absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black,transparent_70%)]" />
    </div>
  );
};

export default BandBg;
