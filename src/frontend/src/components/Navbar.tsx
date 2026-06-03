import { Link, NavLink, useLocation, useNavigate } from "react-router-dom";
import { useAuthStore } from "@/store/auth";
import { classNames } from "@/lib/utils";
import { userLoginLabel } from "@/lib/user";
import { useI18n } from "@/i18n";

const linkClass = ({ isActive }: { isActive: boolean }) =>
  classNames(
    "rounded-md px-3 py-2 text-sm font-medium transition",
    isActive ? "bg-brand-50 text-brand-700" : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
  );

export default function Navbar() {
  const { user, logout } = useAuthStore();
  const { lang, setLang, t } = useI18n();
  const navigate = useNavigate();
  const location = useLocation();
  const isStaff = user?.role === "admin";
  const isClient = user?.role === "client";

  return (
    <header className="sticky top-0 z-10 border-b border-slate-200 bg-white/80 backdrop-blur">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3">
        <Link to="/" className="flex items-center gap-2 text-lg font-semibold text-brand-700">
          <span className="inline-block h-2 w-2 rounded-full bg-brand-600" />
          Sanatorium
        </Link>
        <nav className="flex items-center gap-1">
          <NavLink to="/" end className={linkClass}>
            {t("nav_catalog")}
          </NavLink>
          {isClient && (
            <NavLink to="/bookings" className={linkClass}>
              {t("nav_bookings")}
            </NavLink>
          )}
          {isStaff && (
            <NavLink
              to="/admin/bookings"
              className={({ isActive }) =>
                linkClass({ isActive: isActive || location.pathname.startsWith("/admin") })
              }
            >
              {t("nav_admin")}
            </NavLink>
          )}
        </nav>
        <div className="flex items-center gap-3">
          <div className="inline-flex rounded-md border border-slate-300 bg-white p-0.5">
            <button
              type="button"
              className={classNames(
                "rounded px-2 py-1 text-xs font-medium",
                lang === "ru" ? "bg-brand-600 text-white" : "text-slate-600 hover:bg-slate-100",
              )}
              onClick={() => setLang("ru")}
            >
              RU
            </button>
            <button
              type="button"
              className={classNames(
                "rounded px-2 py-1 text-xs font-medium",
                lang === "en" ? "bg-brand-600 text-white" : "text-slate-600 hover:bg-slate-100",
              )}
              onClick={() => setLang("en")}
            >
              EN
            </button>
          </div>
          {user ? (
            <>
              <div className="text-right text-sm leading-tight">
                <div className="font-medium text-slate-800">{userLoginLabel(user)}</div>
                <div className="text-xs text-slate-500">{t(`role_${user.role}`)}</div>
              </div>
              <button
                type="button"
                className="btn-secondary"
                onClick={() => {
                  logout();
                  navigate("/login");
                }}
              >
                {t("nav_logout")}
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="btn-secondary">
                {t("nav_login")}
              </Link>
              <Link to="/register" className="btn-primary">
                {t("nav_register")}
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
