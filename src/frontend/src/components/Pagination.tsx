import { classNames } from "@/lib/utils";
import { useI18n } from "@/i18n";

interface Props {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
}

export default function Pagination({ page, totalPages, onChange }: Props) {
  const { t } = useI18n();
  if (totalPages <= 1) return null;
  const prev = () => onChange(Math.max(page - 1, 1));
  const next = () => onChange(Math.min(page + 1, totalPages));

  return (
    <div className="flex items-center justify-center gap-2 py-4">
      <button
        type="button"
        className={classNames("btn-secondary", page <= 1 && "pointer-events-none opacity-50")}
        onClick={prev}
        disabled={page <= 1}
        aria-label={t("pagination_prev")}
      >
        ← {t("pagination_prev")}
      </button>
      <div className="text-sm text-slate-600">
        {t("pagination_page", { page, total: totalPages })}
      </div>
      <button
        type="button"
        className={classNames("btn-secondary", page >= totalPages && "pointer-events-none opacity-50")}
        onClick={next}
        disabled={page >= totalPages}
        aria-label={t("pagination_next")}
      >
        {t("pagination_next")} →
      </button>
    </div>
  );
}
