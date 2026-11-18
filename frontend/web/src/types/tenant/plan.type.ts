import type { TenantSubscription } from "./account.type";
export interface PlanMember { id: string; name: string; role: "owner" | "staff"; enabled: boolean }
export interface PlanShop { id: string; name: string; status: string; enabled: boolean; members: PlanMember[] }
export interface PlanOverview { history: TenantSubscription[]; shops: PlanShop[] }
export interface PlanSelection { shop_ids: string[]; member_ids: string[] }
