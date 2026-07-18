"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/lib/store/auth-store";

export function useSession() {
  const [hydrated, setHydrated] = useState(false);
  const user = useAuthStore((s) => s.user);
  const tokens = useAuthStore((s) => s.tokens);

  useEffect(() => setHydrated(true), []);

  return {
    hydrated,
    user: hydrated ? user : null,
    isAuthenticated: hydrated && !!tokens,
  };
}
