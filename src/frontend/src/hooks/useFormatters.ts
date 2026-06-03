import { formatCurrency as fmtCurrency, formatDate as fmtDate } from "@/lib/utils";
import { useI18n } from "@/i18n";

/** Форматирование сумм и дат с учётом выбранного языка UI. */
export function useFormatters() {
  const { locale } = useI18n();
  return {
    formatCurrency: (amount: number, currency = "RUB") => fmtCurrency(amount, currency, locale),
    formatDate: (value: string | Date) => fmtDate(value, locale),
  };
}
