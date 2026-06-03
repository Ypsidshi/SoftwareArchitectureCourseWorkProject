import { dealApi } from "./client";
import type { CatalogFilters, PagedResponse, Sanatorium, SanatoriumDetail } from "@/types";

export async function listSanatoriums(filters: CatalogFilters): Promise<PagedResponse<Sanatorium>> {
  const params: Record<string, string | number> = {};
  for (const [key, value] of Object.entries(filters)) {
    if (value === undefined || value === null || value === "") continue;
    params[key] = value as string | number;
  }
  const res = await dealApi.get<PagedResponse<Sanatorium>>("/api/sanatoriums", { params });
  return res.data;
}

export async function getSanatorium(
  id: string,
  range?: { check_in: string; check_out: string },
): Promise<SanatoriumDetail> {
  const res = await dealApi.get<SanatoriumDetail>(`/api/sanatoriums/${id}`, { params: range });
  return res.data;
}
