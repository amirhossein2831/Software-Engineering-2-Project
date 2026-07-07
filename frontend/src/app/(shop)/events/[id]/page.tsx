"use client";

import { use, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { CalendarDays, MapPin, Loader2 } from "lucide-react";
import { getEvent } from "@/lib/api/catalog";
import { seatStates, hold } from "@/lib/api/reservation";
import { ApiError } from "@/lib/api/client";
import { useRealtime } from "@/lib/hooks/use-realtime";
import { useSession } from "@/lib/hooks/use-session";
import { SeatMap } from "@/components/shop/seat-map";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import { formatDate, formatMoney } from "@/lib/utils";
import type { SeatStatus } from "@/lib/types";

export default function EventDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const router = useRouter();
  const { isAuthenticated } = useSession();

  const eventQuery = useQuery({
    queryKey: ["event", id],
    queryFn: () => getEvent(id),
  });
  const statesQuery = useQuery({
    queryKey: ["seat-states", id],
    queryFn: () => seatStates(id),
    refetchInterval: 15_000,
  });

  const [statuses, setStatuses] = useState<Record<string, SeatStatus>>({});
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!statesQuery.data) return;
    const next: Record<string, SeatStatus> = {};
    for (const s of statesQuery.data.seats) next[s.seat_id] = s.status;
    setStatuses(next);
  }, [statesQuery.data]);

  useRealtime([`event:${id}`], (msg) => {
    const seatIds = (msg.seat_ids as string[]) ?? [];
    const status: SeatStatus | null =
      msg.type === "seat.locked"
        ? "locked"
        : msg.type === "seat.booked"
          ? "booked"
          : msg.type === "seat.released"
            ? "available"
            : null;
    if (!status) return;
    setStatuses((prev) => {
      const next = { ...prev };
      for (const sid of seatIds) next[sid] = status;
      return next;
    });
    if (status !== "available") {
      setSelected((prev) => {
        const next = new Set(prev);
        for (const sid of seatIds) next.delete(sid);
        return next;
      });
    }
  });

  const detail = eventQuery.data;

  const priceBySector = useMemo(() => {
    const map = new Map<string, number>();
    for (const p of detail?.pricing ?? []) map.set(p.sector_id, p.amount);
    return map;
  }, [detail]);

  const currency = detail?.pricing[0]?.currency ?? "USD";

  const total = useMemo(() => {
    let sum = 0;
    for (const seat of detail?.seats ?? []) {
      if (selected.has(seat.id)) sum += priceBySector.get(seat.sector_id) ?? 0;
    }
    return sum;
  }, [detail, selected, priceBySector]);

  const toggle = (seatId: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(seatId)) next.delete(seatId);
      else next.add(seatId);
      return next;
    });

  async function onHold() {
    if (!isAuthenticated) {
      router.push(`/login?next=/events/${id}`);
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const admission =
        sessionStorage.getItem(`admission:${id}`) ?? undefined;
      const res = await hold(id, [...selected], admission);
      router.push(`/checkout/${res.id}`);
    } catch (e) {
      if (e instanceof ApiError && e.status === 403) {
        router.push(`/queue/${id}`);
        return;
      }
      setError(
        e instanceof ApiError ? e.message : "Could not hold those seats.",
      );
    } finally {
      setSubmitting(false);
    }
  }

  if (eventQuery.isLoading) return <LoadingState label="Loading event…" />;
  if (eventQuery.isError || !detail)
    return <EmptyState title="Event not found" />;

  const { event } = detail;

  return (
    <div className="grid gap-8 lg:grid-cols-[1fr_320px]">
      <div className="space-y-6">
        <div>
          <div className="mb-2 flex items-center gap-2">
            {event.genre && <Badge tone="brand">{event.genre}</Badge>}
            <Badge tone={event.status === "published" ? "success" : "neutral"}>
              {event.status}
            </Badge>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">{event.title}</h1>
          <p className="mt-2 text-muted">{event.description}</p>
          <div className="mt-4 flex flex-wrap gap-4 text-sm text-muted">
            <span className="flex items-center gap-2">
              <CalendarDays className="h-4 w-4" />
              {formatDate(event.starts_at)}
            </span>
            <span className="flex items-center gap-2">
              <MapPin className="h-4 w-4" />
              {event.location || "TBA"}
            </span>
          </div>
        </div>

        <Card className="p-6">
          <h2 className="mb-4 text-base font-semibold">Choose your seats</h2>
          {detail.seats.length === 0 ? (
            <EmptyState title="No seats configured for this event yet." />
          ) : (
            <SeatMap
              seats={detail.seats}
              statuses={statuses}
              selected={selected}
              onToggle={toggle}
            />
          )}
        </Card>
      </div>

      <aside className="lg:sticky lg:top-24 lg:h-fit">
        <Card className="p-6">
          <h2 className="text-base font-semibold">Your selection</h2>
          <p className="mt-1 text-sm text-muted">
            {selected.size === 0
              ? "No seats selected yet."
              : `${selected.size} seat${selected.size > 1 ? "s" : ""} selected`}
          </p>

          <div className="my-5 flex items-baseline justify-between border-t border-line pt-5">
            <span className="text-sm text-muted">Total</span>
            <span className="text-2xl font-bold">
              {formatMoney(total, currency)}
            </span>
          </div>

          {error && (
            <p className="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-danger">
              {error}
            </p>
          )}

          <Button
            className="w-full"
            size="lg"
            disabled={selected.size === 0 || submitting}
            onClick={onHold}
          >
            {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
            Hold seats & checkout
          </Button>
          <p className="mt-3 text-center text-xs text-muted">
            Seats are held for a limited time while you complete payment.
          </p>
        </Card>
      </aside>
    </div>
  );
}
