import type { User, UserRole } from "@/types";

function asString(value: unknown, fallback = ""): string {
  if (value == null) return fallback;
  return String(value);
}

/** Строка для отображения входа: email надёжнее full_name (кодировка БД). */
export function userLoginLabel(user: Pick<User, "email" | "full_name">): string {
  const email = user.email?.trim();
  if (email) return email;
  const name = user.full_name?.trim();
  if (name && !/^[\uFFFD?]+$/.test(name.replace(/\s/g, ""))) return name;
  return "";
}

/** Нормализует объект пользователя из API (deal/auth proxy). */
export function normalizeUser(raw: unknown): User {
  const u = (raw && typeof raw === "object" ? raw : {}) as Record<string, unknown>;
  const roleRaw = asString(u.role).toLowerCase();
  const role: UserRole = roleRaw === "admin" ? "admin" : "client";

  return {
    id: asString(u.id),
    email: asString(u.email),
    full_name: asString(u.full_name || u.fullName),
    role,
    created_at: asString(u.created_at || u.createdAt, new Date().toISOString()),
  };
}
