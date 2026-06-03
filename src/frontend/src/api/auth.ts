import { dealApi } from "./client";
import { normalizeUser } from "@/lib/user";
import type { LoginResponse, User } from "@/types";

export async function login(email: string, password: string): Promise<LoginResponse> {
  const res = await dealApi.post<{ access_token: string; user: unknown }>("/api/auth/login", { email, password });
  return {
    access_token: res.data.access_token,
    user: normalizeUser(res.data.user),
  };
}

export interface RegisterPayload {
  email: string;
  password: string;
  full_name: string;
}

export async function register(payload: RegisterPayload): Promise<User> {
  const res = await dealApi.post<unknown>("/api/auth/register", {
    email: payload.email,
    password: payload.password,
    full_name: payload.full_name,
    role: "client",
  });
  return normalizeUser(res.data);
}
