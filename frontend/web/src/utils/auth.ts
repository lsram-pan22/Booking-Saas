
import type {
  AuthSession,
  SuperAdminUser,
} from "../types/auth.type";

const TOKEN_KEY = "token";
const USER_KEY = "user";
const EXPIRES_KEY = "expires_at";

// Save Login Session
export function saveAuthSession(data: AuthSession) {
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
export function isAuthenticated(): boolean {
  return Boolean(getAccessToken());
}

// Clear Session
export function clearAuthSession() {
  sessionStorage.removeItem(TOKEN_KEY);
  sessionStorage.removeItem(USER_KEY);
  sessionStorage.removeItem(EXPIRES_KEY);

  // ล้าง Key จากเวอร์ชันก่อนหน้าด้วย
  sessionStorage.clear();
}

// Logout
export function logout() {
  clearAuthSession();
}
