import { Link } from "react-router-dom";
import StatusBadge from "@/components/StatusBadge";
import { daysBetween } from "@/lib/utils";
import { useFormatters } from "@/hooks/useFormatters";
import { useI18n } from "@/i18n";
import type { Booking } from "@/types";
import type { ReactNode } from "react";

interface Props {
  booking: Booking;
  actions?: ReactNode;
  showClient?: boolean;
}

export default function BookingCard({ booking, actions, showClient }: Props) {
  const { t } = useI18n();
  const { formatCurrency, formatDate } = useFormatters();
  const b = booking;

  return (
    <div className="card p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-medium">{b.sanatorium_name || `#${b.sanatorium_id.slice(0, 8)}`}</span>
            <StatusBadge status={b.status} />
            <StatusBadge status={b.payment_status || "unpaid"} />
            <span className="text-xs text-slate-500">#{b.id.slice(0, 8)}</span>
          </div>
          <div className="mt-1 text-sm text-slate-600">
            {formatDate(b.check_in)} → {formatDate(b.check_out)} (
            {t("bookings_nights", { count: daysBetween(b.check_in, b.check_out) })})
          </div>
          {showClient && (
            <div className="mt-1 text-xs text-slate-500">
              {b.client_email || b.client_id} · {t("bookings_guests", { count: b.guests })}
            </div>
          )}
          {!showClient && (
            <div className="text-xs text-slate-500">{t("bookings_guests", { count: b.guests })}</div>
          )}
          {b.amount != null && (
            <div className="mt-1 font-medium text-brand-700">{formatCurrency(b.amount, b.currency || "RUB")}</div>
          )}
          {b.payment_error && <p className="mt-1 text-xs text-red-600">{b.payment_error}</p>}
          <Link to={`/sanatoriums/${b.sanatorium_id}`} className="mt-1 inline-block text-xs text-brand-700 hover:underline">
            {t("bookings_view_sanatorium")} →
          </Link>
        </div>
        {actions && <div className="flex flex-col gap-2">{actions}</div>}
      </div>
    </div>
  );
}
