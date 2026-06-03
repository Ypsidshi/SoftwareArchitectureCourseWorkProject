import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getSanatorium } from "@/api/sanatoriums";
import { createBooking } from "@/api/bookings";
import { useExtractError } from "@/hooks/useExtractError";
import { useAuthStore } from "@/store/auth";
import { daysBetween, plusDaysIsoDate, todayIsoDate } from "@/lib/utils";
import { useFormatters } from "@/hooks/useFormatters";
import Spinner from "@/components/Spinner";
import ErrorAlert from "@/components/ErrorAlert";
import { getSanatoriumCover, getSanatoriumFallback, getSanatoriumGallery } from "@/lib/sanatoriumImages";
import { useI18n } from "@/i18n";
import { localizeSanatorium } from "@/lib/sanatoriumLocalization";
import MedicalProfileChip from "@/components/MedicalProfileChip";

export default function SanatoriumDetailPage() {
  const { t, lang } = useI18n();
  const toError = useExtractError();
  const { formatCurrency, formatDate } = useFormatters();
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const isClient = user?.role === "client";

  const [checkIn, setCheckIn] = useState("");
  const [checkOut, setCheckOut] = useState("");
  const [guests, setGuests] = useState(1);
  const [coverImage, setCoverImage] = useState("");
  const [bookingError, setBookingError] = useState<string | null>(null);
  const minCheckIn = todayIsoDate();
  const minCheckOut = checkIn ? plusDaysIsoDate(checkIn, 1) : todayIsoDate();
  const hasInvalidDates = !!checkIn && !!checkOut && (checkIn < minCheckIn || checkOut <= checkIn);

  const range =
    checkIn && checkOut && checkIn < checkOut
      ? { check_in: checkIn, check_out: checkOut }
      : undefined;

  const query = useQuery({
    queryKey: ["sanatorium", id, range],
    queryFn: () => getSanatorium(id, range),
    enabled: !!id,
  });

  const booking = useMutation({
    mutationFn: () =>
      createBooking({
        sanatorium_id: id,
        check_in: checkIn,
        check_out: checkOut,
        guests,
      }),
    onSuccess: () => {
      setBookingError(null);
      qc.invalidateQueries({ queryKey: ["bookings"] });
      navigate("/bookings");
    },
    onError: (err) => {
      setBookingError(toError(err));
    },
  });

  const nights = useMemo(
    () => (range ? daysBetween(range.check_in, range.check_out) : 0),
    [range],
  );

  useEffect(() => {
    if (!query.data?.sanatorium) return;
    setCoverImage(getSanatoriumCover(query.data.sanatorium));
  }, [query.data?.sanatorium]);

  if (query.isLoading) return <Spinner />;
  if (query.isError) return <ErrorAlert message={toError(query.error)} />;
  if (!query.data) return null;

  const { sanatorium, available } = query.data;
  const localized = localizeSanatorium(sanatorium, lang);
  const fallbackCover = getSanatoriumFallback(sanatorium);
  const galleryImages = getSanatoriumGallery(sanatorium);
  const total = nights > 0 ? sanatorium.price_per_night * nights : 0;
  const canBook =
    isClient && !!range && nights > 0 && guests > 0 && guests <= sanatorium.total_places && !hasInvalidDates;

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1.6fr_1fr]">
      <div className="space-y-4">
        <div className="card overflow-hidden">
          <div className="aspect-video w-full bg-slate-100">
            {coverImage ? (
              <img
                src={coverImage}
                alt={localized.name}
                className="h-full w-full object-cover"
                onError={() => {
                  if (coverImage !== fallbackCover) setCoverImage(fallbackCover);
                }}
              />
            ) : (
              <div className="flex h-full items-center justify-center text-slate-400">{t("card_no_photo")}</div>
            )}
          </div>
          <div className="p-5">
            <h1 className="text-2xl font-semibold">{localized.name}</h1>
            <div className="mt-1 text-sm text-slate-500">
              {localized.city} · {localized.address} · {sanatorium.distance_to_sea_km} км до моря
            </div>
            <p className="mt-4 whitespace-pre-line text-slate-700">{localized.description}</p>
          </div>
        </div>

        <div className="card p-5">
          <h2 className="mb-2 font-semibold">{t("details_profiles")}</h2>
          <div className="flex flex-wrap gap-2">
            {sanatorium.medical_profiles.map((slug) => (
              <MedicalProfileChip key={slug} slug={slug} />
            ))}
          </div>
          <h2 className="mb-2 mt-4 font-semibold">{t("details_amenities")}</h2>
          <div className="flex flex-wrap gap-2">
            {localized.amenities.map((a) => (
              <span key={a} className="badge bg-slate-100 text-slate-700">
                {a}
              </span>
            ))}
          </div>
        </div>

        {galleryImages.length > 0 && (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {galleryImages.map((url) => (
              <img
                key={`${sanatorium.id}-${url}`}
                src={url}
                alt=""
                className="aspect-[4/3] w-full rounded-md object-cover"
                onError={(e) => {
                  const target = e.currentTarget;
                  const fallback = getSanatoriumFallback(sanatorium, 2);
                  if (!target.src.endsWith(fallback)) {
                    target.src = fallback;
                    return;
                  }
                  target.style.display = "none";
                }}
              />
            ))}
          </div>
        )}
      </div>

      <aside className="h-fit space-y-4">
        <div className="card p-5">
          <div className="flex items-baseline gap-2">
            <span className="text-2xl font-semibold text-brand-700">
              {formatCurrency(sanatorium.price_per_night)}
            </span>
            <span className="text-sm text-slate-500">{t("card_per_night")}</span>
          </div>
          <div className="mt-4 space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="label">{t("filters_check_in")}</label>
                <input
                  type="date"
                  className="input"
                  min={minCheckIn}
                  value={checkIn}
                  onChange={(e) => {
                    const next = e.target.value;
                    setCheckIn(next);
                    if (checkOut && checkOut <= next) {
                      setCheckOut(plusDaysIsoDate(next, 1));
                    }
                  }}
                />
              </div>
              <div>
                <label className="label">{t("filters_check_out")}</label>
                <input
                  type="date"
                  className="input"
                  min={minCheckOut}
                  value={checkOut}
                  onChange={(e) => setCheckOut(e.target.value)}
                />
              </div>
            </div>
            <div>
              <label className="label">{t("details_guests")}</label>
              <input
                type="number"
                min={1}
                max={sanatorium.total_places || 10}
                className="input"
                value={guests}
                onChange={(e) => setGuests(Number(e.target.value))}
              />
              <p className="mt-1 text-xs text-slate-500">
                {t("details_max_places", { count: sanatorium.total_places })}
              </p>
            </div>
          </div>
          {hasInvalidDates && (
            <p className="mt-3 rounded-md bg-red-50 p-2 text-xs text-red-700">{t("details_invalid_dates")}</p>
          )}

          {range && (
            <div className="mt-4 rounded-md bg-slate-50 p-3 text-sm">
              <div className="flex justify-between">
                <span>{formatDate(range.check_in)} — {formatDate(range.check_out)}</span>
                <span>{t("details_nights", { count: nights })}</span>
              </div>
              <div className="mt-1 flex justify-between font-semibold">
                <span>{t("details_total")}</span>
                <span>{formatCurrency(total)}</span>
              </div>
              <div className="mt-1 text-xs">
                {available ? (
                  <span className="text-emerald-700">{t("details_available")}</span>
                ) : (
                  <span className="text-red-700">{t("details_not_available")}</span>
                )}
              </div>
            </div>
          )}

          {bookingError && (
            <div className="mt-3">
              <ErrorAlert message={bookingError} />
            </div>
          )}

          {!user && (
            <button
              type="button"
              className="btn-primary mt-4 w-full"
              onClick={() => navigate("/login", { state: { from: `/sanatoriums/${id}` } })}
            >
              {t("details_login_to_book")}
            </button>
          )}
          {user && !isClient && (
            <p className="mt-4 rounded-md bg-amber-50 p-3 text-xs text-amber-800">
              {t("details_only_clients")}
            </p>
          )}
          {isClient && (
            <button
              type="button"
              className="btn-primary mt-4 w-full"
              disabled={!canBook || !available || booking.isPending}
              onClick={() => booking.mutate()}
            >
              {booking.isPending ? t("details_creating") : t("details_create_booking")}
            </button>
          )}
        </div>
      </aside>
    </div>
  );
}
