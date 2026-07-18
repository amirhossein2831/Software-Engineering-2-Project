"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Plus, Copy, Check, Trash2, Pencil, X } from "lucide-react";
import {
  listVenues,
  createVenue,
  updateVenue,
  deleteVenue,
  addSector,
  deleteSector,
} from "@/lib/api/catalog";
import { ApiError } from "@/lib/api/client";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import type { Venue } from "@/lib/types";

export default function AdminVenuesPage() {
  const qc = useQueryClient();
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [error, setError] = useState<string | null>(null);

  const venuesQuery = useQuery({ queryKey: ["venues"], queryFn: listVenues });
  const venues = venuesQuery.data?.venues ?? [];

  const create = useMutation({
    mutationFn: () => createVenue({ name, address }),
    onSuccess: () => {
      setName("");
      setAddress("");
      setError(null);
      qc.invalidateQueries({ queryKey: ["venues"] });
    },
    onError: (e) =>
      setError(e instanceof ApiError ? e.message : "Could not create venue."),
  });

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold tracking-tight">Venues</h1>

      <Card>
        <CardHeader>
          <CardTitle>Create a venue</CardTitle>
        </CardHeader>
        <CardBody>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              create.mutate();
            }}
            className="grid gap-4 sm:grid-cols-2"
          >
            <div>
              <Label>Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div>
              <Label>Address</Label>
              <Input
                value={address}
                onChange={(e) => setAddress(e.target.value)}
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
                Create venue
              </Button>
            </div>
          </form>
        </CardBody>
      </Card>

      <div className="space-y-3">
        <h2 className="text-base font-semibold">All venues</h2>
        {venuesQuery.isLoading ? (
          <LoadingState />
        ) : venues.length === 0 ? (
          <EmptyState
            title="No venues yet"
            description="Create one above to get started."
          />
        ) : (
          venues.map((venue) => <VenueRow key={venue.id} venue={venue} />)
        )}
      </div>
    </div>
  );
}

