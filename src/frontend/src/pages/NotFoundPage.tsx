import { Link } from "react-router-dom";
import { useI18n } from "@/i18n";

export default function NotFoundPage() {
  const { t } = useI18n();
  return (
    <div className="mx-auto max-w-md py-20 text-center">
      <h1 className="text-4xl font-bold text-slate-800">404</h1>
      <p className="mt-2 text-slate-500">{t("notfound_title")}</p>
      <Link to="/" className="btn-primary mt-6 inline-flex">
        {t("notfound_home")}
      </Link>
    </div>
  );
}
