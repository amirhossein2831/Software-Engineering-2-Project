"use client";

import { useQuery } from "@tanstack/react-query";
import { CalendarDays, CheckCircle2, Tags, MapPin } from "lucide-react";
import { listEvents } from "@/lib/api/catalog";
import { StatCard } from "@/components/admin/stat-card";
import { Card } from "@/components/ui/card";
import { LoadingState } from "@/components/ui/spinner";

export default function DashboardPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["events", "", ""],
    queryFn: () => listEvents(),
  });

  if (isLoading) return <LoadingState label="Loading dashboard…" />;

  const events = data?.events ?? [];
  const published = events.filter((e) => e.status === "published").length;
  const genres = new Set(events.map((e) => e.genre).filter(Boolean)).size;
  const locations = new Set(events.map((e) => e.location).filter(Boolean)).size;

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
        <p className="mt-1 text-sm text-muted">
          A snapshot of your live catalog.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Live events" value={events.length} icon={CalendarDays} />
        <StatCard label="Published" value={published} icon={CheckCircle2} />
        <StatCard label="Genres" value={genres} icon={Tags} />
        <StatCard label="Locations" value={locations} icon={MapPin} />
      </div>

      <Card className="p-6">
        <h2 className="text-base font-semibold">Revenue & sales analytics</h2>
        <p className="mt-2 text-sm text-muted">
          Detailed sales, revenue, and remaining-capacity dashboards are served
          by the dedicated <span className="font-medium">Analytics</span>{" "}
          service (projected from the Kafka event stream). Wire{" "}
          <code className="rounded bg-canvas px-1.5 py-0.5 text-xs">
            /analytics/events/:id
          </code>{" "}
          here once that service is deployed.
        </p>
      </Card>
    </div>
  );
}
