"use client";

import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Users } from "lucide-react";
import { joinQueue, queueStatus } from "@/lib/api/queue";
import { Card } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";

export default function QueuePage({
  params,
}: {
  params: Promise<{ eventId: string }>;
}) {
  const { eventId } = use(params);
  const router = useRouter();
  const [position, setPosition] = useState<number | null>(null);

  useEffect(() => {
    let active = true;
    let timer: ReturnType<typeof setTimeout>;

    const admit = (token: string) => {
      sessionStorage.setItem(`admission:${eventId}`, token);
      router.replace(`/events/${eventId}`);
    };

    const tick = async () => {
      try {
        const status = await queueStatus(eventId);
        if (!active) return;
        if (status.status === "admitted" && status.admission_token) {
          admit(status.admission_token);
          return;
        }
        setPosition(status.position ?? null);
      } catch {
        // keep polling
      }
      timer = setTimeout(tick, 2000);
    };

    joinQueue(eventId)
      .then((res) => {
        if (!active) return;
        if (res.status === "admitted" && res.admission_token) {
          admit(res.admission_token);
          return;
        }
        setPosition(res.position ?? null);
        timer = setTimeout(tick, 2000);
      })
      .catch(() => {
        timer = setTimeout(tick, 2000);
      });

    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [eventId, router]);

  return (
    <div className="mx-auto max-w-md py-12">
      <Card className="p-8 text-center">
        <span className="mx-auto grid h-14 w-14 place-items-center rounded-full bg-brand-soft text-brand">
          <Users className="h-7 w-7" />
        </span>
        <h1 className="mt-5 text-2xl font-bold tracking-tight">
          You're in the queue
        </h1>
        <p className="mt-2 text-muted">
          This event is in high demand. We'll let you in as soon as it's your
          turn — keep this tab open.
        </p>

        <div className="my-8">
          {position === null ? (
            <div className="flex items-center justify-center gap-2 text-muted">
              <Spinner />
              <span>Finding your place…</span>
            </div>
          ) : (
            <>
              <p className="text-5xl font-bold text-brand">
                {position.toLocaleString()}
              </p>
              <p className="mt-1 text-sm text-muted">people ahead of you</p>
            </>
          )}
        </div>
      </Card>
    </div>
  );
}
