import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { login, register } from "@/api/auth";
import { useExtractError } from "@/hooks/useExtractError";
import { useAuthStore } from "@/store/auth";
import ErrorAlert from "@/components/ErrorAlert";
import { useI18n } from "@/i18n";

export default function RegisterPage() {
  const { t } = useI18n();
  const toError = useExtractError();
  const navigate = useNavigate();
  const setSession = useAuthStore((s) => s.setSession);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const validate = (): string | null => {
    if (!fullName.trim()) return t("auth_err_name_required");
    if (!email.trim()) return t("auth_err_email_required");
    if (!email.includes("@") || email.trim().length < 3) return t("auth_err_email_invalid");
    if (!password) return t("auth_err_password_required");
    if (password.length < 8) return t("auth_err_password_short");
    return null;
  };

  const mutation = useMutation({
    mutationFn: async () => {
      await register({ email, password, full_name: fullName });
      return login(email, password);
    },
    onSuccess: (data) => {
      setErrorMsg(null);
      setSession(data.access_token, data.user);
      navigate("/", { replace: true });
    },
    onError: (err) => setErrorMsg(toError(err)),
  });

  return (
    <div className="mx-auto max-w-md py-12">
      <div className="card p-6">
        <h1 className="mb-4 text-2xl font-semibold">{t("auth_register_title")}</h1>
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
            <label className="label" htmlFor="fullName">{t("auth_full_name")}</label>
            <input id="fullName" className="input" value={fullName} onChange={(e) => setFullName(e.target.value)} />
          </div>
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
              autoComplete="new-password"
              className="input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <p className="mt-1 text-xs text-slate-500">{t("auth_password_hint")}</p>
          </div>
          {errorMsg && <ErrorAlert message={errorMsg} />}
          <button type="submit" className="btn-primary w-full" disabled={mutation.isPending}>
            {mutation.isPending ? t("auth_register_pending") : t("auth_register_submit")}
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-600">
          {t("auth_has_account")}{" "}
          <Link to="/login" className="font-medium text-brand-700 hover:underline">
            {t("auth_go_login")}
          </Link>
        </p>
      </div>
    </div>
  );
}
