
import { Navigate, Route, Routes } from "react-router-dom";

import AdminLayout from "../components/layouts/AdminLayout";
import ProtectedRoute from "../components/layouts/ProtectedRoute";

import Login from "../pages/super-admin/login/Login";
import Dashboard from "../pages/super-admin/dashboard/Dashboard";
import Businesses from "../pages/super-admin/businesses/Businesses";
import Users from "../pages/super-admin/users/Users";
import Packages from "../pages/super-admin/packages/Packages";
import Subscriptions from "../pages/super-admin/subscriptions/Subscriptions";
import Reports from "../pages/super-admin/reports/Reports";
import Settings from "../pages/super-admin/settings/Settings";

export default function AppRoutes() {
  return (
    <Routes>
      <Route
        path="/"
        element={<Navigate to="/super-admin/login" replace />}
      />

      <Route
        path="/super-admin/login"
        element={<Login />}
      />

      <Route element={<ProtectedRoute />}>
        <Route path="/super-admin" element={<AdminLayout />}>
          <Route
            index
            element={<Navigate to="dashboard" replace />}
          />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="businesses" element={<Businesses />} />
          <Route path="users" element={<Users />} />
          <Route path="packages" element={<Packages />} />
          <Route path="subscriptions" element={<Subscriptions />} />
          <Route path="reports" element={<Reports />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Route>

      <Route
        path="*"
        element={<Navigate to="/super-admin/login" replace />}
      />
    </Routes>
  );
}
