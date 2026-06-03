import { classNames } from "@/lib/utils";
import { useI18n } from "@/i18n";

const palette: Record<string, string> = {
  created: "bg-slate-100 text-slate-700",
  confirmed: "bg-emerald-100 text-emerald-700",
  completed: "bg-brand-100 text-brand-700",
  cancelled: "bg-red-100 text-red-700",
  unpaid: "bg-slate-100 text-slate-600",
  invoice_requested: "bg-amber-100 text-amber-700",
  invoice_issued: "bg-blue-100 text-blue-700",
  invoice_failed: "bg-red-100 text-red-700",
  paid: "bg-emerald-100 text-emerald-700",
  failed: "bg-red-100 text-red-700",
  refunded: "bg-slate-200 text-slate-700",
};

export default function StatusBadge({ status }: { status: string }) {
  const { t } = useI18n();
  return (
    <span className={classNames("badge", palette[status] ?? "bg-slate-100 text-slate-700")}>
      {t(`status_${status}`)}
    </span>
  );
}
