import axios, { AxiosError, type AxiosInstance } from "axios";
import { useAuthStore } from "@/store/auth";

function buildClient(baseURL: string): AxiosInstance {
  const instance = axios.create({ baseURL, headers: { "Content-Type": "application/json" } });

  instance.interceptors.request.use((config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers = config.headers ?? {};
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (response) => response,
    (error: AxiosError<{ message?: string; error?: string }>) => {
      if (error.response?.status === 401 && useAuthStore.getState().token) {
        useAuthStore.getState().logout();
      }
      return Promise.reject(error);
    },
  );

  return instance;
}

const dealBase = import.meta.env.VITE_DEAL_URL || "";

export const dealApi = buildClient(dealBase);

export function extractErrorMessage(err: unknown, fallback = "Что-то пошло не так"): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { message?: string; error?: string } | undefined;
    return data?.message || data?.error || err.message || fallback;
  }
  if (err instanceof Error) return err.message;
  return fallback;
}
