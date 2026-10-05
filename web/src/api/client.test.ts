import { api, ApiError } from './client';

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

describe('api client', () => {
  beforeEach(() => {
    api.setTokens({ access: 'old', refresh: 'r1' });
    api.onSignedOut = null;
    api.onTokensChanged = null;
  });

  it('refreshes once on 401 and retries with the new token', async () => {
    const calls: string[] = [];
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
      const url = String(input);
      const auth = (init?.headers as Record<string, string> | undefined)?.Authorization ?? '';
      calls.push(`${url} ${auth}`);
      if (url.endsWith('/api/auth/refresh')) return jsonResponse(200, { access_token: 'new', refresh_token: 'r2' });
      if (auth === 'Bearer old') return jsonResponse(401, { error: { code: 'unauthorized', message: 'x' } });
      return jsonResponse(200, { ok: true });
    });

    const res = await api.get<{ ok: boolean }>('/me');
    expect(res.ok).toBe(true);
    expect(calls).toEqual(['/api/me Bearer old', '/api/auth/refresh ', '/api/me Bearer new']);
    expect(api.getTokens()).toEqual({ access: 'new', refresh: 'r2' });
    vi.restoreAllMocks();
  });

  it('signs out when refresh fails', async () => {
    const signedOut = vi.fn();
    api.onSignedOut = signedOut;
    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
      const url = String(input);
      if (url.endsWith('/api/auth/refresh')) return jsonResponse(401, { error: { code: 'unauthorized', message: 'nope' } });
      return jsonResponse(401, { error: { code: 'unauthorized', message: 'x' } });
    });

    await expect(api.get('/me')).rejects.toBeInstanceOf(ApiError);
    expect(signedOut).toHaveBeenCalled();
    expect(api.getTokens()).toBeNull();
    vi.restoreAllMocks();
  });

  it('surfaces server error messages', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse(409, { error: { code: 'conflict', message: 'already exists' } }));
    await expect(api.post('/x', {})).rejects.toMatchObject({ status: 409, code: 'conflict', message: 'already exists' });
    vi.restoreAllMocks();
  });
});