function VenueRow({ venue }: { venue: Venue }) {
  const qc = useQueryClient();
  const sectors = venue.sectors ?? [];
  const [sector, setSector] = useState({ name: "", rows: 5, cols: 8 });
  const [copied, setCopied] = useState(false);
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState({ name: venue.name, address: venue.address });
  const [error, setError] = useState<string | null>(null);

  const refetch = () => qc.invalidateQueries({ queryKey: ["venues"] });

  const add = useMutation({
    mutationFn: () =>
      addSector(venue.id, {
        name: sector.name,
        row_count: sector.rows,
        col_count: sector.cols,
      }),
    onSuccess: () => {
      setSector({ name: "", rows: 5, cols: 8 });
      refetch();
    },
  });

  const removeSector = useMutation({
    mutationFn: (sectorId: string) => deleteSector(venue.id, sectorId),
    onSuccess: refetch,
  });

  const save = useMutation({
    mutationFn: () => updateVenue(venue.id, draft),
    onSuccess: () => {
      setEditing(false);
      setError(null);
      refetch();
    },
    onError: (e) =>
      setError(e instanceof ApiError ? e.message : "Could not save venue."),
  });

  const remove = useMutation({
    mutationFn: () => deleteVenue(venue.id),
    onSuccess: refetch,
    onError: (e) =>
      setError(
        e instanceof ApiError
          ? e.message
          : "Could not delete venue.",
      ),
  });

  const totalSeats = sectors.reduce(
    (n, s) => n + s.row_count * s.col_count,
    0,
  );

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-start justify-between gap-3">
        {editing ? (
          <div className="flex flex-1 flex-wrap items-end gap-3">
            <div className="flex-1">
              <Label>Name</Label>
              <Input
                value={draft.name}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, name: e.target.value }))
                }
              />
            </div>
            <div className="flex-1">
              <Label>Address</Label>
              <Input
                value={draft.address}
                onChange={(e) =>
                  setDraft((d) => ({ ...d, address: e.target.value }))
                }
              />
            </div>
            <Button size="sm" onClick={() => save.mutate()} disabled={save.isPending}>
              Save
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => {
                setEditing(false);
                setDraft({ name: venue.name, address: venue.address });
                setError(null);
              }}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        ) : (
          <div>
            <p className="font-medium">{venue.name}</p>
            {venue.address && (
              <p className="text-sm text-muted">{venue.address}</p>
            )}
            <button
              className="mt-0.5 flex items-center gap-1.5 font-mono text-xs text-muted hover:text-ink"
              onClick={() => {
                navigator.clipboard.writeText(venue.id);
                setCopied(true);
                setTimeout(() => setCopied(false), 1500);
              }}
            >
              {venue.id}
              {copied ? (
                <Check className="h-3 w-3 text-success" />
              ) : (
                <Copy className="h-3 w-3" />
              )}
            </button>
          </div>
        )}

        {!editing && (
          <div className="flex shrink-0 items-center gap-1">
            <Button size="sm" variant="ghost" onClick={() => setEditing(true)}>
              <Pencil className="h-4 w-4" />
            </Button>
            <Button
              size="sm"
              variant="ghost"
              className="text-danger hover:bg-red-50 hover:text-danger"
              onClick={() => remove.mutate()}
              disabled={remove.isPending}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        )}
      </div>

      {error && (
        <p className="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-danger">
          {error}
        </p>
      )}

      <form
        onSubmit={(e) => {
          e.preventDefault();
          add.mutate();
        }}
        className="flex flex-wrap items-end gap-3"
      >
        <div className="flex-1">
          <Label>Sector name</Label>
          <Input
            value={sector.name}
            onChange={(e) => setSector((s) => ({ ...s, name: e.target.value }))}
            placeholder="Floor A"
            required
          />
        </div>
        <div className="w-20">
          <Label>Rows</Label>
          <Input
            type="number"
            min={1}
            value={sector.rows}
            onChange={(e) =>
              setSector((s) => ({ ...s, rows: Number(e.target.value) }))
            }
          />
        </div>
        <div className="w-20">
          <Label>Cols</Label>
          <Input
            type="number"
            min={1}
            value={sector.cols}
            onChange={(e) =>
              setSector((s) => ({ ...s, cols: Number(e.target.value) }))
            }
          />
        </div>
        <Button type="submit" variant="secondary" disabled={add.isPending}>
          Add sector
        </Button>
      </form>

      {add.isError && (
        <p className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-danger">
          Could not add the sector — try again.
        </p>
      )}

      <div className="mt-4 border-t border-line pt-3">
        <div className="mb-2 flex items-center justify-between">
          <p className="text-xs font-medium text-muted">
            Sectors ({sectors.length})
          </p>
          <p className="text-xs text-muted">{totalSeats} seats total</p>
        </div>
        {sectors.length === 0 ? (
          <p className="text-sm text-muted">
            No sectors yet — add one above before creating an event.
          </p>
        ) : (
          <ul className="space-y-1.5">
            {sectors.map((s) => (
              <li
                key={s.id}
                className="flex items-center justify-between rounded-lg bg-canvas px-3 py-2 text-sm transition-colors hover:bg-line/40"
              >
                <span className="flex items-center gap-2">
                  <Check className="h-3.5 w-3.5 text-success" />
                  <span className="font-medium">{s.name}</span>
                </span>
                <span className="flex items-center gap-3">
                  <span className="text-muted">
                    {s.row_count} × {s.col_count} ={" "}
                    <span className="font-medium text-ink">
                      {s.row_count * s.col_count}
                    </span>{" "}
                    seats
                  </span>
                  <button
                    className="text-muted hover:text-danger"
                    onClick={() => removeSector.mutate(s.id)}
                    disabled={removeSector.isPending}
                    aria-label="Delete sector"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </Card>
  );
}
