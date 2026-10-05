import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { authApi, usersApi } from '@/api/endpoints';
import type { Session, User } from '@/api/types';
import { loadTokens, saveTokens } from './storage';

type Status = 'loading' | 'signed-out' | 'signed-in';

type AuthValue = {
  status: Status;
  user: User | null;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
  /** Accept a session returned by any auth endpoint (invite completion, password change). */
  acceptSession: (s: Session) => void;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const qc = useQueryClient();
  const [status, setStatus] = useState<Status>('loading');
  const [user, setUser] = useState<User | null>(null);

  const clear = useCallback(() => {
    api.setTokens(null);
    saveTokens(null);
    setUser(null);
    setStatus('signed-out');
    qc.clear();
  }, [qc]);

  // Wire the client's token lifecycle to storage and sign-out routing.
  useEffect(() => {
    api.onTokensChanged = saveTokens;
    api.onSignedOut = clear;
    return () => {
      api.onTokensChanged = null;
      api.onSignedOut = null;
    };
  }, [clear]);

  // Restore on boot.
  useEffect(() => {
    const stored = loadTokens();
    if (!stored) {
      setStatus('signed-out');
      return;
    }
    api.setTokens(stored);
    usersApi
      .me()
      .then((u) => {
        setUser(u);
        setStatus('signed-in');
      })
      .catch(clear);
  }, [clear]);

  const acceptSession = useCallback((s: Session) => {
    const t = { access: s.access_token, refresh: s.refresh_token };
    api.setTokens(t);
    saveTokens(t);
    setUser(s.user);
    setStatus('signed-in');
  }, []);

  const signIn = useCallback(
    async (email: string, password: string) => {
      acceptSession(await authApi.login(email, password));
    },
    [acceptSession],
  );

  const signOut = useCallback(async () => {
    const t = api.getTokens();
    if (t) await authApi.logout(t.refresh).catch(() => undefined);
    clear();
  }, [clear]);

  const refreshUser = useCallback(async () => {
    setUser(await usersApi.me());
  }, []);

  const value = useMemo(
    () => ({ status, user, signIn, signOut, acceptSession, refreshUser }),
    [status, user, signIn, signOut, acceptSession, refreshUser],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthValue {
  const v = useContext(AuthContext);
  if (!v) throw new Error('useAuth must be used within AuthProvider');
  return v;
}
