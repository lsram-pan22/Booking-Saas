import { Navigate } from "react-router-dom";
import type { RouteObject } from "react-router-dom";
import AdminLayout from "../components/layouts/super-admin/AdminLayout";
import ProtectedRoute from "../components/layouts/super-admin/ProtectedRoute";
import Login from "../pages/super-admin/login/Login";
import Dashboard from "../pages/super-admin/dashboard/Dashboard";
import Businesses from "../pages/super-admin/businesses/Businesses";
import Users from "../pages/super-admin/users/Users";
import Packages from "../pages/super-admin/packages/Packages";
import Subscriptions from "../pages/super-admin/subscriptions/Subscriptions";
import Reports from "../pages/super-admin/reports/Reports";
import Settings from "../pages/super-admin/settings/Settings";

export const superAdminRoutes: RouteObject = {
  path: "/super-admin",
  children: [
    { path: "login", element: <Login /> },
    {
      element: <ProtectedRoute />,
      children: [
        {
          element: <AdminLayout />,
          children: [
            { index: true, element: <Navigate to="dashboard" replace /> },
            { path: "dashboard", element: <Dashboard /> },
            { path: "businesses", element: <Businesses /> },
            { path: "users", element: <Users /> },
            { path: "packages", element: <Packages /> },
            { path: "subscriptions", element: <Subscriptions /> },
            { path: "reports", element: <Reports /> },
            { path: "settings", element: <Settings /> },
          ],
        },
      ],
    },
    { path: "*", element: <Navigate to="/super-admin/login" replace /> },
  ],
};
