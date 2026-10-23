import React, { createContext, useContext, useEffect, useState } from "react";
import {
  apiRequest,
  clearAccessToken,
  getAccessToken,
  setAccessToken,
} from "@/lib/api";

export interface UserSummary {
  public_id: string;
  username: string;
  email: string;
  google_id?: string;
  email_verified: boolean;
  terms_accepted: boolean;
  subscription_status?: string;
  days_left?: number;
}

export interface CreatorSummary {
  public_id: string;
  name: string;
  social_name?: string;
  email: string;
  cpf?: string;
  phone?: string;
  onboarding_completed: boolean;
  charges_enabled?: boolean;
  payouts_enabled?: boolean;
  needs_stripe_onboarding?: boolean;
}

interface AuthContextValue {
  user: UserSummary | null;
  creator: CreatorSummary | null;
  onboardingRequired: boolean;
  isLoading: boolean;
  loginWithGoogle: (idToken: string) => Promise<{
    onboarding_required: boolean;
    creator?: CreatorSummary | null;
  }>;
  refreshSession: () => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserSummary | null>(null);
  const [creator, setCreator] = useState<CreatorSummary | null>(null);
  const [onboardingRequired, setOnboardingRequired] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const refreshSession = async () => {
    const token = getAccessToken();
    if (!token) {
      setUser(null);
      setCreator(null);
      setOnboardingRequired(false);
      setIsLoading(false);
      return;
    }

    try {
      const res = await apiRequest<{
        user: UserSummary;
        creator?: CreatorSummary;
        onboarding_required: boolean;
      }>("/api/auth/me");
      setUser(res.user || null);
      setCreator(res.creator || null);
      setOnboardingRequired(Boolean(res.onboarding_required));
    } catch {
      clearAccessToken();
      setUser(null);
      setCreator(null);
      setOnboardingRequired(false);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    refreshSession();
  }, []);

  const loginWithGoogle = async (idToken: string) => {
    const res = await apiRequest<{
      access_token: string;
      onboarding_required: boolean;
      user: UserSummary;
      creator?: CreatorSummary;
    }>("/api/auth/google", {
      method: "POST",
      body: JSON.stringify({ id_token: idToken }),
    });

    if (res.access_token) {
      setAccessToken(res.access_token);
    }
    setUser(res.user);
    setCreator(res.creator || null);
    setOnboardingRequired(Boolean(res.onboarding_required));
    await refreshSession();

    return {
      onboarding_required: Boolean(res.onboarding_required),
      creator: res.creator || null,
    };
  };

  const logout = async () => {
    try {
      await apiRequest("/api/auth/logout", { method: "POST" });
    } catch {
      // Ignore network errors on logout
    } finally {
      clearAccessToken();
      setUser(null);
      setCreator(null);
      setOnboardingRequired(false);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        creator,
        onboardingRequired,
        isLoading,
        loginWithGoogle,
        refreshSession,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
