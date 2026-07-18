"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Sidebar } from "@/components/admin/sidebar";
import { useSession } from "@/lib/hooks/use-session";
import { LoadingState } from "@/components/ui/spinner";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { hydrated, user } = useSession();
  const isStaff = user?.role === "organizer" || user?.role === "admin";

  useEffect(() => {
    if (hydrated && !isStaff) router.replace("/login?next=/admin");
  }, [hydrated, isStaff, router]);

  if (!hydrated || !isStaff) {
    return <LoadingState label="Checking access…" />;
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 overflow-x-hidden bg-canvas p-8">
        <div className="mx-auto max-w-4xl">{children}</div>
      </main>
    </div>
  );
}
