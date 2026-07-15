"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Plus, Pencil, Trash2, X } from "lucide-react";
import {
  listAllEvents,
  createEvent,
  updateEvent,
  deleteEvent,
  publishEvent,
  listVenues,
} from "@/lib/api/catalog";
import { ApiError } from "@/lib/api/client";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input, Select, Label } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import { formatDate } from "@/lib/utils";
import type { Event } from "@/lib/types";

const emptyForm = {
  title: "",
  description: "",
  genre: "",
  location: "",
  venue_id: "",
  starts_at: "",
};

function toLocalInput(iso: string) {
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(
    d.getHours(),
  )}:${pad(d.getMinutes())}`;
}

export default function AdminEventsPage() {
  const qc = useQueryClient();
  const [form, setForm] = useState(emptyForm);
  const [error, setError] = useState<string | null>(null);

  const eventsQuery = useQuery({
    queryKey: ["events", "all"],
    queryFn: listAllEvents,
  });
  const venuesQuery = useQuery({ queryKey: ["venues"], queryFn: listVenues });

  const create = useMutation({
    mutationFn: () =>
      createEvent({
        ...form,
        starts_at: new Date(form.starts_at).toISOString(),
      }),
    onSuccess: () => {
      setForm(emptyForm);
      setError(null);
      qc.invalidateQueries({ queryKey: ["events"] });
    },
    onError: (e) =>
      setError(e instanceof ApiError ? e.message : "Could not create event."),
  });

  const set =
    (k: keyof typeof form) =>
    (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setForm((f) => ({ ...f, [k]: e.target.value }));

  const events = eventsQuery.data?.events ?? [];
  const venues = venuesQuery.data?.venues ?? [];

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
              <Input value={form.description} onChange={set("description")} />
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
              <Label>Venue</Label>
              <Select value={form.venue_id} onChange={set("venue_id")} required>
                <option value="" disabled>
                  Select a venue…
                </option>
                {venues.map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.name}
                  </option>
                ))}
              </Select>
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
        <h2 className="text-base font-semibold">All events</h2>
        {eventsQuery.isLoading ? (
          <LoadingState />
        ) : events.length === 0 ? (
          <EmptyState
            title="No events yet"
            description="Create an event above, then publish it."
          />
        ) : (
          events.map((event) => <EventRow key={event.id} event={event} />)
        )}
      </div>
    </div>
  );
}

function EventRow({ event }: { event: Event }) {
  const qc = useQueryClient();
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState({
    title: event.title,
    description: event.description,
    genre: event.genre,
    location: event.location,
    starts_at: toLocalInput(event.starts_at),
  });
  const [error, setError] = useState<string | null>(null);

  const refetch = () => qc.invalidateQueries({ queryKey: ["events"] });
  const isDraft = event.status !== "published";

  const publish = useMutation({
    mutationFn: () => publishEvent(event.id),
    onSuccess: refetch,
  });

  const save = useMutation({
    mutationFn: () =>
      updateEvent(event.id, {
        ...draft,
        starts_at: new Date(draft.starts_at).toISOString(),
      }),
    onSuccess: () => {
      setEditing(false);
      setError(null);
      refetch();
    },
    onError: (e) =>
      setError(e instanceof ApiError ? e.message : "Could not save event."),
  });

  const remove = useMutation({
    mutationFn: () => deleteEvent(event.id),
    onSuccess: refetch,
    onError: (e) =>
      setError(e instanceof ApiError ? e.message : "Could not delete event."),
  });

  const setD =
    (k: keyof typeof draft) => (e: React.ChangeEvent<HTMLInputElement>) =>
      setDraft((d) => ({ ...d, [k]: e.target.value }));

  if (editing) {
    return (
      <Card className="space-y-3 p-4">
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="sm:col-span-2">
            <Label>Title</Label>
            <Input value={draft.title} onChange={setD("title")} />
          </div>
          <div className="sm:col-span-2">
            <Label>Description</Label>
            <Input value={draft.description} onChange={setD("description")} />
          </div>
          <div>
            <Label>Genre</Label>
            <Input value={draft.genre} onChange={setD("genre")} />
          </div>
          <div>
            <Label>Location</Label>
            <Input value={draft.location} onChange={setD("location")} />
          </div>
          <div>
            <Label>Starts at</Label>
            <Input
              type="datetime-local"
              value={draft.starts_at}
              onChange={setD("starts_at")}
            />
          </div>
        </div>
        {error && (
          <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-danger">
            {error}
          </p>
        )}
        <div className="flex gap-2">
          <Button size="sm" onClick={() => save.mutate()} disabled={save.isPending}>
            Save changes
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => {
              setEditing(false);
              setError(null);
            }}
          >
            <X className="h-4 w-4" />
            Cancel
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <Card className="p-4 transition-colors hover:bg-canvas">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="font-medium">{event.title}</p>
          <p className="text-sm text-muted">
            {event.location || "TBA"} · {formatDate(event.starts_at)}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <Badge tone={event.status === "published" ? "success" : "neutral"}>
            {event.status}
          </Badge>
          {isDraft && (
            <Button
              size="sm"
              variant="secondary"
              onClick={() => publish.mutate()}
              disabled={publish.isPending}
            >
              Publish
            </Button>
          )}
          <Button size="sm" variant="ghost" onClick={() => setEditing(true)}>
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            className="text-danger hover:bg-red-50 hover:text-danger disabled:opacity-40"
            onClick={() => remove.mutate()}
            disabled={!isDraft || remove.isPending}
            title={isDraft ? "Delete draft" : "Published events can't be deleted"}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
      {error && (
        <p className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-danger">
          {error}
        </p>
      )}
    </Card>
  );
}
