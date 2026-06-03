import { Outlet } from "react-router-dom";
import Navbar from "./Navbar";
import { useI18n } from "@/i18n";

export default function Layout() {
  const { t } = useI18n();
  return (
    <div className="min-h-screen">
      <Navbar />
      <main className="mx-auto max-w-7xl px-4 py-6">
        <Outlet />
      </main>
      <footer className="mt-12 border-t border-slate-200 py-6 text-center text-xs text-slate-500">
        {t("footer_copyright", { year: new Date().getFullYear() })}
      </footer>
    </div>
  );
}
