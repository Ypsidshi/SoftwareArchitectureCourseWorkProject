import { NavLink, Outlet, Navigate } from "react-router-dom";
import { classNames } from "@/lib/utils";
import { useI18n } from "@/i18n";

const tabClass = ({ isActive }: { isActive: boolean }) =>
  classNames(
    "rounded-md px-4 py-2 text-sm font-medium transition",
    isActive ? "bg-brand-600 text-white" : "bg-slate-100 text-slate-700 hover:bg-slate-200",
  );

export default function AdminLayout() {
  const { t } = useI18n();

  return (
    <div>
      <h1 className="mb-4 text-2xl font-semibold">{t("admin_panel_title")}</h1>
      <nav className="mb-6 flex flex-wrap gap-2 overflow-x-auto pb-1">
        <NavLink to="/admin/bookings" className={tabClass}>
          {t("admin_tab_bookings")}
        </NavLink>
        <NavLink to="/admin/sanatoriums" className={tabClass}>
          {t("admin_tab_sanatoriums")}
        </NavLink>
      </nav>
      <Outlet />
    </div>
  );
}

export function AdminIndexRedirect() {
  return <Navigate to="/admin/bookings" replace />;
}
