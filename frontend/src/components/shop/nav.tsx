"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { Ticket, LogOut, LayoutDashboard } from "lucide-react";
import { useAuthStore } from "@/lib/store/auth-store";
import { useSession } from "@/lib/hooks/use-session";
import { Button } from "@/components/ui/button";

export function Nav() {
  const { user, isAuthenticated } = useSession();
  const clear = useAuthStore((s) => s.clear);
  const router = useRouter();
  const isStaff = user?.role === "organizer" || user?.role === "admin";

  return (
    <header className="sticky top-0 z-40 border-b border-line bg-surface/80 backdrop-blur">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4">
        <Link href="/events" className="flex items-center gap-2 font-semibold">
          <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand text-brand-fg">
            <Ticket className="h-4 w-4" />
          </span>
          <span>LiveTickets</span>
        </Link>

        <nav className="flex items-center gap-1.5">
          <Link href="/events">
            <Button variant="ghost" size="sm">
              Events
            </Button>
          </Link>
          {isAuthenticated ? (
            <>
              <Link href="/tickets">
                <Button variant="ghost" size="sm">
                  My tickets
                </Button>
              </Link>
              {isStaff && (
                <Link href="/admin">
                  <Button variant="ghost" size="sm">
                    <LayoutDashboard className="h-4 w-4" />
                    Studio
                  </Button>
                </Link>
              )}
              <Button
                variant="secondary"
                size="sm"
                onClick={() => {
                  clear();
                  router.push("/events");
                }}
              >
                <LogOut className="h-4 w-4" />
                Sign out
              </Button>
            </>
          ) : (
            <>
              <Link href="/login">
                <Button variant="ghost" size="sm">
                  Sign in
                </Button>
              </Link>
              <Link href="/register">
                <Button size="sm">Get started</Button>
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
