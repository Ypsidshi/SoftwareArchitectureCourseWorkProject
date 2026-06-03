import { dealApi } from "./client";
import type { Booking, PagedResponse, Sanatorium } from "@/types";

export interface AdminBookingsQuery {
  page?: number;
  page_size?: number;
  status?: string;
  payment_status?: string;
  city?: string;
  sanatorium_id?: string;
}

export interface SanatoriumInput {
  name: string;
  description: string;
  city: string;
  address: string;
  distance_to_sea_km: number;
  amenities: string[];
  image_urls: string[];
  price_per_night: number;
  total_places: number;
  latitude?: number;
  longitude?: number;
  medical_profiles?: string[];
}

export async function listAdminBookings(q: AdminBookingsQuery = {}): Promise<PagedResponse<Booking>> {
  const res = await dealApi.get<PagedResponse<Booking>>("/api/admin/bookings", { params: q });
  return res.data;
}

export async function adminCancelBooking(id: string): Promise<Booking> {
  const res = await dealApi.delete<Booking>(`/api/admin/bookings/${id}`);
  return res.data;
}

export async function adminCheckoutBooking(id: string): Promise<Booking> {
  const res = await dealApi.post<Booking>(`/api/admin/bookings/${id}/checkout`);
  return res.data;
}

export async function adminPayBooking(id: string, idempotencyKey: string): Promise<{ booking: Booking }> {
  const res = await dealApi.post<{ booking: Booking }>(`/api/admin/bookings/${id}/pay`, null, {
    headers: { "Idempotency-Key": idempotencyKey },
  });
  return res.data;
}

export async function listAdminSanatoriums(page = 1, pageSize = 20): Promise<PagedResponse<Sanatorium>> {
  const res = await dealApi.get<PagedResponse<Sanatorium>>("/api/admin/sanatoriums", {
    params: { page, page_size: pageSize },
  });
  return res.data;
}

export async function createAdminSanatorium(input: SanatoriumInput): Promise<Sanatorium> {
  const res = await dealApi.post<Sanatorium>("/api/admin/sanatoriums", input);
  return res.data;
}

export async function updateAdminSanatorium(id: string, input: SanatoriumInput): Promise<Sanatorium> {
  const res = await dealApi.put<Sanatorium>(`/api/admin/sanatoriums/${id}`, input);
  return res.data;
}

export async function deleteAdminSanatorium(id: string): Promise<void> {
  await dealApi.delete(`/api/admin/sanatoriums/${id}`);
}
