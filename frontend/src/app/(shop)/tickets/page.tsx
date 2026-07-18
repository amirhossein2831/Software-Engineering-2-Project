"use client";

import { useQuery } from "@tanstack/react-query";
import { listOrders, listTickets } from "@/lib/api/checkout";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { QrHashArt } from "@/components/shop/qr";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import type { Ticket } from "@/lib/types";

export default function TicketsPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["tickets"],
    queryFn: async () => {
      const { orders } = await listOrders();
      const paid = orders.filter((o) => o.status === "paid");
      const results = await Promise.all(
        paid.map((o) => listTickets(o.id).then((r) => r.tickets)),
      );
      return results.flat();
    },
  });

  if (isLoading) return <LoadingState label="Loading tickets…" />;
  if (isError) return <EmptyState title="Sign in to view your tickets" />;

  const tickets = data ?? [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">My tickets</h1>
      {tickets.length === 0 ? (
        <EmptyState
          title="No tickets yet"
          description="Book an event and your QR tickets will show up here."
        />
      ) : (
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {tickets.map((ticket) => (
            <TicketCard key={ticket.id} ticket={ticket} />
          ))}
        </div>
      )}
    </div>
  );
}

function TicketCard({ ticket }: { ticket: Ticket }) {
  return (
    <Card className="overflow-hidden">
      <div className="border-b border-dashed border-line bg-canvas p-5">
        <QrHashArt hash={ticket.qr_hash} className="mx-auto max-w-[180px]" />
      </div>
      <div className="space-y-2 p-5">
        <div className="flex items-center justify-between">
          <Badge tone="success">Valid</Badge>
          <span className="font-mono text-xs text-muted">
            {ticket.id.slice(0, 8)}
          </span>
        </div>
        <p className="text-sm">
          <span className="text-muted">Seat</span>{" "}
          <span className="font-mono">{ticket.seat_id.slice(0, 8)}</span>
        </p>
        <p className="break-all font-mono text-[10px] leading-tight text-muted">
          {ticket.qr_hash}
        </p>
      </div>
    </Card>
  );
}
