import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useExtractError } from "@/hooks/useExtractError";
import { generateIdempotencyKey } from "@/lib/utils";
import { useState } from "react";
import type { Booking } from "@/types";

type CheckoutFn = (id: string) => Promise<unknown>;
type PayFn = (id: string, key: string) => Promise<unknown>;
type CancelFn = (id: string) => Promise<unknown>;

interface ClientPayOptions {
  queryKey: string[];
  checkout: CheckoutFn;
  pay: PayFn;
}

/** Checkout + pay in one step (client flow). */
export function useClientBookingPay({ queryKey, checkout, pay }: ClientPayOptions) {
  const qc = useQueryClient();
  const toError = useExtractError();
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async (booking: Booking) => {
      const status = booking.payment_status || "unpaid";
      if (status !== "invoice_issued") {
        await checkout(booking.id);
      }
      return pay(booking.id, generateIdempotencyKey());
    },
    onSuccess: () => {
      setError(null);
      qc.invalidateQueries({ queryKey });
    },
    onError: (e) => setError(toError(e)),
  });

  return { pay: mutation, error, clearError: () => setError(null) };
}

interface AdminPayOptions {
  queryKey: string[];
  checkout: CheckoutFn;
  pay: PayFn;
  cancel: CancelFn;
}

/** Separate checkout / pay / cancel (admin flow). */
export function useAdminBookingActions({ queryKey, checkout, pay, cancel }: AdminPayOptions) {
  const qc = useQueryClient();
  const toError = useExtractError();
  const [error, setError] = useState<string | null>(null);

  const invalidate = () => {
    setError(null);
    qc.invalidateQueries({ queryKey });
  };

  const onError = (e: unknown) => setError(toError(e));

  const cancelMut = useMutation({ mutationFn: cancel, onSuccess: invalidate, onError });
  const checkoutMut = useMutation({ mutationFn: checkout, onSuccess: invalidate, onError });
  const payMut = useMutation({
    mutationFn: (id: string) => pay(id, generateIdempotencyKey()),
    onSuccess: invalidate,
    onError,
  });

  return {
    error,
    cancel: cancelMut,
    checkout: checkoutMut,
    pay: payMut,
    isBusy: cancelMut.isPending || checkoutMut.isPending || payMut.isPending,
  };
}
