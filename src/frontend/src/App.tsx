import { Navigate, Route, Routes } from "react-router-dom";
import Layout from "@/components/Layout";
import ProtectedRoute from "@/components/ProtectedRoute";
import CatalogPage from "@/pages/CatalogPage";
import SanatoriumDetailPage from "@/pages/SanatoriumDetailPage";
import LoginPage from "@/pages/LoginPage";
import RegisterPage from "@/pages/RegisterPage";
import MyBookingsPage from "@/pages/MyBookingsPage";
import AdminLayout, { AdminIndexRedirect } from "@/pages/admin/AdminLayout";
import AdminBookingsTab from "@/pages/admin/AdminBookingsTab";
import AdminSanatoriumsTab from "@/pages/admin/AdminSanatoriumsTab";
import NotFoundPage from "@/pages/NotFoundPage";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<CatalogPage />} />
        <Route path="sanatoriums/:id" element={<SanatoriumDetailPage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route
          path="bookings"
          element={
            <ProtectedRoute roles={["client"]}>
              <MyBookingsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="admin"
          element={
            <ProtectedRoute roles={["admin"]}>
              <AdminLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<AdminIndexRedirect />} />
          <Route path="bookings" element={<AdminBookingsTab />} />
          <Route path="sanatoriums" element={<AdminSanatoriumsTab />} />
        </Route>
        <Route path="404" element={<NotFoundPage />} />
        <Route path="*" element={<Navigate to="/404" replace />} />
      </Route>
    </Routes>
  );
}
