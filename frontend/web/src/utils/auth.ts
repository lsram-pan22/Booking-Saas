
import type {
  AuthSession,
  SuperAdminUser,
} from "../types/shared/auth.type";

const TOKEN_KEY = "token";
const USER_KEY = "user";
const EXPIRES_KEY = "expires_at";

// Save Login Session
export function saveAuthSession(data: AuthSession) {
  sessionStorage.removeItem("tenant_access_token");
  sessionStorage.removeItem("tenant_expires_at");
  sessionStorage.setItem(TOKEN_KEY, data.access_token);

  if (data.user) {
    sessionStorage.setItem(
      USER_KEY,
      JSON.stringify(data.user)
    );
  } else {
    sessionStorage.removeItem(USER_KEY);
  }

  if (data.expires_in !== undefined) {
    const expiresAt = Date.now() + data.expires_in * 1000;

    sessionStorage.setItem(
      EXPIRES_KEY,
      expiresAt.toString()
    );
  } else {
    sessionStorage.removeItem(EXPIRES_KEY);
  }
}

// Get Token
export function getAccessToken(): string | null {
  const token = sessionStorage.getItem(TOKEN_KEY);
  const expires = sessionStorage.getItem(EXPIRES_KEY);

  if (expires && Date.now() >= Number(expires)) {
    clearAuthSession();
    return null;
  }

  return token;
}

// Get User
export function updateCurrentUser(user: SuperAdminUser) {
  sessionStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function getCurrentUser(): SuperAdminUser | null {
  const stored = sessionStorage.getItem(USER_KEY);

  if (!stored) return null;

  try {
    return JSON.parse(stored) as SuperAdminUser;
  } catch {
    sessionStorage.removeItem(USER_KEY);
    return null;
  }
}

// Check Login
export function isAuthenticated(role?: string): boolean {
  return Boolean(getAccessToken()) && (!role || getCurrentUser()?.role === role);
}

export function isTenantAuthenticated(): boolean {
  if (!getAccessToken()) return false;
  const user = getCurrentUser();
  // Accept legacy sessions until /me refreshes their tenant role.
  return Boolean(user && (user.role === "user" || (user.system_role === "user" && (user.role === "owner" || user.role === "staff"))));
}

// Clear Session
export function clearAuthSession() {
  sessionStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(USER_KEY);
  sessionStorage.removeItem(EXPIRES_KEY);

  sessionStorage.removeItem("tenant_access_token");
  sessionStorage.removeItem("tenant_expires_at");
}

// Logout
export function logout() {
  clearAuthSession();
}
