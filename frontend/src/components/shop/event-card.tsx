import Link from "next/link";
import { CalendarDays, MapPin } from "lucide-react";
import type { Event } from "@/lib/types";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { formatDate } from "@/lib/utils";

export function EventCard({ event }: { event: Event }) {
  return (
    <Link href={`/events/${event.id}`} className="group block">
      <Card className="h-full overflow-hidden transition-shadow group-hover:shadow-md">
        <div className="flex h-32 items-center justify-center bg-gradient-to-br from-brand/90 to-indigo-400 text-brand-fg">
          <span className="text-3xl font-bold tracking-tight">
            {event.title.slice(0, 1).toUpperCase()}
          </span>
        </div>
        <div className="p-5">
          <div className="mb-2 flex items-center gap-2">
            {event.genre && <Badge tone="brand">{event.genre}</Badge>}
          </div>
          <h3 className="line-clamp-1 text-base font-semibold text-ink">
            {event.title}
          </h3>
          <p className="mt-1 line-clamp-2 text-sm text-muted">
            {event.description || "Live event"}
          </p>
          <div className="mt-4 space-y-1.5 text-sm text-muted">
            <div className="flex items-center gap-2">
              <CalendarDays className="h-4 w-4" />
              {formatDate(event.starts_at)}
            </div>
            <div className="flex items-center gap-2">
              <MapPin className="h-4 w-4" />
              {event.location || "TBA"}
            </div>
          </div>
        </div>
      </Card>
    </Link>
  );
}
