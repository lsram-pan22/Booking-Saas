
export interface SuperAdminUser {
  id: number;
  username?: string;
  email?: string;
  full_name?: string;
  role?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface AuthSession {
  access_token: string;
  token_type: string;
  expires_in?: number;
  expires_at?: number;
  user?: SuperAdminUser;
}

export interface LoginResponse {
  success: boolean;
  message: string;
  data: AuthSession;
}
