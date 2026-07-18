"use client";

import { useMemo } from "react";
import type { Seat, SeatStatus } from "@/lib/types";
import { cn } from "@/lib/utils";

interface SeatMapProps {
  seats: Seat[];
  statuses: Record<string, SeatStatus>;
  selected: Set<string>;
  onToggle: (seatId: string) => void;
}

const seatClasses: Record<SeatStatus | "selected", string> = {
  available: "bg-surface border-line hover:border-brand hover:bg-brand-soft",
  locked: "bg-amber-100 border-amber-200 cursor-not-allowed",
  booked: "bg-zinc-200 border-zinc-300 cursor-not-allowed",
  selected: "bg-brand border-brand text-brand-fg",
};

export function SeatMap({ seats, statuses, selected, onToggle }: SeatMapProps) {
  const rows = useMemo(() => {
    const grouped = new Map<string, Seat[]>();
    for (const seat of seats) {
      const list = grouped.get(seat.row_label) ?? [];
      list.push(seat);
      grouped.set(seat.row_label, list);
    }
    return [...grouped.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([row, list]) => ({
        row,
        seats: list.sort((a, b) => a.number - b.number),
      }));
  }, [seats]);

  return (
    <div className="space-y-4">
      <div className="mx-auto w-full max-w-md rounded-lg bg-gradient-to-b from-zinc-200 to-transparent py-2 text-center text-xs font-medium uppercase tracking-widest text-muted">
        Stage
      </div>

      <div className="space-y-2 overflow-x-auto">
        {rows.map(({ row, seats: rowSeats }) => (
          <div key={row} className="flex items-center gap-2">
            <span className="w-6 shrink-0 text-xs font-medium text-muted">
              {row}
            </span>
            <div className="flex flex-wrap gap-1.5">
              {rowSeats.map((seat) => {
                const status = statuses[seat.id] ?? "available";
                const isSelected = selected.has(seat.id);
                const interactive = status === "available";
                return (
                  <button
                    key={seat.id}
                    type="button"
                    disabled={!interactive && !isSelected}
                    onClick={() => interactive && onToggle(seat.id)}
                    title={`Row ${row}, Seat ${seat.number}`}
                    className={cn(
                      "grid h-8 w-8 place-items-center rounded-md border text-xs font-medium transition-colors",
                      isSelected
                        ? seatClasses.selected
                        : seatClasses[status],
                    )}
                  >
                    {seat.number}
                  </button>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-4 pt-2 text-xs text-muted">
        <Legend className="bg-surface border-line" label="Available" />
        <Legend className="bg-brand border-brand" label="Selected" />
        <Legend className="bg-amber-100 border-amber-200" label="On hold" />
        <Legend className="bg-zinc-200 border-zinc-300" label="Booked" />
      </div>
    </div>
  );
}

function Legend({ className, label }: { className: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5">
      <span className={cn("h-3.5 w-3.5 rounded border", className)} />
      {label}
    </span>
  );
}
