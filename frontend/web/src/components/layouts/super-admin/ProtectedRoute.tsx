
import { Navigate, Outlet } from "react-router-dom";
import { isAuthenticated } from "../../../utils/auth";

export default function ProtectedRoute() {
  if (!isAuthenticated("super_admin")) {
    return (
      <Navigate
        to="/super-admin/login"
        replace
      />
    );
  }

  return <Outlet />;
}
