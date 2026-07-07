"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Loader2, Plus, Copy, Check } from "lucide-react";
import { createVenue, addSector } from "@/lib/api/catalog";
import { ApiError } from "@/lib/api/client";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input, Label } from "@/components/ui/input";
import type { Venue } from "@/lib/types";

export default function AdminVenuesPage() {
  const [venues, setVenues] = useState<Venue[]>([]);
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [error, setError] = useState<string | null>(null);

  const create = useMutation({
    mutationFn: () => createVenue({ name, address }),
    onSuccess: (venue) => {
      setVenues((v) => [venue, ...v]);
      setName("");
      setAddress("");
      setError(null);
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

      {venues.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-base font-semibold">
            Created this session — add sectors
          </h2>
          {venues.map((venue) => (
            <VenueRow key={venue.id} venue={venue} />
          ))}
        </div>
      )}
    </div>
  );
}

function VenueRow({ venue }: { venue: Venue }) {
  const [sector, setSector] = useState({ name: "", rows: 5, cols: 8 });
  const [copied, setCopied] = useState(false);

  const add = useMutation({
    mutationFn: () =>
      addSector(venue.id, {
        name: sector.name,
        row_count: sector.rows,
        col_count: sector.cols,
      }),
    onSuccess: () => setSector({ name: "", rows: 5, cols: 8 }),
  });

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-center justify-between">
        <div>
          <p className="font-medium">{venue.name}</p>
          <button
            className="flex items-center gap-1.5 font-mono text-xs text-muted hover:text-ink"
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
      </div>
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
    </Card>
  );
}
