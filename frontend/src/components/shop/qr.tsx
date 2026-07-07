import { cn } from "@/lib/utils";

const SIZE = 15;

export function QrHashArt({ hash, className }: { hash: string; className?: string }) {
  const cells: boolean[] = [];
  for (let i = 0; i < SIZE * SIZE; i++) {
    const char = hash.charCodeAt((i * 7) % hash.length) || 0;
    cells.push(((char >> i % 5) & 1) === 1);
  }

  const isFinder = (r: number, c: number) => {
    const inBox = (br: number, bc: number) =>
      r >= br && r < br + 3 && c >= bc && c < bc + 3;
    return inBox(0, 0) || inBox(0, SIZE - 3) || inBox(SIZE - 3, 0);
  };

  return (
    <div
      className={cn(
        "grid aspect-square w-full gap-px rounded-lg bg-white p-2",
        className,
      )}
      style={{ gridTemplateColumns: `repeat(${SIZE}, minmax(0, 1fr))` }}
    >
      {cells.map((on, i) => {
        const r = Math.floor(i / SIZE);
        const c = i % SIZE;
        const filled = isFinder(r, c) || on;
        return (
          <span
            key={i}
            className={cn(
              "rounded-[1px]",
              filled ? "bg-ink" : "bg-transparent",
            )}
          />
        );
      })}
    </div>
  );
}
