import { extractErrorMessage } from "@/api/client";
import { useI18n } from "@/i18n";

const ERROR_PATTERNS: { match: RegExp | string; key: string }[] = [
  { match: /sanatorium has active bookings/i, key: "err_san_has_bookings" },
  { match: /sanatorium not found/i, key: "err_san_not_found" },
  { match: /invalid json/i, key: "err_invalid_json" },
  { match: /forbidden/i, key: "err_forbidden" },
  { match: /unauthorized/i, key: "err_unauthorized" },
  { match: /conflict/i, key: "err_conflict" },
  { match: /Scan, not \d+/i, key: "err_server" },
  { match: /sql:/i, key: "err_server" },
];

export function useExtractError() {
  const { t } = useI18n();
  return (err: unknown) => {
    const raw = extractErrorMessage(err, t("err_generic"));
    for (const { match, key } of ERROR_PATTERNS) {
      if (typeof match === "string" ? raw.includes(match) : match.test(raw)) {
        return t(key);
      }
    }
    return raw;
  };
}
