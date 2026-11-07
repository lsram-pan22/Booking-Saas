import { Navigate } from "react-router-dom";
import type { RouteObject } from "react-router-dom";
import ProtectedRoute from "../components/layouts/tenant/ProtectedRoute";
import TenantLayout from "../components/layouts/tenant/TenantLayout";
import Register from "../pages/tenant/register/Register";
import Login from "../pages/tenant/login/Login";
import Dashboard from "../pages/tenant/dashboard/Dashboard";
import Packages from "../pages/tenant/packages/Packages";
import Settings from "../pages/tenant/settings/Settings";
import Shop from "../pages/tenant/shops/Shop";
import OwnerRoute from "../components/layouts/tenant/OwnerRoute";
import Home from "../pages/tenant/home/Home";
import ShopDashboard from "../pages/tenant/shops/Dashboard";
import Services from "../pages/tenant/services/Services";

export const tenantRoutes: RouteObject = {
  path: "/tenant",
  children: [
    { index: true, element: <Navigate to="home" replace /> },
    { path: "register", element: <Register /> },
    { path: "login", element: <Login /> },
    {
      element: <ProtectedRoute />,
      children: [{ element: <TenantLayout />, children: [
        { path: "home", element: <Home /> },
        { path: "staff", element: <ShopDashboard /> },
        { path: "shops/:shopID/dashboard", element: <ShopDashboard /> },
        { path: "shops/:shopID/services", element: <Services /> },
        { element: <OwnerRoute />, children: [
          { path: "dashboard", element: <Dashboard /> },
          { path: "packages", element: <Packages /> },
          { path: "shops/:shopID", element: <Shop /> },
        ] },
        { path: "settings", element: <Settings /> },
      ] }],
    },
    { path: "*", element: <Navigate to="/tenant/login" replace /> },
  ],
};
