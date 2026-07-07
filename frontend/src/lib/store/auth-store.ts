"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { SessionUser, Tokens, Role } from "@/lib/types";

interface JwtPayload {
  uid: string;
  role: Role;
  sub: string;
  exp: number;
}

function decodeUser(accessToken: string): SessionUser | null {
  try {
    const [, payload] = accessToken.split(".");
    const json = JSON.parse(
      atob(payload.replace(/-/g, "+").replace(/_/g, "/")),
    ) as JwtPayload;
    return { id: json.uid, email: "", role: json.role };
  } catch {
    return null;
  }
}

interface AuthState {
  tokens: Tokens | null;
  user: SessionUser | null;
  setSession: (tokens: Tokens, email: string) => void;
  clear: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      tokens: null,
      user: null,
      setSession: (tokens, email) => {
        const user = decodeUser(tokens.access_token);
        set({ tokens, user: user ? { ...user, email } : null });
      },
      clear: () => set({ tokens: null, user: null }),
    }),
    { name: "ticketing-auth" },
  ),
);
