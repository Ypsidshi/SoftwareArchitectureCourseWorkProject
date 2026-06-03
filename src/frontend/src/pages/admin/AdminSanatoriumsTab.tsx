import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createAdminSanatorium,
  deleteAdminSanatorium,
  listAdminSanatoriums,
  updateAdminSanatorium,
  type SanatoriumInput,
} from "@/api/admin";
import { useExtractError } from "@/hooks/useExtractError";
import { useConfirm } from "@/hooks/useConfirm";
import Spinner from "@/components/Spinner";
import ErrorAlert from "@/components/ErrorAlert";
import Pagination from "@/components/Pagination";
import Modal from "@/components/Modal";
import { useFormatters } from "@/hooks/useFormatters";
import { useI18n } from "@/i18n";
import { profileOptions, profileSlugFromInput, profilesInputFromSlugs } from "@/lib/medicalProfiles";
import MedicalProfileChip from "@/components/MedicalProfileChip";
import type { Sanatorium } from "@/types";

const emptyForm = (): SanatoriumInput => ({
  name: "",
  description: "",
  city: "",
  address: "",
  distance_to_sea_km: 0,
  amenities: [],
  image_urls: [],
  price_per_night: 0,
  total_places: 1,
  medical_profiles: [],
});

export default function AdminSanatoriumsTab() {
  const { t, lang } = useI18n();
  const askConfirm = useConfirm();
  const toError = useExtractError();
  const { formatCurrency } = useFormatters();
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Sanatorium | null>(null);
  const [form, setForm] = useState<SanatoriumInput>(emptyForm());
  const [amenitiesText, setAmenitiesText] = useState("");
  const [profilesText, setProfilesText] = useState("");
  const [error, setError] = useState<string | null>(null);

  const profileChoices = profileOptions(lang);

  const query = useQuery({
    queryKey: ["admin-sanatoriums", page],
    queryFn: () => listAdminSanatoriums(page, 10),
  });

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
    setForm(emptyForm());
    setAmenitiesText("");
    setProfilesText("");
    setError(null);
  };

  const openCreate = () => {
    setEditing(null);
    setForm(emptyForm());
    setAmenitiesText("");
    setProfilesText("");
    setError(null);
    setModalOpen(true);
  };

  const openEdit = (s: Sanatorium) => {
    setEditing(s);
    setForm({
      name: s.name,
      description: s.description,
      city: s.city,
      address: s.address,
      distance_to_sea_km: s.distance_to_sea_km,
      amenities: s.amenities,
      image_urls: s.image_urls,
      price_per_night: s.price_per_night,
      total_places: s.total_places,
      latitude: s.latitude,
      longitude: s.longitude,
      medical_profiles: s.medical_profiles,
    });
    setAmenitiesText(s.amenities.join(", "));
    setProfilesText(profilesInputFromSlugs(s.medical_profiles, lang));
    setError(null);
    setModalOpen(true);
  };

  const buildPayload = (): SanatoriumInput => ({
    ...form,
    amenities: amenitiesText.split(",").map((x) => x.trim()).filter(Boolean),
    medical_profiles: profilesText
      .split(",")
      .map((x) => profileSlugFromInput(x, lang))
      .filter(Boolean),
  });

  const save = useMutation({
    mutationFn: async () => {
      const payload = buildPayload();
      if (editing) return updateAdminSanatorium(editing.id, payload);
      return createAdminSanatorium(payload);
    },
    onSuccess: () => {
      closeModal();
      qc.invalidateQueries({ queryKey: ["admin-sanatoriums"] });
      qc.invalidateQueries({ queryKey: ["sanatoriums"] });
    },
    onError: (e) => setError(toError(e)),
  });

  const remove = useMutation({
    mutationFn: deleteAdminSanatorium,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-sanatoriums"] });
      qc.invalidateQueries({ queryKey: ["sanatoriums"] });
    },
    onError: (e) => setError(toError(e)),
  });

  const handleCloseModal = () => {
    const dirty =
      form.name ||
      form.city ||
      form.address ||
      form.description ||
      amenitiesText ||
      profilesText ||
      form.price_per_night > 0;
    if (dirty && !askConfirm("admin_san_discard_confirm")) return;
    closeModal();
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const key = editing ? "admin_san_save_confirm_edit" : "admin_san_save_confirm_create";
    if (!askConfirm(key)) return;
    save.mutate();
  };

  const profileHint = profileChoices.map((p) => p.label).join(", ");

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-slate-600">{t("admin_san_list_hint")}</p>
        <button type="button" className="btn-primary" onClick={openCreate}>
          {t("admin_san_add")}
        </button>
      </div>

      {error && !modalOpen && <ErrorAlert message={error} className="mb-4" />}

      {query.isLoading && <Spinner />}
      {query.isError && <ErrorAlert message={toError(query.error)} />}
      {!query.isLoading && query.data?.items.length === 0 && (
        <div className="card p-6 text-center text-sm text-slate-500">{t("admin_san_empty")}</div>
      )}

      <div className="space-y-3">
        {query.data?.items.map((s) => (
          <div key={s.id} className="card p-4">
            <div className="flex justify-between gap-2">
              <div>
                <h3 className="font-semibold">{s.name}</h3>
                <p className="text-sm text-slate-600">
                  {s.city} · {formatCurrency(s.price_per_night)} / {t("card_per_night")}
                </p>
                <p className="text-xs text-slate-500">
                  {t("admin_san_places")}: {s.total_places} · {t("admin_distance_km", { km: s.distance_to_sea_km })}
                </p>
                  {s.medical_profiles.length > 0 && (
                    <div className="mt-2 flex flex-wrap gap-1">
                      {s.medical_profiles.map((slug) => (
                        <MedicalProfileChip key={slug} slug={slug} />
                      ))}
                    </div>
                  )}
              </div>
              <div className="flex flex-col gap-2">
                <button type="button" className="btn-secondary text-sm" onClick={() => openEdit(s)}>
                  {t("admin_edit")}
                </button>
                <button
                  type="button"
                  className="btn-danger text-sm"
                  disabled={remove.isPending}
                  onClick={() => {
                    if (askConfirm("admin_san_delete_confirm")) remove.mutate(s.id);
                  }}
                >
                  {t("admin_delete")}
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {query.data && (
        <Pagination page={query.data.page} totalPages={query.data.total_pages} onChange={setPage} />
      )}

      <Modal
        open={modalOpen}
        title={editing ? t("admin_san_edit") : t("admin_san_create")}
        onClose={handleCloseModal}
      >
        <form className="space-y-3" onSubmit={handleSubmit}>
          <div>
            <label className="label">{t("admin_san_name")}</label>
            <input className="input" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </div>
          <div>
            <label className="label">{t("admin_san_city")}</label>
            <input className="input" required value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
          </div>
          <div>
            <label className="label">{t("admin_san_address")}</label>
            <input className="input" required value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
          </div>
          <div>
            <label className="label">{t("admin_san_description")}</label>
            <textarea className="input min-h-[80px]" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="label">{t("admin_san_price")}</label>
              <input type="number" min={1} required className="input" value={form.price_per_night || ""} onChange={(e) => setForm({ ...form, price_per_night: Number(e.target.value) })} />
            </div>
            <div>
              <label className="label">{t("admin_san_places")}</label>
              <input type="number" min={1} required className="input" value={form.total_places || ""} onChange={(e) => setForm({ ...form, total_places: Number(e.target.value) })} />
            </div>
          </div>
          <div>
            <label className="label">{t("filters_distance")}</label>
            <input type="number" min={0} step={0.1} className="input" value={form.distance_to_sea_km} onChange={(e) => setForm({ ...form, distance_to_sea_km: Number(e.target.value) })} />
          </div>
          <div>
            <label className="label">{t("details_amenities")} ({t("admin_comma_sep")})</label>
            <input className="input" value={amenitiesText} onChange={(e) => setAmenitiesText(e.target.value)} placeholder={t("admin_san_amenities_ph")} />
          </div>
          <div>
            <label className="label">{t("filters_profiles")} ({t("admin_comma_sep")})</label>
            <input
              className="input"
              list="medical-profiles-list"
              value={profilesText}
              onChange={(e) => setProfilesText(e.target.value)}
              placeholder={profileHint}
            />
            <datalist id="medical-profiles-list">
              {profileChoices.map((p) => (
                <option key={p.slug} value={p.label} />
              ))}
            </datalist>
            <p className="mt-1 text-xs text-slate-500">{t("admin_san_profiles_hint", { names: profileHint })}</p>
          </div>
          {error && <ErrorAlert message={error} />}
          <div className="flex gap-2 pt-1">
            <button type="submit" className="btn-primary flex-1" disabled={save.isPending}>
              {save.isPending ? t("admin_saving") : t("admin_save")}
            </button>
            <button type="button" className="btn-secondary" onClick={handleCloseModal}>
              {t("admin_cancel_edit")}
            </button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
