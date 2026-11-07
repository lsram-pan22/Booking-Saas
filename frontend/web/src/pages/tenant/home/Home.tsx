import { Alert, Spin } from "antd";
import { Navigate, useOutletContext } from "react-router-dom";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";
import { getTenantLandingPath } from "../../../utils/tenant-access";

export default function Home() {
  const { profile, error } = useOutletContext<TenantLayoutContext>();
  if (error) return <Alert type="error" showIcon title={error} />;
  return profile ? <Navigate to={getTenantLandingPath(profile)} replace /> : <Spin />;
}
