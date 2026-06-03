import { Navigate, useLocation } from "react-router-dom";
import { useAuthStore } from "@/store/auth";
import type { UserRole } from "@/types";
import type { ReactNode } from "react";

interface Props {
  children: ReactNode;
  roles?: UserRole[];
}

export default function ProtectedRoute({ children, roles }: Props) {
  const location = useLocation();
  const { token, user } = useAuthStore();

  if (!token || !user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  if (roles && roles.length > 0 && !roles.includes(user.role)) {
    const fallback = user.role === "admin" ? "/admin/bookings" : "/";
    return <Navigate to={fallback} replace />;
  }
  return <>{children}</>;
}
