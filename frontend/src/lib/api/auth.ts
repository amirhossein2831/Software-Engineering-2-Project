import { apiFetch } from "./client";
import type { Tokens } from "@/lib/types";

export function register(email: string, password: string) {
  return apiFetch<{ id: string; email: string; role: string }>(
    "/auth/register",
    { method: "POST", body: { email, password }, auth: false },
  );
}

export function login(email: string, password: string) {
  return apiFetch<Tokens>("/auth/login", {
    method: "POST",
    body: { email, password },
    auth: false,
  });
}

export function logout(refreshToken: string) {
  return apiFetch<void>("/auth/logout", {
    method: "POST",
    body: { refresh_token: refreshToken },
    auth: false,
  });
}
