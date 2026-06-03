import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useFormatters } from "@/hooks/useFormatters";
import type { Sanatorium } from "@/types";
import { getSanatoriumCover, getSanatoriumFallback } from "@/lib/sanatoriumImages";
import { useI18n } from "@/i18n";
import { localizeSanatorium } from "@/lib/sanatoriumLocalization";
import MedicalProfileChip from "@/components/MedicalProfileChip";

export default function SanatoriumCard({ item }: { item: Sanatorium }) {
  const { t, lang } = useI18n();
  const { formatCurrency } = useFormatters();
  const localized = useMemo(() => localizeSanatorium(item, lang), [item, lang]);
  const fallbackCover = getSanatoriumFallback(item);
  const [cover, setCover] = useState("");

  useEffect(() => {
    setCover(getSanatoriumCover(item));
  }, [item]);
  return (
    <Link to={`/sanatoriums/${item.id}`} className="card group flex flex-col overflow-hidden transition hover:shadow-md">
      <div className="aspect-[16/10] w-full bg-slate-100">
        {cover ? (
          <img
            src={cover}
            alt={localized.name}
            loading="lazy"
            className="h-full w-full object-cover transition group-hover:scale-[1.02]"
            onError={() => {
              if (cover !== fallbackCover) setCover(fallbackCover);
            }}
          />
        ) : (
          <div className="flex h-full items-center justify-center text-sm text-slate-400">{t("card_no_photo")}</div>
        )}
      </div>
      <div className="flex flex-1 flex-col gap-2 p-4">
        <div className="flex items-start justify-between gap-3">
          <h3 className="font-semibold text-slate-900">{localized.name}</h3>
          <div className="shrink-0 text-right">
            <div className="text-sm font-semibold text-brand-700">{formatCurrency(item.price_per_night)}</div>
            <div className="text-[11px] text-slate-500">{t("card_per_night")}</div>
          </div>
        </div>
        <div className="text-xs text-slate-500">
          {localized.city} · {localized.address}
        </div>
        <p className="line-clamp-2 text-sm text-slate-600">{localized.description}</p>
        <div className="mt-auto flex flex-wrap gap-1 pt-2">
          {item.medical_profiles.slice(0, 4).map((slug) => (
            <MedicalProfileChip key={slug} slug={slug} />
          ))}
          {item.distance_to_sea_km <= 2 && (
            <span className="badge bg-emerald-50 text-emerald-700">
              {t("card_to_sea", { km: item.distance_to_sea_km })}
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
