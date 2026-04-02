const TOKEN_KEY = 'coahgpt_token';
const USER_KEY = 'coahgpt_user';

const API_BASE = '';

export interface User {
  readonly email: string;
  readonly name: string;
  readonly verified: boolean;
}

function isBrowser(): boolean {
  return typeof window !== 'undefined';
}

export function getToken(): string | null {
  if (!isBrowser()) return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  if (isBrowser()) {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function clearToken(): void {
  if (isBrowser()) {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
}

function cacheUser(user: User): void {
  if (isBrowser()) {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

export function getUser(): User | null {
  if (!isBrowser()) return null;
  try {
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object') return null;
    return {
      email: String(parsed.email ?? ''),
      name: String(parsed.name ?? ''),
      verified: Boolean(parsed.verified),
    };
  } catch {
    return null;
  }
}

export function isLoggedIn(): boolean {
  return getToken() !== null;
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<{ ok: boolean; status: number; data: T }> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });
  const data = (await response.json()) as T;
  return { ok: response.ok, status: response.status, data };
}

function authHeaders(): Record<string, string> {
  const token = getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

export async function signup(
  email: string,
  name: string,
  password: string,
): Promise<{ ok: boolean; error?: string }> {
  try {
    const { ok, data } = await apiFetch<{ ok?: boolean; error?: string }>(
      '/api/auth/signup',
      {
        method: 'POST',
        body: JSON.stringify({ email, name, password }),
      },
    );
    if (!ok) {
      return { ok: false, error: (data as { error?: string }).error ?? 'signup failed' };
    }
    return { ok: true };
  } catch {
    return { ok: false, error: 'network error, try again' };
  }
}

export async function login(
  email: string,
  password: string,
): Promise<{ ok: boolean; token?: string; user?: User; error?: string; status?: number }> {
  try {
    const { ok, status, data } = await apiFetch<{
      ok?: boolean;
      token?: string;
      user?: { email: string; name: string; verified: boolean };
      error?: string;
    }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    if (!ok) {
      return {
        ok: false,
        status,
        error: (data as { error?: string }).error ?? 'login failed',
      };
    }
    const token = data.token ?? '';
    const user: User = {
      email: data.user?.email ?? email,
      name: data.user?.name ?? '',
      verified: data.user?.verified ?? false,
    };
    setToken(token);
    cacheUser(user);
    return { ok: true, token, user };
  } catch {
    return { ok: false, error: 'network error, try again' };
  }
}

export async function verifyEmail(
  email: string,
  code: string,
): Promise<{ ok: boolean; error?: string }> {
  try {
    const { ok, data } = await apiFetch<{ ok?: boolean; error?: string }>(
      '/api/auth/verify',
      {
        method: 'POST',
        body: JSON.stringify({ email, code }),
      },
    );
    if (!ok) {
      return { ok: false, error: (data as { error?: string }).error ?? 'verification failed' };
    }
    return { ok: true };
  } catch {
    return { ok: false, error: 'network error, try again' };
  }
}

export async function resendVerification(
  email: string,
): Promise<{ ok: boolean; error?: string }> {
  try {
    const { ok, data } = await apiFetch<{ ok?: boolean; error?: string }>(
      '/api/auth/resend',
      {
        method: 'POST',
        body: JSON.stringify({ email }),
      },
    );
    if (!ok) {
      return { ok: false, error: (data as { error?: string }).error ?? 'resend failed' };
    }
    return { ok: true };
  } catch {
    return { ok: false, error: 'network error, try again' };
  }
}

export async function getMe(): Promise<User | null> {
  const token = getToken();
  if (!token) return null;
  try {
    const { ok, data } = await apiFetch<{
      user?: { email: string; name: string; verified: boolean };
    }>('/api/auth/me', {
      method: 'GET',
      headers: authHeaders(),
    });
    if (!ok) {
      clearToken();
      return null;
    }
    if (!data.user) {
      clearToken();
      return null;
    }
    const user: User = {
      email: data.user.email,
      name: data.user.name,
      verified: data.user.verified,
    };
    cacheUser(user);
    return user;
  } catch {
    return null;
  }
}

export async function logout(): Promise<void> {
  const token = getToken();
  if (token) {
    try {
      await apiFetch('/api/auth/logout', {
        method: 'POST',
        headers: authHeaders(),
      });
    } catch {
      // logout is best-effort
    }
  }
  clearToken();
}
