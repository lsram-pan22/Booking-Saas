export interface Package {
  shop_limit: number;
  id: string;
  code: string;
  name: string;
  description: string;
  monthly_price_satang: number;
  user_limit: number;
  staff_management: "basic" | "full";
  reports: "basic" | "advanced";
  export_excel: boolean;
  api_webhook: boolean;
  active: boolean;
  sort_order: number;
}

export type PackageInput = Omit<Package, "id">;

export type PackageForm = Omit<PackageInput, "monthly_price_satang"> & { monthly_price: number };
