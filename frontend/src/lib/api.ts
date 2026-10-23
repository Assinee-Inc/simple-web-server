const TOKEN_STORAGE_KEY = "docffy_access_token";

export interface ApiErrorPayload {
  success: false;
  code?: string;
  error: string;
  details?: Record<string, string>;
  onboarding_status?: unknown;
}

export class ApiError extends Error {
  status: number;
  code?: string;
  details?: Record<string, string>;

  constructor(
    status: number,
    message: string,
    code?: string,
    details?: Record<string, string>
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY);
}

export function setAccessToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token);
}

export function clearAccessToken(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY);
}

export async function apiRequest<T = any>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getAccessToken();
  const headers = new Headers(options.headers || {});

  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  if (
    options.body &&
    !(options.body instanceof FormData) &&
    !headers.has("Content-Type")
  ) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, {
    ...options,
    headers,
  });

  const contentType = response.headers.get("content-type") || "";
  const isJson = contentType.includes("application/json");
  const payload = isJson ? await response.json() : await response.text();

  if (!response.ok) {
    const errObj = (isJson ? payload : {}) as Partial<ApiErrorPayload>;
    const message =
      errObj.error || `Erro na requisição (${response.status})`;
    const code = errObj.code;

    if (response.status === 401 && !path.includes("/api/auth/google")) {
      clearAccessToken();
    }

    throw new ApiError(response.status, message, code, errObj.details);
  }

  // Unwrap data envelope if backend uses {"success": true, "data": ...} alongside top-level fields
  if (
    payload &&
    typeof payload === "object" &&
    "data" in payload &&
    Object.keys(payload).length <= 3 &&
    payload.data &&
    typeof payload.data === "object" &&
    !Array.isArray(payload.data)
  ) {
    return { ...payload, ...payload.data } as T;
  }

  return payload as T;
}
