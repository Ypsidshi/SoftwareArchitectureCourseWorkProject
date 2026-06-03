import { dealApi } from "./client";
import type { Booking, PagedResponse } from "@/types";

export interface BookingInput {
  sanatorium_id: string;
  check_in: string;
  check_out: string;
  guests: number;
}

export interface BookingUpdate {
  check_in: string;
  check_out: string;
  guests: number;
}

export async function createBooking(input: BookingInput): Promise<Booking> {
  const res = await dealApi.post<Booking>("/api/bookings", input);
  return res.data;
}

export async function listBookings(page = 1, pageSize = 10): Promise<PagedResponse<Booking>> {
  const res = await dealApi.get<PagedResponse<Booking>>("/api/bookings", {
    params: { page, page_size: pageSize },
  });
  return res.data;
}

export async function getBooking(id: string): Promise<Booking> {
  const res = await dealApi.get<Booking>(`/api/bookings/${id}`);
  return res.data;
}

export async function updateBooking(id: string, input: BookingUpdate): Promise<Booking> {
  const res = await dealApi.put<Booking>(`/api/bookings/${id}`, input);
  return res.data;
}

export async function cancelBooking(id: string): Promise<Booking> {
  const res = await dealApi.delete<Booking>(`/api/bookings/${id}`);
  return res.data;
}

export async function checkoutBooking(id: string): Promise<Booking> {
  const res = await dealApi.post<Booking>(`/api/bookings/${id}/checkout`);
  return res.data;
}

export interface PayBookingResult {
  booking: Booking;
  duplicate: boolean;
}

export async function payBooking(id: string, idempotencyKey: string): Promise<PayBookingResult> {
  const res = await dealApi.post<PayBookingResult>(`/api/bookings/${id}/pay`, null, {
    headers: { "Idempotency-Key": idempotencyKey },
  });
  return res.data;
}
