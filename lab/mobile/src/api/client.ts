import { clearTokens, loadTokens, saveTokens } from '../auth/storage';

const defaultBase = 'http://127.0.0.1:8080/api';

export function apiBase(): string {
  const fromEnv = process.env.EXPO_PUBLIC_API_URL;
  return (fromEnv && fromEnv.replace(/\/$/, '')) || defaultBase;
}

export function mediaUrl(path: string): string {
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  if (path.startsWith('/api/')) {
    return `${apiBase()}${path.slice(4)}`;
  }
  if (path.startsWith('/')) {
    return `${apiBase()}${path}`;
  }
  return `${apiBase()}/${path}`;
}

export type ApiError = Error & { status?: number };

function asApiError(message: string, status?: number): ApiError {
  const err = new Error(message) as ApiError;
  err.status = status;
  return err;
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string };
    if (body.error) {
      return body.error;
    }
  } catch {
    // ignore
  }
  return res.statusText || 'request failed';
}

export type AuthSession = {
  token: string;
  refresh_token: string;
  user: AuthUser;
};

export type AuthUser = {
  id: string;
  name: string;
  unit_number: string;
  is_admin: boolean;
  email?: string;
  status?: string;
};

let accessToken: string | null = null;
let refreshToken: string | null = null;
let refreshInFlight: Promise<boolean> | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export async function hydrateSession(): Promise<AuthUser | null> {
  const stored = await loadTokens();
  accessToken = stored.access;
  refreshToken = stored.refresh;
  if (!accessToken) {
    return null;
  }
  try {
    return await getMe();
  } catch (err) {
    const status = (err as ApiError).status;
    if (status === 401 && (await tryRefresh())) {
      return await getMe();
    }
    await signOutLocal();
    return null;
  }
}

export async function signIn(email: string, password: string): Promise<AuthUser> {
  const session = await request<AuthSession>('/auth/login', {
    method: 'POST',
    auth: false,
    body: { email, password },
  });
  await acceptSession(session);
  return session.user;
}

export async function signOut(): Promise<void> {
  const currentRefresh = refreshToken;
  try {
    if (currentRefresh) {
      await request('/auth/logout', {
        method: 'POST',
        auth: false,
        body: { refresh_token: currentRefresh },
      });
    }
  } catch {
    // still clear local session
  }
  await signOutLocal();
}

export async function completeRegistration(input: {
  token: string;
  passcode: string;
  password: string;
  name: string;
}): Promise<AuthUser> {
  const session = await request<AuthSession>('/auth/complete-registration', {
    method: 'POST',
    auth: false,
    body: input,
  });
  await acceptSession(session);
  return session.user;
}

export async function acceptSession(session: AuthSession): Promise<void> {
  accessToken = session.token;
  refreshToken = session.refresh_token;
  await saveTokens(session.token, session.refresh_token);
}

async function signOutLocal(): Promise<void> {
  accessToken = null;
  refreshToken = null;
  await clearTokens();
}

export async function getMe(): Promise<AuthUser> {
  return request<AuthUser>('/auth/me', { method: 'GET' });
}

type RequestOpts = {
  method: string;
  body?: unknown;
  form?: FormData;
  auth?: boolean;
  retry?: boolean;
};

export async function request<T>(path: string, opts: RequestOpts): Promise<T> {
  const headers: Record<string, string> = {};
  if (opts.body !== undefined && !opts.form) {
    headers['Content-Type'] = 'application/json';
  }
  const useAuth = opts.auth !== false;
  if (useAuth && accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }

  let body: BodyInit | undefined;
  if (opts.form) {
    body = opts.form;
  } else if (opts.body !== undefined) {
    body = JSON.stringify(opts.body);
  }

  const res = await fetch(`${apiBase()}${path}`, {
    method: opts.method,
    headers,
    body,
  });

  if (res.status === 401 && useAuth && opts.retry !== false) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      return request<T>(path, { ...opts, retry: false });
    }
    await signOutLocal();
    throw asApiError('session expired', 401);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  if (!res.ok) {
    throw asApiError(await readError(res), res.status);
  }

  return (await res.json()) as T;
}

async function tryRefresh(): Promise<boolean> {
  if (!refreshToken) {
    return false;
  }
  if (!refreshInFlight) {
    refreshInFlight = doRefresh().finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

async function doRefresh(): Promise<boolean> {
  const current = refreshToken;
  if (!current) {
    return false;
  }
  try {
    const session = await request<AuthSession>('/auth/refresh', {
      method: 'POST',
      auth: false,
      body: { refresh_token: current },
    });
    await acceptSession(session);
    return true;
  } catch {
    return false;
  }
}
