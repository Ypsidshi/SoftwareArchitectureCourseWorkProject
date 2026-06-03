import { useState } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { cancelBooking, checkoutBooking, listBookings, payBooking } from "@/api/bookings";
import BookingCard from "@/components/BookingCard";
import Spinner from "@/components/Spinner";
import ErrorAlert from "@/components/ErrorAlert";
import Pagination from "@/components/Pagination";
import { useClientBookingPay } from "@/hooks/useBookingPayment";
import { useConfirm } from "@/hooks/useConfirm";
import { useExtractError } from "@/hooks/useExtractError";
import { useI18n } from "@/i18n";
import type { Booking } from "@/types";

function payButtonLabel(booking: Booking, t: (key: string) => string): string {
  switch (booking.payment_status) {
    case "invoice_issued":
      return t("bookings_pay_complete");
    case "invoice_failed":
      return t("bookings_pay_retry");
    default:
      return t("bookings_pay");
  }
}

export default function MyBookingsPage() {
  const { t } = useI18n();
  const askConfirm = useConfirm();
  const toError = useExtractError();
  const [page, setPage] = useState(1);
  const [cancelError, setCancelError] = useState<string | null>(null);
  const qc = useQueryClient();

  const query = useQuery({
    queryKey: ["bookings", page],
    queryFn: () => listBookings(page, 10),
  });

  const { pay, error: payError } = useClientBookingPay({
    queryKey: ["bookings"],
    checkout: checkoutBooking,
    pay: payBooking,
  });

  const cancel = useMutation({
    mutationFn: cancelBooking,
    onSuccess: () => {
      setCancelError(null);
      qc.invalidateQueries({ queryKey: ["bookings"] });
    },
    onError: (err) => setCancelError(toError(err)),
  });

  const actionError = payError || cancelError;

  return (
    <div>
      <h1 className="mb-4 text-2xl font-semibold">{t("bookings_title")}</h1>

      {query.isLoading && <Spinner />}
      {query.isError && <ErrorAlert message={toError(query.error)} />}

      {query.data?.items.length === 0 && (
        <div className="card p-8 text-center text-slate-500">
          {t("bookings_empty")}{" "}
          <Link to="/" className="text-brand-700 hover:underline">
            {t("bookings_to_catalog")} →
          </Link>
        </div>
      )}

      {query.data && query.data.items.length > 0 && (
        <div className="space-y-3">
          {query.data.items.map((b) => (
            <BookingCard
              key={b.id}
              booking={b}
              actions={
                b.status !== "cancelled" ? (
                  <>
                    {b.payment_status !== "paid" && (
                      <button
                        type="button"
                        className="btn-primary text-sm"
                        disabled={pay.isPending || cancel.isPending}
                        onClick={() => {
                          if (askConfirm("bookings_pay_confirm")) pay.mutate(b);
                        }}
                      >
                        {pay.isPending ? t("bookings_paying") : payButtonLabel(b, t)}
                      </button>
                    )}
                    <button
                      type="button"
                      className="btn-danger text-sm"
                      disabled={cancel.isPending || pay.isPending}
                      onClick={() => {
                        if (askConfirm("bookings_cancel_confirm")) cancel.mutate(b.id);
                      }}
                    >
                      {t("bookings_cancel")}
                    </button>
                  </>
                ) : undefined
              }
            />
          ))}
          <Pagination page={query.data.page} totalPages={query.data.total_pages} onChange={setPage} />
        </div>
      )}

      {actionError && <ErrorAlert message={actionError} />}
    </div>
  );
}
