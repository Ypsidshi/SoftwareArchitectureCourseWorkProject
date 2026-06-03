import { useI18n } from "@/i18n";

/** Подтверждение действия с текстом из словаря i18n. */
export function useConfirm() {
  const { t } = useI18n();
  return (messageKey: string) => window.confirm(t(messageKey));
}

