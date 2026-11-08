import { Alert, Spin } from "antd";
import { Navigate, useOutletContext } from "react-router-dom";
import type { TenantLayoutContext } from "../../../types/tenant/layout.type";
import { getTenantLandingPath } from "../../../utils/tenant-access";

export default function StaffAccess() {
  const { profile, error } = useOutletContext<TenantLayoutContext>();
  if (error) return <Alert type="error" showIcon title={error} />;
  if (!profile) return <Spin />;
  const path = getTenantLandingPath(profile);
  if (path !== "/tenant/staff") return <Navigate to={path} replace />;
  return <Alert type="warning" showIcon title="ยังไม่มีร้านค้าที่คุณเข้าใช้งานได้" description="กรุณาติดต่อเจ้าของร้านเพื่อตรวจสอบสิทธิ์และสถานะร้านค้า" />;
}
