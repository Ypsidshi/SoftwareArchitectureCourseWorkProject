import type { Language } from "@/i18n";

/** Slugs stored in DB (deal.medical_profiles.name). */
export const MEDICAL_PROFILE_SLUGS = [
  "cardiology",
  "pulmonology",
  "musculoskeletal",
  "neurology",
  "gastroenterology",
  "endocrinology",
  "dermatology",
  "urology",
  "pediatrics",
  "balneology",
  "rehabilitation",
  "general_therapy",
] as const;

export type MedicalProfileSlug = (typeof MEDICAL_PROFILE_SLUGS)[number];

const ruLabel: Record<string, string> = {
  cardiology: "кардиология",
  pulmonology: "пульмонология",
  musculoskeletal: "опорно-двигательный аппарат",
  neurology: "неврология",
  gastroenterology: "гастроэнтерология",
  endocrinology: "эндокринология",
  dermatology: "дерматология",
  urology: "урология",
  pediatrics: "педиатрия",
  balneology: "бальнеология",
  rehabilitation: "реабилитация",
  general_therapy: "общая терапия",
};

const enLabel: Record<string, string> = {
  cardiology: "Cardiology",
  pulmonology: "Pulmonology",
  musculoskeletal: "Musculoskeletal",
  neurology: "Neurology",
  gastroenterology: "Gastroenterology",
  endocrinology: "Endocrinology",
  dermatology: "Dermatology",
  urology: "Urology",
  pediatrics: "Pediatrics",
  balneology: "Balneology",
  rehabilitation: "Rehabilitation",
  general_therapy: "General therapy",
};

const ruToSlug: Record<string, string> = Object.fromEntries(
  Object.entries(ruLabel).map(([slug, label]) => [label.toLowerCase(), slug]),
);

const enToSlug: Record<string, string> = Object.fromEntries(
  Object.entries(enLabel).map(([slug, label]) => [label.toLowerCase(), slug]),
);

export function normalizeProfileSlug(slug: string): string {
  return slug.trim().toLowerCase();
}

export function displayProfile(slug: string, lang: Language): string {
  const key = normalizeProfileSlug(slug);
  if (lang === "ru") return ruLabel[key] ?? key.replace(/_/g, " ");
  return enLabel[key] ?? key.replace(/_/g, " ");
}

export function profileSlugFromInput(raw: string, lang: Language): string {
  const trimmed = raw.trim().toLowerCase();
  if (!trimmed) return "";
  if (lang === "ru") {
    const fromRu = ruToSlug[trimmed];
    if (fromRu) return fromRu;
  } else {
    const fromEn = enToSlug[trimmed];
    if (fromEn) return fromEn;
  }
  if ((MEDICAL_PROFILE_SLUGS as readonly string[]).includes(trimmed)) return trimmed;
  return trimmed;
}

export function profileOptions(lang: Language): { slug: string; label: string }[] {
  return MEDICAL_PROFILE_SLUGS.map((slug) => ({ slug, label: displayProfile(slug, lang) }));
}

export function formatProfilesList(slugs: string[], lang: Language): string {
  return slugs.map((s) => displayProfile(s, lang)).join(", ");
}

export function profilesInputFromSlugs(slugs: string[], lang: Language): string {
  return slugs.map((s) => displayProfile(s, lang)).join(", ");
}

export function profileTokenToApi(token: string, lang: Language = "ru"): string {
  return profileSlugFromInput(token, lang);
}

export function profileApiToLabel(code: string, lang: Language): string {
  return displayProfile(code.toLowerCase(), lang);
}
