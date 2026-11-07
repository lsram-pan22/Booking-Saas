import { Navigate, Outlet } from "react-router-dom";
import { isTenantAuthenticated } from "../../../utils/auth";

export default function ProtectedRoute() {
  return isTenantAuthenticated() ? <Outlet /> : <Navigate to="/tenant/login" replace />;
}
