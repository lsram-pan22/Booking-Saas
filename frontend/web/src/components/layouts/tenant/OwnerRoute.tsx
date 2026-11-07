import { Alert, Spin } from "antd";
import { Navigate, Outlet, useOutletContext } from "react-router-dom";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";
import { getTenantLandingPath } from "../../../utils/tenant-access";

export default function OwnerRoute() {
  const context = useOutletContext<TenantLayoutContext>();
  if (context.error) return <Alert type="error" title={context.error} showIcon />;
  if (!context.profile) return <Spin />;
  return context.profile.workspace_role === "owner" ? <Outlet context={context} /> : <Navigate to={getTenantLandingPath(context.profile)} replace />;
}
