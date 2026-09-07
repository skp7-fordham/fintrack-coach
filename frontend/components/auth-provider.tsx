"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import {
  clearSession,
  getStoredUser,
  getToken,
  setSession,
  type StoredUser,
} from "@/lib/auth";

type AuthContextValue = {
  user: StoredUser | null;
  ready: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<StoredUser | null>(null);
  const [ready, setReady] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const timer = window.setTimeout(() => {
      const token = getToken();
      const stored = getStoredUser();
      if (token && stored) {
        setUser(stored);
      } else {
        clearSession();
        setUser(null);
      }
      setReady(true);
    }, 0);
    return () => window.clearTimeout(timer);
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await api.login(email, password);
    const nextUser = { id: res.data.user.id, email: res.data.user.email };
    setSession(res.data.access_token, nextUser);
    setUser(nextUser);
    router.replace("/dashboard");
  }, [router]);

  const register = useCallback(async (email: string, password: string) => {
    const res = await api.register(email, password);
    const nextUser = { id: res.data.user.id, email: res.data.user.email };
    setSession(res.data.access_token, nextUser);
    setUser(nextUser);
    router.replace("/dashboard");
  }, [router]);

  const logout = useCallback(() => {
    clearSession();
    setUser(null);
    router.replace("/login");
  }, [router]);

  const value = useMemo(
    () => ({ user, ready, login, register, logout }),
    [user, ready, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}

export function getErrorMessage(error: unknown, fallback = "Something went wrong") {
  if (error instanceof ApiError) return error.message;
  if (error instanceof Error) return error.message;
  return fallback;
}
