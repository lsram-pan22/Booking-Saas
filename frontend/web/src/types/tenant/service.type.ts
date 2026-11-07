export interface ShopService {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  duration_minutes: number;
  price_satang: number;
  status: "active" | "inactive";
}

export type ServiceRequest = Omit<ShopService, "id" | "tenant_id">;
export interface ServiceForm extends Omit<ServiceRequest, "price_satang" | "status"> {
  price_baht: number;
  active: boolean;
}
export interface ServiceList {
  shop_name: string;
  shop_status: string;
  role: "owner" | "staff";
  items: ShopService[];
}
export interface ServiceDrawerProps {
  shopID: string;
  open: boolean;
  service?: ShopService;
  onClose: () => void;
  onSaved: (service: ShopService) => void;
}

export interface ShopServicesProps {
  shopID: string;
}
