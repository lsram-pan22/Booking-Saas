
export interface LoginForm {
  email: string;
  password: string;
  remember: boolean;
}

export interface RegisterForm {
  full_name: string;
  email: string;
  phone: string;
  password: string;
  confirm_password: string;
  accept_terms: boolean;
}
