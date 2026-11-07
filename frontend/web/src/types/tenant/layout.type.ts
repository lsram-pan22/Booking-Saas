import type { TenantProfile, TenantSubscription, TenantUser, TenantSettingsResult } from "./account.type";
import type { TenantShop } from "./shop.type";

export interface TenantSidebarProps {
  profile?: TenantProfile;
  collapsed?: boolean;
  onNavigate?: () => void;
  onToggle?: () => void;
}

export interface TenantHeaderProps {
  collapsed: boolean;
  onToggle: () => void;
  showToggle?: boolean;
  subscription?: TenantSubscription | null;
  user?: TenantUser;
}

export interface TenantLayoutContext {
  profile?: TenantProfile;
  error: string;
  updateUser: (user: TenantSettingsResult) => void;
  addShop: (shop: TenantShop) => void;
  updateShop: (shop: TenantShop) => void;
}
