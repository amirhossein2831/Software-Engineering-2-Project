"use client";

import { use, useState } from "react";
import Link from "next/link";
import { CheckCircle2, XCircle, Loader2, CreditCard } from "lucide-react";
import { checkout } from "@/lib/api/checkout";
import { ApiError } from "@/lib/api/client";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatMoney } from "@/lib/utils";
import type { Order } from "@/lib/types";

export default function CheckoutPage({
  params,
}: {
  params: Promise<{ holdId: string }>;
}) {
  const { holdId } = use(params);
  const [idempotencyKey] = useState(() => crypto.randomUUID());
  const [order, setOrder] = useState<Order | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function pay() {
    setSubmitting(true);
    setError(null);
    try {
      const admission =
        typeof window !== "undefined"
          ? (sessionStorage.getItem(`admission:${holdId}`) ?? undefined)
          : undefined;
      const result = await checkout(holdId, idempotencyKey, admission);
      setOrder(result);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Checkout failed.");
    } finally {
      setSubmitting(false);
    }
  }

  if (order) return <OrderResult order={order} />;

  return (
    <div className="mx-auto max-w-lg">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">Checkout</h1>
      <Card className="p-6">
        <div className="flex items-center gap-3">
          <span className="grid h-10 w-10 place-items-center rounded-lg bg-brand-soft text-brand">
            <CreditCard className="h-5 w-5" />
          </span>
          <div>
            <p className="font-medium">Confirm and pay</p>
            <p className="text-sm text-muted">
              Your seats are held. Complete payment to secure your tickets.
            </p>
          </div>
        </div>

        {error && (
          <p className="mt-5 rounded-lg bg-red-50 px-3 py-2 text-sm text-danger">
            {error}
          </p>
        )}

        <Button
          className="mt-6 w-full"
          size="lg"
          onClick={pay}
          disabled={submitting}
        >
          {submitting && <Loader2 className="h-4 w-4 animate-spin" />}
          Pay now
        </Button>
        <p className="mt-3 text-center text-xs text-muted">
          Payments run through a mock gateway in this environment.
        </p>
      </Card>
    </div>
  );
}

function OrderResult({ order }: { order: Order }) {
  const paid = order.status === "paid";
  return (
    <div className="mx-auto max-w-lg">
      <Card className="p-8 text-center">
        {paid ? (
          <CheckCircle2 className="mx-auto h-14 w-14 text-success" />
        ) : (
          <XCircle className="mx-auto h-14 w-14 text-danger" />
        )}
        <h1 className="mt-4 text-2xl font-bold tracking-tight">
          {paid ? "You're going!" : "Payment not completed"}
        </h1>
        <p className="mt-2 text-muted">
          {paid
            ? "Your payment succeeded and your seats are booked."
            : "Your seats were released. You can try booking again."}
        </p>

        <div className="my-6 flex items-center justify-between rounded-xl border border-line bg-canvas px-4 py-3 text-left">
          <div>
            <p className="text-xs uppercase tracking-wide text-muted">Order</p>
            <p className="font-mono text-sm">{order.id.slice(0, 8)}</p>
          </div>
          <Badge tone={paid ? "success" : "danger"}>{order.status}</Badge>
          <p className="text-lg font-bold">
            {formatMoney(order.amount, order.currency)}
          </p>
        </div>

        <div className="flex gap-3">
          {paid ? (
            <Link href="/tickets" className="flex-1">
              <Button className="w-full">View my tickets</Button>
            </Link>
          ) : (
            <Link href="/events" className="flex-1">
              <Button className="w-full" variant="secondary">
                Back to events
              </Button>
            </Link>
          )}
        </div>
      </Card>
    </div>
  );
}
