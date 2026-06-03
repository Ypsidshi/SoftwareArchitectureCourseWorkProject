import type { CatalogFilters } from "@/types";
import type { Language } from "@/i18n";
import { profileApiToLabel, profileTokenToApi } from "@/lib/medicalProfiles";

const cityRuToApi: Record<string, string> = {
  сочи: "Sochi",
  кисловодск: "Kislovodsk",
  светлогорск: "Svetlogorsk",
  геленджик: "Gelendzhik",
  белокуриха: "Belokurikha",
  анапа: "Anapa",
  ессентуки: "Yessentuki",
  ялта: "Yalta",
};

const cityApiToRu: Record<string, string> = {
  sochi: "Сочи",
  kislovodsk: "Кисловодск",
  svetlogorsk: "Светлогорск",
  gelendzhik: "Геленджик",
  belokurikha: "Белокуриха",
  anapa: "Анапа",
  yessentuki: "Ессентуки",
  yalta: "Ялта",
};

function splitCsv(input?: string): string[] {
  if (!input) return [];
  return input
    .split(",")
    .map((v) => v.trim())
    .filter(Boolean);
}

export { profileTokenToApi, profileApiToLabel };

export function splitProfiles(input?: string): string[] {
  return splitCsv(input);
}

export function normalizeFiltersForApi(filters: CatalogFilters, lang: Language = "ru"): CatalogFilters {
  const cityRaw = (filters.city ?? "").trim();
  const cityKey = cityRaw.toLowerCase();
  const city = cityRaw ? cityRuToApi[cityKey] ?? cityRaw : undefined;

  const profiles = splitCsv(filters.profiles)
    .map((p) => profileTokenToApi(p, lang))
    .join(",");

  return {
    ...filters,
    city,
    profiles: profiles || undefined,
  };
}

export function localizeFiltersForUi(filters: CatalogFilters, lang: Language = "ru"): CatalogFilters {
  const cityRaw = (filters.city ?? "").trim();
  const city = cityRaw ? cityApiToRu[cityRaw.toLowerCase()] ?? cityRaw : undefined;
  const profiles = splitCsv(filters.profiles)
    .map((p) => profileApiToLabel(p, lang))
    .join(", ");

  return {
    ...filters,
    city,
    profiles: profiles || undefined,
  };
}
