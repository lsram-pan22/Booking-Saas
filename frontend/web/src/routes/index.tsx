import { Navigate, useRoutes } from "react-router-dom";
import { superAdminRoutes } from "./super-admin.routes";
import { tenantRoutes } from "./tenant.routes";

export default function AppRoutes() {
  return useRoutes([
    { path: "/", element: <Navigate to="/super-admin/login" replace /> },
    tenantRoutes,
    superAdminRoutes,
    { path: "*", element: <Navigate to="/super-admin/login" replace /> },
  ]);
}
