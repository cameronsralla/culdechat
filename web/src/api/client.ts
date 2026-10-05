import type { ApiErrorBody } from './types';

const BASE = (import.meta.env.VITE_API_URL as string | undefined)?.replace(/\/$/, '') ?? '';

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}

type Tokens = { access: string; refresh: string };

/**
 * Token store + fetch wrapper. Bearer access token on every call; on 401 it
 * refreshes once (single-flight) and retries. If refresh fails the session
 * is cleared and `onSignedOut` fires so the app can route to /login.
 */
class ApiClient {
  private tokens: Tokens | null = null;
  private refreshing: Promise<boolean> | null = null;
  onSignedOut: (() => void) | null = null;

  setTokens(t: Tokens | null) {
    this.tokens = t;
  }
  getTokens() {
    return this.tokens;
  }

  async request<T>(method: string, path: string, body?: unknown, opts: { auth?: boolean; retry?: boolean } = {}): Promise<T> {
    const { auth = true, retry = true } = opts;
    const headers: Record<string, string> = {};
    if (body !== undefined) headers['Content-Type'] = 'application/json';
    if (auth && this.tokens) headers.Authorization = `Bearer ${this.tokens.access}`;

    const res = await fetch(`${BASE}/api${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    if (res.status === 401 && auth && retry && this.tokens) {
      if (await this.refresh()) return this.request<T>(method, path, body, { auth, retry: false });
      throw new ApiError(401, 'unauthorized', 'Please sign in again.');
    }
    if (res.status === 204) return undefined as T;

    const text = await res.text();
    const data = text ? (JSON.parse(text) as unknown) : null;
    if (!res.ok) {
      const err = (data as ApiErrorBody | null)?.error;
      throw new ApiError(res.status, err?.code ?? 'error', err?.message ?? res.statusText);
    }
    return data as T;
  }

  get<T>(path: string, opts?: { auth?: boolean }) {
    return this.request<T>('GET', path, undefined, opts);
  }
  post<T>(path: string, body?: unknown, opts?: { auth?: boolean }) {
    return this.request<T>('POST', path, body ?? {}, opts);
  }
  put<T>(path: string, body: unknown) {
    return this.request<T>('PUT', path, body);
  }
  patch<T>(path: string, body: unknown) {
    return this.request<T>('PATCH', path, body);
  }
  delete<T>(path: string) {
    return this.request<T>('DELETE', path);
  }

  /** Single-flight refresh. Resolves true if new tokens were obtained. */
  private refresh(): Promise<boolean> {
    if (!this.refreshing) {
      this.refreshing = (async () => {
        const refresh = this.tokens?.refresh;
        if (!refresh) return false;
        try {
          const res = await fetch(`${BASE}/api/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refresh }),
          });
          if (!res.ok) throw new Error('refresh failed');
          const s = (await res.json()) as { access_token: string; refresh_token: string };
          this.tokens = { access: s.access_token, refresh: s.refresh_token };
          this.onTokensChanged?.(this.tokens);
          return true;
        } catch {
          this.tokens = null;
          this.onTokensChanged?.(null);
          this.onSignedOut?.();
          return false;
        } finally {
          this.refreshing = null;
        }
      })();
    }
    return this.refreshing;
  }

  onTokensChanged: ((t: Tokens | null) => void) | null = null;
}

export const api = new ApiClient();
