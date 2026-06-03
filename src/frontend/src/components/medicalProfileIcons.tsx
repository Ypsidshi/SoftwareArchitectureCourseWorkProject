import type { ReactNode } from "react";
import type { MedicalProfileSlug } from "@/lib/medicalProfiles";

type IconProps = { className?: string };

function IconBase({ className, children }: IconProps & { children: ReactNode }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden
    >
      {children}
    </svg>
  );
}

const icons: Record<MedicalProfileSlug, (p: IconProps) => ReactNode> = {
  cardiology: (p) => (
    <IconBase {...p}>
      <path d="M12 20c-3-2.5-6-5.2-6-9a4 4 0 0 1 7-2 4 4 0 0 1 7 2c0 3.8-3 6.5-6 9z" />
    </IconBase>
  ),
  pulmonology: (p) => (
    <IconBase {...p}>
      <path d="M8 4c0 2 1 3 3 3h2c2 0 3-1 3-3" />
      <path d="M11 7v14M8 21h6" />
      <path d="M14 10c2 0 3 1 3 3v2c0 2-1 3-3 3" />
    </IconBase>
  ),
  musculoskeletal: (p) => (
    <IconBase {...p}>
      <path d="M8 12l4-6 4 6-4 6z" />
      <path d="M12 6v12" />
    </IconBase>
  ),
  neurology: (p) => (
    <IconBase {...p}>
      <path d="M9 4c0 2 1.5 3 3.5 3S16 6 16 4" />
      <path d="M12 7c3 0 5 2 5 5 0 4-3 7-5 7s-5-3-5-7c0-3 2-5 5-5z" />
    </IconBase>
  ),
  gastroenterology: (p) => (
    <IconBase {...p}>
      <ellipse cx="12" cy="14" rx="6" ry="5" />
      <path d="M9 9c0-3 1.5-5 3-5s3 2 3 5" />
    </IconBase>
  ),
  endocrinology: (p) => (
    <IconBase {...p}>
      <path d="M12 3v4" />
      <path d="M8 7h8" />
      <path d="M10 11c0 3 1 5 2 8s2 5 2 8-1 5-2 8-2-5-2-8 1-5 2-8 2-5 2-8z" />
    </IconBase>
  ),
  dermatology: (p) => (
    <IconBase {...p}>
      <path d="M6 8c0-2 2-4 6-4s6 2 6 4-2 4-6 6-6-4-6-6z" />
      <path d="M6 16c0 2 2 4 6 4s6-2 6-4" />
    </IconBase>
  ),
  urology: (p) => (
    <IconBase {...p}>
      <path d="M9 6c0-2 1.5-3 3-3s3 1 3 3v12c0 2-1.5 3-3 3s-3-1-3-3" />
      <path d="M15 10h4l-2 4 2 4h-4" />
    </IconBase>
  ),
  pediatrics: (p) => (
    <IconBase {...p}>
      <circle cx="12" cy="7" r="2.5" />
      <path d="M8 20v-5c0-2 1.8-3 4-3s4 1 4 3v5" />
      <circle cx="17" cy="9" r="1.5" />
      <path d="M17 11.5v3" />
    </IconBase>
  ),
  balneology: (p) => (
    <IconBase {...p}>
      <path d="M4 14c2-2 4-2 6 0s4 2 6 0 4-2 6 0" />
      <path d="M6 18h12" />
      <path d="M12 4v3M10 6h4" />
    </IconBase>
  ),
  rehabilitation: (p) => (
    <IconBase {...p}>
      <path d="M12 6v12" />
      <path d="M8 10l4-4 4 4" />
      <path d="M16 14l-4 4-4-4" />
    </IconBase>
  ),
  general_therapy: (p) => (
    <IconBase {...p}>
      <path d="M12 6v12M6 12h12" />
      <rect x="4" y="4" width="16" height="16" rx="3" />
    </IconBase>
  ),
};

export function MedicalProfileIcon({
  slug,
  className = "h-4 w-4 shrink-0",
}: {
  slug: string;
  className?: string;
}) {
  const key = slug.trim().toLowerCase() as MedicalProfileSlug;
  const Icon = icons[key];
  if (!Icon) {
    return (
      <IconBase className={className}>
        <circle cx="12" cy="12" r="4" />
      </IconBase>
    );
  }
  return <Icon className={className} />;
}
