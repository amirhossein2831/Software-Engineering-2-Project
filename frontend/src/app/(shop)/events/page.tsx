"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { listEvents } from "@/lib/api/catalog";
import { EventCard } from "@/components/shop/event-card";
import { Input } from "@/components/ui/input";
import { LoadingState, EmptyState } from "@/components/ui/spinner";

export default function EventsPage() {
  const [genre, setGenre] = useState("");
  const [location, setLocation] = useState("");

  const { data, isLoading, isError } = useQuery({
    queryKey: ["events", genre, location],
    queryFn: () => listEvents({ genre, location }),
  });

  const events = data?.events ?? [];

  return (
    <div className="space-y-8">
      <section className="rounded-xl border border-line bg-gradient-to-br from-brand-soft to-surface p-8">
        <h1 className="text-3xl font-bold tracking-tight text-ink">
          Find your next live experience
        </h1>
        <p className="mt-2 max-w-xl text-muted">
          Browse events, pick your seats on a live map, and check out in
          seconds — with your seats held while you decide.
        </p>
        <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:max-w-xl">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
            <Input
              placeholder="Genre (e.g. concert)"
              value={genre}
              onChange={(e) => setGenre(e.target.value)}
              className="pl-9"
            />
          </div>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
            <Input
              placeholder="Location"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              className="pl-9"
            />
          </div>
        </div>
      </section>

      {isLoading ? (
        <LoadingState label="Loading events…" />
      ) : isError ? (
        <EmptyState
          title="Could not load events"
          description="The catalog service may be unavailable. Try again shortly."
        />
      ) : events.length === 0 ? (
        <EmptyState
          title="No events found"
          description="Try clearing your filters or check back soon."
        />
      ) : (
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {events.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      )}
    </div>
  );
}
