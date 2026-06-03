import { displayProfile } from "@/lib/medicalProfiles";
import { useI18n } from "@/i18n";
import { classNames } from "@/lib/utils";

/** Иконки отключены; SVG-заготовки — в `medicalProfileIcons.tsx`. */

type Variant = "color" | "mono";

interface Props {
  slug: string;
  variant?: Variant;
  className?: string;
}

export default function MedicalProfileChip({ slug, variant = "color", className }: Props) {
  const { lang } = useI18n();
  const key = slug.trim().toLowerCase();
  const label = displayProfile(key, lang);

  const styles =
    variant === "color"
      ? "bg-brand-50 text-brand-800 ring-1 ring-brand-100"
      : "bg-slate-100 text-slate-700 ring-1 ring-slate-200";

  return (
    <span
      className={classNames(
        "inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium",
        styles,
        className,
      )}
      title={label}
    >
      {label}
    </span>
  );
}
