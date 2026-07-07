"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listUsers, setUserRole } from "@/lib/api/admin";
import { Card } from "@/components/ui/card";
import { Select } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { LoadingState, EmptyState } from "@/components/ui/spinner";
import type { AdminUser, Role } from "@/lib/types";

const roles: Role[] = ["buyer", "organizer", "admin"];

export default function AdminUsersPage() {
  const qc = useQueryClient();
  const { data, isLoading, isError } = useQuery({
    queryKey: ["admin-users"],
    queryFn: listUsers,
  });

  const changeRole = useMutation({
    mutationFn: ({ id, role }: { id: string; role: Role }) =>
      setUserRole(id, role),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin-users"] }),
  });

  if (isLoading) return <LoadingState label="Loading users…" />;
  if (isError)
    return (
      <EmptyState
        title="Admin access required"
        description="Only admins can manage the user directory."
      />
    );

  const users = data?.users ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Users</h1>
        <p className="mt-1 text-sm text-muted">
          Manage roles across the platform.
        </p>
      </div>

      <Card className="divide-y divide-line">
        {users.length === 0 ? (
          <EmptyState title="No users found" />
        ) : (
          users.map((user) => (
            <UserRow
              key={user.id}
              user={user}
              onChange={(role) => changeRole.mutate({ id: user.id, role })}
              pending={changeRole.isPending}
            />
          ))
        )}
      </Card>
    </div>
  );
}

function UserRow({
  user,
  onChange,
  pending,
}: {
  user: AdminUser;
  onChange: (role: Role) => void;
  pending: boolean;
}) {
  return (
    <div className="flex items-center justify-between p-4">
      <div className="flex items-center gap-3">
        <span className="grid h-9 w-9 place-items-center rounded-full bg-brand-soft text-sm font-semibold text-brand">
          {user.email.slice(0, 1).toUpperCase()}
        </span>
        <div>
          <p className="text-sm font-medium">{user.email || "—"}</p>
          <p className="font-mono text-xs text-muted">{user.id.slice(0, 8)}</p>
        </div>
      </div>
      <div className="flex items-center gap-3">
        <Badge
          tone={
            user.role === "admin"
              ? "brand"
              : user.role === "organizer"
                ? "warning"
                : "neutral"
          }
        >
          {user.role}
        </Badge>
        <Select
          className="w-36"
          value={user.role}
          disabled={pending}
          onChange={(e) => onChange(e.target.value as Role)}
        >
          {roles.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </Select>
      </div>
    </div>
  );
}
