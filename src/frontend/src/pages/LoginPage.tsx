import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { login } from "@/api/auth";
import { useExtractError } from "@/hooks/useExtractError";
import { useAuthStore } from "@/store/auth";
import ErrorAlert from "@/components/ErrorAlert";
import { useI18n } from "@/i18n";

export default function LoginPage() {
  const { t } = useI18n();
  const toError = useExtractError();
  const navigate = useNavigate();
  const location = useLocation();
  const setSession = useAuthStore((s) => s.setSession);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const validate = (): string | null => {
    if (!email.trim()) return t("auth_err_email_required");
    if (!email.includes("@") || email.trim().length < 3) return t("auth_err_email_invalid");
    if (!password) return t("auth_err_password_required");
    return null;
  };

  const mutation = useMutation({
    mutationFn: () => login(email, password),
    onSuccess: (data) => {
      setErrorMsg(null);
      setSession(data.access_token, data.user);
      const from = (location.state as { from?: string } | null)?.from;
      const defaultPath = data.user.role === "admin" ? "/admin/bookings" : "/";
      navigate(from && data.user.role === "client" ? from : defaultPath, { replace: true });
    },
    onError: (err) => setErrorMsg(toError(err)),
  });

  return (
    <div className="mx-auto max-w-md py-12">
      <div className="card p-6">
        <h1 className="mb-4 text-2xl font-semibold">{t("auth_login_title")}</h1>
        <form
          className="space-y-4"
          noValidate
          onSubmit={(e) => {
            e.preventDefault();
            const err = validate();
            if (err) {
              setErrorMsg(err);
              return;
            }
            setErrorMsg(null);
            mutation.mutate();
          }}
        >
          <div>
            <label className="label" htmlFor="email">{t("auth_email")}</label>
            <input
              id="email"
              type="email"
              autoComplete="email"
              className="input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <div>
            <label className="label" htmlFor="password">{t("auth_password")}</label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              className="input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          {errorMsg && <ErrorAlert message={errorMsg} />}
          <button type="submit" className="btn-primary w-full" disabled={mutation.isPending}>
            {mutation.isPending ? t("auth_login_pending") : t("auth_login_submit")}
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-600">
          {t("auth_no_account")}{" "}
          <Link to="/register" className="font-medium text-brand-700 hover:underline">
            {t("auth_go_register")}
          </Link>
        </p>
        <p className="mt-3 rounded-md bg-slate-50 p-2 text-center text-xs text-slate-500">{t("auth_login_demo")}</p>
      </div>
    </div>
  );
}
