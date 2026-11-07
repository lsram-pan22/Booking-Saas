
import type {
  LoginRequest,
  LoginResponse,
} from "../types/auth.type";

const API_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080";

export async function loginSuperAdmin(
  body: LoginRequest
): Promise<LoginResponse> {
  const response = await fetch(
    `${API_URL}/api/super-admin/login`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    }
  );

  const result = await response.json();

  if (!response.ok || !result.success) {
    throw new Error(result.message || "Login failed");
  }

  if (!result.data?.access_token) {
    throw new Error("Access token missing");
  }

  return result as LoginResponse;
}
