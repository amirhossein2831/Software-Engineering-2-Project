"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { listOrders } from "@/lib/api/checkout";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import { formatMoney, formatDate } from "@/lib/utils";
import type { OrderStatus } from "@/lib/types";

const tone: Record<OrderStatus, "success" | "danger" | "warning" | "neutral"> = {
  paid: "success",
  failed: "danger",
  compensated: "warning",
  pending: "neutral",
};

export default function OrdersPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ["orders"],
    queryFn: listOrders,
  });

  if (isLoading) return <LoadingState label="Loading orders…" />;
  if (isError)
    return <EmptyState title="Sign in to view your orders" />;

  const orders = data?.orders ?? [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Your orders</h1>
      {orders.length === 0 ? (
        <EmptyState
          title="No orders yet"
          description="When you book seats, your orders will appear here."
        />
      ) : (
        <div className="space-y-3">
          {orders.map((order) => (
            <Link key={order.id} href={`/tickets?order=${order.id}`}>
              <Card className="flex items-center justify-between p-4 transition-shadow hover:shadow-md">
                <div>
                  <p className="font-mono text-sm">{order.id.slice(0, 8)}</p>
                  <p className="text-sm text-muted">
                    {order.seat_ids.length} seat
                    {order.seat_ids.length > 1 ? "s" : ""} ·{" "}
                    {formatDate(order.created_at)}
                  </p>
                </div>
                <div className="flex items-center gap-4">
                  <Badge tone={tone[order.status]}>{order.status}</Badge>
                  <span className="font-semibold">
                    {formatMoney(order.amount, order.currency)}
                  </span>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
