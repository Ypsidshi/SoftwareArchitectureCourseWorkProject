export type UserRole = "admin" | "client";

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: UserRole;
  created_at: string;
}

export interface LoginResponse {
  access_token: string;
  user: User;
}

export interface Sanatorium {
  id: string;
  name: string;
  description: string;
  city: string;
  address: string;
  distance_to_sea_km: number;
  amenities: string[];
  image_urls: string[];
  price_per_night: number;
  total_places: number;
  medical_profiles: string[];
  latitude?: number;
  longitude?: number;
  created_at: string;
  updated_at: string;
}

export interface SanatoriumDetail {
  sanatorium: Sanatorium;
  available: boolean;
}

export interface PagedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export type BookingStatus = "created" | "confirmed" | "cancelled";

export type BookingPaymentStatus = "unpaid" | "invoice_issued" | "invoice_failed" | "paid";

export interface Booking {
  id: string;
  client_id: string;
  client_email?: string;
  sanatorium_id: string;
  sanatorium_name?: string;
  check_in: string;
  check_out: string;
  guests: number;
  status: BookingStatus;
  amount?: number;
  currency?: string;
  payment_status?: BookingPaymentStatus;
  payment_error?: string;
  invoice_id?: string | null;
  created_at: string;
  updated_at: string;
  cancelled_at?: string;
}

export type ContractStatus = "created" | "confirmed" | "cancelled" | "completed";
export type PaymentStatus =
  | "invoice_requested"
  | "invoice_issued"
  | "invoice_failed"
  | "paid";

export interface Contract {
  id: string;
  resident_id: string;
  room_id: string;
  manager_id: string;
  start_date: string;
  end_date: string;
  amount: number;
  currency: string;
  status: ContractStatus;
  payment_status: PaymentStatus;
  payment_error: string;
  invoice_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface ApiError {
  message: string;
  status?: number;
}

export interface CatalogFilters {
  page?: number;
  page_size?: number;
  city?: string;
  profiles?: string;
  max_distance_to_sea?: number;
  price_min?: number;
  price_max?: number;
  check_in?: string;
  check_out?: string;
  sort?: "price_asc" | "price_desc" | "distance_asc" | "distance_desc";
}
