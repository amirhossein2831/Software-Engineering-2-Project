"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Plus } from "lucide-react";
import { listEvents, createEvent, publishEvent } from "@/lib/api/catalog";
import { ApiError } from "@/lib/api/client";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import { formatDate } from "@/lib/utils";

export default function AdminEventsPage() {
  const qc = useQueryClient();
  const [form, setForm] = useState({
    title: "",
    description: "",
    genre: "",
    location: "",
    venue_id: "",
    starts_at: "",
  });
  const [error, setError] = useState<string | null>(null);

  const eventsQuery = useQuery({
    queryKey: ["events", "", ""],
    queryFn: () => listEvents(),
  });

  const create = useMutation({
    mutationFn: () =>
      createEvent({
        ...form,
        starts_at: new Date(form.starts_at).toISOString(),
      }),
    onSuccess: () => {
      setForm({
        title: "",
        description: "",
        genre: "",
        location: "",
        venue_id: "",
        starts_at: "",
      });
      setError(null);
      qc.invalidateQueries({ queryKey: ["events"] });
    },
    onError: (e) =>
      setError(e instanceof ApiError ? e.message : "Could not create event."),
  });

  const publish = useMutation({
    mutationFn: (id: string) => publishEvent(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["events"] }),
  });

  const set = (k: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  const events = eventsQuery.data?.events ?? [];

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold tracking-tight">Events</h1>

      <Card>
        <CardHeader>
          <CardTitle>Create an event</CardTitle>
        </CardHeader>
        <CardBody>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              create.mutate();
            }}
            className="grid gap-4 sm:grid-cols-2"
          >
            <div className="sm:col-span-2">
              <Label>Title</Label>
              <Input value={form.title} onChange={set("title")} required />
            </div>
            <div className="sm:col-span-2">
              <Label>Description</Label>
              <Input
                value={form.description}
                onChange={set("description")}
              />
            </div>
            <div>
              <Label>Genre</Label>
              <Input value={form.genre} onChange={set("genre")} />
            </div>
            <div>
              <Label>Location</Label>
              <Input value={form.location} onChange={set("location")} />
            </div>
            <div>
              <Label>Venue ID</Label>
              <Input
                value={form.venue_id}
                onChange={set("venue_id")}
                placeholder="uuid from Venues"
                required
              />
            </div>
            <div>
              <Label>Starts at</Label>
              <Input
                type="datetime-local"
                value={form.starts_at}
                onChange={set("starts_at")}
                required
              />
            </div>
            {error && (
              <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-danger sm:col-span-2">
                {error}
              </p>
            )}
            <div className="sm:col-span-2">
              <Button type="submit" disabled={create.isPending}>
                {create.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Plus className="h-4 w-4" />
                )}
                Create event
              </Button>
            </div>
          </form>
        </CardBody>
      </Card>

      <div className="space-y-3">
        <h2 className="text-base font-semibold">Published events</h2>
        {eventsQuery.isLoading ? (
          <LoadingState />
        ) : events.length === 0 ? (
          <EmptyState
            title="No published events"
            description="Create an event above, then publish it."
          />
        ) : (
          events.map((event) => (
            <Card
              key={event.id}
              className="flex items-center justify-between p-4"
            >
              <div>
                <p className="font-medium">{event.title}</p>
                <p className="text-sm text-muted">
                  {event.location || "TBA"} · {formatDate(event.starts_at)}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <Badge
                  tone={event.status === "published" ? "success" : "neutral"}
                >
                  {event.status}
                </Badge>
                {event.status !== "published" && (
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => publish.mutate(event.id)}
                    disabled={publish.isPending}
                  >
                    Publish
                  </Button>
                )}
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  );
}
