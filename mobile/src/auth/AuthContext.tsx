import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import {
  completeRegistration as apiComplete,
  getMe,
  hydrateSession,
  signIn as apiSignIn,
  signOut as apiSignOut,
  type AuthUser,
} from '../api/client';

type AuthContextValue = {
  ready: boolean;
  user: AuthUser | null;
  login: (email: string, password: string) => Promise<void>;
  completeRegister: (input: { token: string; passcode: string; password: string; name: string }) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false);
  const [user, setUser] = useState<AuthUser | null>(null);

  useEffect(() => {
    let cancelled = false;
    hydrateSession()
      .then((next) => {
        if (!cancelled) {
          setUser(next);
        }
      })
      .finally(() => {
        if (!cancelled) {
          setReady(true);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      ready,
      user,
      login: async (email, password) => {
        const next = await apiSignIn(email, password);
        try {
          setUser(await getMe());
        } catch {
          setUser(next);
        }
      },
      completeRegister: async (input) => {
        const next = await apiComplete(input);
        try {
          setUser(await getMe());
        } catch {
          setUser(next);
        }
      },
      logout: async () => {
        await apiSignOut();
        setUser(null);
      },
      refreshUser: async () => {
        setUser(await getMe());
      },
    }),
    [ready, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
