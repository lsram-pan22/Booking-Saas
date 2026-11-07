import type { TenantShop } from "./shop.type";

export interface TenantRegistration {
  full_name: string;
  email: string;
  phone: string;
  password: string;
}

export interface TenantLoginRequest {
  username: string;
  password: string;
}

export interface TenantProfile {
  workspace_role: "owner" | "staff";
  owned_shop_count: number;
	 subscription: TenantSubscription | null;
  user: TenantUser;
  tenants: TenantShop[];
}

export interface TenantSession extends TenantProfile {
  access_token: string;
  token_type: string;
  expires_in: number;
}

export interface TenantRegistrationResult {
  user_id: string;
  subscription: TenantSubscription;
}

export interface TenantSubscription {
  shop_limit: number;
  id: string;
  package_id: string;
  package_code: string;
  package_name: string;
  status: string;
  monthly_price_satang: number;
  user_limit: number;
  staff_management_level: "basic" | "full";
  report_level: "basic" | "advanced";
  export_excel: boolean;
  api_webhook: boolean;
  starts_at: string;
  ends_at: string | null;
}

export interface TenantUser {
  role: "owner" | "staff";
  system_role: "user";
  id: string;
  email: string;
  full_name: string;
  phone: string;
}

export type TenantSettingsForm = Pick<TenantUser, "full_name" | "phone">;
export type TenantSettingsResult = Pick<TenantUser, "id" | "email" | "full_name" | "phone">;

export interface RegistrationForm {
  full_name: string;
  phone: string;
  email: string;
  password: string;
  confirm_password: string;
}

