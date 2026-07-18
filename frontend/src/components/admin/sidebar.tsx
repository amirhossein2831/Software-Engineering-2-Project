"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  LayoutDashboard,
  CalendarDays,
  Building2,
  Users,
  ArrowLeft,
  LogOut,
} from "lucide-react";
import { useAuthStore } from "@/lib/store/auth-store";
import { useSession } from "@/lib/hooks/use-session";
import { cn } from "@/lib/utils";

const links = [
  { href: "/admin", label: "Dashboard", icon: LayoutDashboard, exact: true },
  { href: "/admin/events", label: "Events", icon: CalendarDays },
  { href: "/admin/venues", label: "Venues", icon: Building2 },
  { href: "/admin/users", label: "Users", icon: Users, adminOnly: true },
];

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user } = useSession();
  const clear = useAuthStore((s) => s.clear);

  return (
    <aside className="flex w-60 shrink-0 flex-col border-r border-line bg-surface p-4">
      <div className="px-2 py-3">
        <p className="text-sm font-semibold">Studio</p>
        <p className="text-xs capitalize text-muted">{user?.role} workspace</p>
      </div>

      <nav className="mt-2 flex-1 space-y-1">
        {links
          .filter((l) => !l.adminOnly || user?.role === "admin")
          .map((link) => {
            const active = link.exact
              ? pathname === link.href
              : pathname.startsWith(link.href);
            return (
              <Link
                key={link.href}
                href={link.href}
                className={cn(
                  "flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  active
                    ? "bg-brand-soft text-brand"
                    : "text-muted hover:bg-canvas hover:text-ink",
                )}
              >
                <link.icon className="h-4 w-4" />
                {link.label}
              </Link>
            );
          })}
      </nav>

      <div className="space-y-1 border-t border-line pt-3">
        <Link
          href="/events"
          className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-muted hover:bg-canvas hover:text-ink"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to shop
        </Link>
        <button
          onClick={() => {
            clear();
            router.push("/events");
          }}
          className="flex w-full items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-muted hover:bg-canvas hover:text-ink"
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </button>
      </div>
    </aside>
  );
}
