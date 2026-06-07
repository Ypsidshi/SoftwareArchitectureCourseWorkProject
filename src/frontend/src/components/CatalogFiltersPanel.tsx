import { useState } from "react";
import type { CatalogFilters } from "@/types";
import { plusDaysIsoDate, todayIsoDate } from "@/lib/utils";
import { useI18n } from "@/i18n";
import { profileTokenToApi, splitProfiles } from "@/lib/catalogFilters";
import { displayProfile, MEDICAL_PROFILE_SLUGS } from "@/lib/medicalProfiles";
import { classNames } from "@/lib/utils";

interface Props {
  initial: CatalogFilters;
  onApply: (filters: CatalogFilters) => void;
}

export default function CatalogFiltersPanel({ initial, onApply }: Props) {
  const { t, lang } = useI18n();
  const [form, setForm] = useState<CatalogFilters>(initial);
  const [dateError, setDateError] = useState<string | null>(null);
  const minCheckIn = todayIsoDate();
  const minCheckOut = form.check_in ? plusDaysIsoDate(form.check_in, 1) : todayIsoDate();
  const selectedProfiles = new Set(splitProfiles(form.profiles).map((p) => profileTokenToApi(p)));
  const profileOptions = MEDICAL_PROFILE_SLUGS;

  const update = <K extends keyof CatalogFilters>(key: K, value: CatalogFilters[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (form.check_in && form.check_out && form.check_out <= form.check_in) {
      setDateError(t("filters_date_error"));
      return;
    }
    setDateError(null);
    onApply({ ...form, page: 1 });
  };

  const handleReset = () => {
    const empty: CatalogFilters = { page: 1, page_size: initial.page_size ?? 12 };
    setForm(empty);
    setDateError(null);
    onApply(empty);
  };

  const toggleProfile = (profileCode: string) => {
    const next = new Set(selectedProfiles);
    if (next.has(profileCode)) next.delete(profileCode);
    else next.add(profileCode);
    update("profiles", Array.from(next).join(",") || undefined);
  };

  const applyDatePreset = (nights: number) => {
    const start = todayIsoDate();
    update("check_in", start);
    update("check_out", plusDaysIsoDate(start, nights));
    setDateError(null);
  };

  return (
    <form onSubmit={handleSubmit} className="card space-y-4 p-4">
      <h3 className="font-semibold text-slate-800">{t("filters_title")}</h3>

      <div>
        <label className="label">{t("filters_city")}</label>
        <input
          className="input"
          placeholder={t("filters_city_ph")}
          value={form.city ?? ""}
          onChange={(e) => update("city", e.target.value || undefined)}
        />
      </div>

      <div>
        <label className="label">{t("filters_profiles")}</label>
        <div className="max-h-48 space-y-1 overflow-y-auto">
          {profileOptions.map((code) => {
            const checked = selectedProfiles.has(code);
            return (
              <label
                key={code}
                className={classNames(
                  "flex cursor-pointer items-start gap-2 rounded-md border px-2 py-2 text-sm hover:bg-slate-50",
                  checked ? "border-brand-200 bg-brand-50/40" : "border-slate-200",
                )}
              >
                <input
                  type="checkbox"
                  className="mt-0.5 shrink-0"
                  checked={checked}
                  onChange={() => toggleProfile(code)}
                />
                <span className="min-w-0 flex-1 break-words leading-snug text-slate-700">
                  {displayProfile(code, lang)}
                </span>
              </label>
            );
          })}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="label">{t("filters_price_from")}</label>
          <input
            type="number"
            min={0}
            className="input"
            value={form.price_min ?? ""}
            onChange={(e) => update("price_min", e.target.value ? Number(e.target.value) : undefined)}
          />
        </div>
        <div>
          <label className="label">{t("filters_price_to")}</label>
          <input
            type="number"
            min={0}
            className="input"
            value={form.price_max ?? ""}
            onChange={(e) => update("price_max", e.target.value ? Number(e.target.value) : undefined)}
          />
        </div>
      </div>

      <div>
        <label className="label">{t("filters_distance")}</label>
        <input
          type="number"
          min={0}
          step={0.1}
          className="input"
          value={form.max_distance_to_sea ?? ""}
          onChange={(e) =>
            update("max_distance_to_sea", e.target.value ? Number(e.target.value) : undefined)
          }
        />
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="label">{t("filters_check_in")}</label>
          <input
            type="date"
            className="input"
            value={form.check_in ?? ""}
            min={minCheckIn}
            onChange={(e) => {
              const next = e.target.value || undefined;
              update("check_in", next);
              if (next && form.check_out && form.check_out <= next) {
                update("check_out", plusDaysIsoDate(next, 1));
              }
              setDateError(null);
            }}
          />
        </div>
        <div>
          <label className="label">{t("filters_check_out")}</label>
          <input
            type="date"
            className="input"
            value={form.check_out ?? ""}
            min={minCheckOut}
            onChange={(e) => {
              update("check_out", e.target.value || undefined);
              setDateError(null);
            }}
          />
        </div>
      </div>
      <div className="flex flex-wrap gap-2">
        <button type="button" className="btn-secondary" onClick={() => applyDatePreset(3)}>
          {t("filters_dates_weekend")}
        </button>
        <button type="button" className="btn-secondary" onClick={() => applyDatePreset(7)}>
          {t("filters_dates_7")}
        </button>
        <button type="button" className="btn-secondary" onClick={() => applyDatePreset(14)}>
          {t("filters_dates_14")}
        </button>
        <button
          type="button"
          className="btn-secondary"
          onClick={() => {
            update("check_in", undefined);
            update("check_out", undefined);
            setDateError(null);
          }}
        >
          {t("filters_dates_clear")}
        </button>
      </div>
      {dateError && <p className="rounded-md bg-red-50 p-2 text-xs text-red-700">{dateError}</p>}

      <div>
        <label className="label">{t("filters_sort")}</label>
        <select
          className="input"
          value={form.sort ?? ""}
          onChange={(e) => update("sort", (e.target.value || undefined) as CatalogFilters["sort"])}
        >
          <option value="">{t("filters_sort_default")}</option>
          <option value="price_asc">{t("filters_sort_price_asc")}</option>
          <option value="price_desc">{t("filters_sort_price_desc")}</option>
          <option value="distance_asc">{t("filters_sort_dist_asc")}</option>
          <option value="distance_desc">{t("filters_sort_dist_desc")}</option>
        </select>
      </div>

      <div className="flex gap-2">
        <button type="submit" className="btn-primary flex-1">
          {t("filters_apply")}
        </button>
        <button type="button" onClick={handleReset} className="btn-secondary">
          {t("filters_reset")}
        </button>
      </div>
    </form>
  );
}
