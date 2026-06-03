import { useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { listSanatoriums } from "@/api/sanatoriums";
import SanatoriumCard from "@/components/SanatoriumCard";
import CatalogFiltersPanel from "@/components/CatalogFiltersPanel";
import Pagination from "@/components/Pagination";
import Spinner from "@/components/Spinner";
import ErrorAlert from "@/components/ErrorAlert";
import { useExtractError } from "@/hooks/useExtractError";
import type { CatalogFilters } from "@/types";
import { localizeFiltersForUi, normalizeFiltersForApi } from "@/lib/catalogFilters";
import { useI18n } from "@/i18n";

function readFilters(params: URLSearchParams): CatalogFilters {
  const num = (key: string) => {
    const v = params.get(key);
    return v ? Number(v) : undefined;
  };
  return {
    page: num("page") ?? 1,
    page_size: num("page_size") ?? 12,
    city: params.get("city") ?? undefined,
    profiles: params.get("profiles") ?? undefined,
    max_distance_to_sea: num("max_distance_to_sea"),
    price_min: num("price_min"),
    price_max: num("price_max"),
    check_in: params.get("check_in") ?? undefined,
    check_out: params.get("check_out") ?? undefined,
    sort: (params.get("sort") as CatalogFilters["sort"]) ?? undefined,
  };
}

export default function CatalogPage() {
  const { t, lang } = useI18n();
  const toError = useExtractError();
  const [params, setParams] = useSearchParams();
  const filters = useMemo(() => localizeFiltersForUi(readFilters(params), lang), [params, lang]);
  const [showFilters, setShowFilters] = useState(false);
  const apiFilters = useMemo(() => normalizeFiltersForApi(filters, lang), [filters, lang]);

  const query = useQuery({
    queryKey: ["sanatoriums", apiFilters],
    queryFn: () => listSanatoriums(apiFilters),
    placeholderData: keepPreviousData,
  });

  const applyFilters = (next: CatalogFilters) => {
    const searchInit: Record<string, string> = {};
    for (const [k, v] of Object.entries(next)) {
      if (v === undefined || v === null || v === "") continue;
      searchInit[k] = String(v);
    }
    setParams(searchInit);
  };

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[280px_1fr]">
      <aside className={`${showFilters ? "block" : "hidden"} lg:block`}>
        <CatalogFiltersPanel initial={filters} onApply={applyFilters} />
      </aside>
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h1 className="text-2xl font-semibold">{t("catalog_title")}</h1>
          <button
            type="button"
            className="btn-secondary lg:hidden"
            onClick={() => setShowFilters((v) => !v)}
          >
            {showFilters ? t("catalog_hide_filters") : t("filters_title")}
          </button>
        </div>

        {query.isLoading && <Spinner />}
        {query.isError && <ErrorAlert message={toError(query.error)} />}

        {query.data && (
          <>
            <div className="mb-3 text-sm text-slate-500">{t("catalog_found", { count: query.data.total })}</div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {query.data.items.map((s) => (
                <SanatoriumCard key={s.id} item={s} />
              ))}
            </div>
            {query.data.items.length === 0 && (
              <div className="py-12 text-center text-slate-500">{t("catalog_nothing_found")}</div>
            )}
            <Pagination
              page={query.data.page}
              totalPages={query.data.total_pages}
              onChange={(page) => applyFilters({ ...filters, page })}
            />
          </>
        )}
      </section>
    </div>
  );
}
