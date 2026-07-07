import { apiFetch } from "./client";
import type { AdminUser, Role } from "@/lib/types";

export function listUsers() {
  return apiFetch<{ users: AdminUser[] }>("/admin/users");
}

export function setUserRole(id: string, role: Role) {
  return apiFetch<AdminUser>(`/admin/users/${id}/role`, {
    method: "POST",
    body: { role },
  });
}
