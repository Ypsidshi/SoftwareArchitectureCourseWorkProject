import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  adminCancelBooking,
  adminCheckoutBooking,
  adminPayBooking,
  listAdminBookings,
} from "@/api/admin";
import { useExtractError } from "@/hooks/useExtractError";
import { useConfirm } from "@/hooks/useConfirm";
import BookingCard from "@/components/BookingCard";
import Spinner from "@/components/Spinner";
import ErrorAlert from "@/components/ErrorAlert";
import Pagination from "@/components/Pagination";
import { useAdminBookingActions } from "@/hooks/useBookingPayment";
import { useI18n } from "@/i18n";

const STATUS_OPTS = ["", "confirmed", "cancelled"];
const PAY_OPTS = ["", "unpaid", "invoice_issued", "invoice_failed", "paid"];

export default function AdminBookingsTab() {
  const { t } = useI18n();
  const askConfirm = useConfirm();
  const toError = useExtractError();
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const [paymentStatus, setPaymentStatus] = useState("");
  const [city, setCity] = useState("");

  const queryKey = ["admin-bookings", page, status, paymentStatus, city];

  const query = useQuery({
    queryKey,
    queryFn: () =>
      listAdminBookings({
        page,
        page_size: 15,
        status: status || undefined,
        payment_status: paymentStatus || undefined,
        city: city || undefined,
      }),
  });

  const { error, cancel, checkout, pay, isBusy } = useAdminBookingActions({
    queryKey: ["admin-bookings"],
    checkout: adminCheckoutBooking,
    pay: adminPayBooking,
    cancel: adminCancelBooking,
  });

  return (
    <div>
      <p className="mb-4 text-sm text-slate-600">{t("admin_bookings_hint")}</p>

      <div className="card mb-4 grid gap-3 p-4 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label className="label">{t("admin_filter_status")}</label>
          <select className="input" value={status} onChange={(e) => { setPage(1); setStatus(e.target.value); }}>
            {STATUS_OPTS.map((v) => (
              <option key={v || "all"} value={v}>{v ? t(`status_${v}`) : t("admin_filter_all")}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="label">{t("admin_filter_payment")}</label>
          <select className="input" value={paymentStatus} onChange={(e) => { setPage(1); setPaymentStatus(e.target.value); }}>
            {PAY_OPTS.map((v) => (
              <option key={v || "all"} value={v}>{v ? t(`status_${v}`) : t("admin_filter_all")}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="label">{t("filters_city")}</label>
          <input className="input" value={city} onChange={(e) => { setPage(1); setCity(e.target.value); }} placeholder={t("filters_city_ph")} />
        </div>
        <div className="flex items-end">
          <button
            type="button"
            className="btn-secondary w-full"
            disabled={!status && !paymentStatus && !city}
            onClick={() => {
              setPage(1);
              setStatus("");
              setPaymentStatus("");
              setCity("");
            }}
          >
            {t("admin_filters_reset")}
          </button>
        </div>
      </div>

      {query.isLoading && <Spinner />}
      {query.isError && <ErrorAlert message={toError(query.error)} />}
      {error && <ErrorAlert message={error} />}

      {query.data?.items.length === 0 && (
        <div className="card p-8 text-center text-slate-500">{t("admin_bookings_empty")}</div>
      )}

      {query.data && query.data.items.length > 0 && (
        <div className="space-y-3">
          {query.data.items.map((b) => (
            <BookingCard
              key={b.id}
              booking={b}
              showClient
              actions={
                b.status !== "cancelled" ? (
                  <>
                    {b.payment_status !== "paid" && (
                      <>
                        {(b.payment_status === "unpaid" || b.payment_status === "invoice_failed") && (
                          <button type="button" className="btn-secondary text-sm" disabled={isBusy} onClick={() => checkout.mutate(b.id)}>
                            {t("admin_issue_invoice")}
                          </button>
                        )}
                        {b.payment_status === "invoice_issued" && (
                          <button type="button" className="btn-primary text-sm" disabled={isBusy} onClick={() => pay.mutate(b.id)}>
                            {t("admin_mark_paid")}
                          </button>
                        )}
                      </>
                    )}
                    <button
                      type="button"
                      className="btn-danger text-sm"
                      disabled={isBusy}
                      onClick={() => { if (askConfirm("bookings_cancel_confirm")) cancel.mutate(b.id); }}
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
    </div>
  );
}
