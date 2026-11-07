export interface TenantShop {
  id: string;
  name: string;
  status: string;
  role: string;
  business_type?: string;
  slug?: string;
}

export interface CreateShopRequest {
  name: string;
  business_type: string;
}

export interface UpdateShopRequest extends CreateShopRequest {
  status: "active" | "inactive";
}

export interface EditShopForm extends CreateShopRequest {
  active: boolean;
}

export interface ShopDetails extends TenantShop {
  business_type: string;
  slug: string;
}

export interface ShopDetailsProps {
  shopID: string;
}

export interface CreateShopDrawerProps {
  shopLimit: number;
  open: boolean;
  onClose: () => void;
  onCreated: (shop: TenantShop) => void;
}

export interface ShopMember {
  username: string;
  id: string;
  user_id: string;
  full_name: string;
  email: string;
  phone: string;
  role: string;
  status: string;
  user_status: string;
}

export interface ShopMembers {
  active_count: number;
  user_limit: number | null;
  members: ShopMember[];
}

export interface ShopMembersCardProps {
  data?: ShopMembers;
  error: string;
  canAdd: boolean;
  onAdd: () => void;
  canManage: boolean;
  updatingMemberID?: string;
  onStatusChange: (member: ShopMember, active: boolean) => void;
}

export interface MemberStatusRequest {
  status: "active" | "inactive";
}

export interface MemberStatusResult {
  member: ShopMember;
  active_count: number;
  user_limit: number | null;
}

export type AddShopMemberRequest =
  | { mode: "existing"; email: string }
  | { mode: "new"; username: string; email: string; full_name: string; phone: string; password: string };

export interface AddMemberForm {
  username?: string;
  mode: "new" | "existing";
  email: string;
  full_name?: string;
  phone?: string;
  password?: string;
  confirm_password?: string;
}

export interface AddMemberDrawerProps {
  shopID: string;
  open: boolean;
  onClose: () => void;
  onAdded: (member: ShopMember) => void;
}

export const businessTypes = [
  { value: "salon", label: "ร้านทำผม / ร้านเสริมสวย" },
  { value: "spa", label: "สปา / นวด" },
  { value: "clinic", label: "คลินิก" },
  { value: "fitness", label: "ฟิตเนส / สตูดิโอ" },
  { value: "other", label: "ธุรกิจบริการอื่น ๆ" },
];

export const shopStatusLabels: Record<string, string> = {
  active: "เปิดใช้งาน",
  inactive: "ปิดใช้งาน",
  pending: "รอเปิดใช้งาน",
  suspended: "ระงับ",
  closed: "ปิดถาวร",
};
