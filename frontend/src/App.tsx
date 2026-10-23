import React from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { Toaster } from "sonner";
import { AuthProvider, useAuth } from "@/contexts/AuthContext";
import { CreatorLayout } from "@/components/layout/CreatorLayout";
import { LoginPage } from "@/pages/auth/LoginPage";
import { CreatorOnboardingPage } from "@/pages/onboarding/CreatorOnboardingPage";
import { DashboardPage } from "@/pages/dashboard/DashboardPage";
import { EbookListPage } from "@/pages/ebooks/EbookListPage";
import { FileLibraryPage } from "@/pages/files/FileLibraryPage";
import { SalesListPage } from "@/pages/sales/SalesListPage";
import { ClientListPage } from "@/pages/clients/ClientListPage";
import { SettingsPage } from "@/pages/settings/SettingsPage";
import { StripeConnectPage } from "@/pages/stripe/StripeConnectPage";
import { SalesPage } from "@/pages/public/SalesPage";
import { CheckoutPage } from "@/pages/public/CheckoutPage";
import { PurchaseSuccessPage } from "@/pages/public/PurchaseSuccessPage";
import { DownloadPage } from "@/pages/public/DownloadPage";

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Carregando sessão...
      </div>
    );
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function RequireCreator({ children }: { children: React.ReactNode }) {
  const { user, onboardingRequired, isLoading } = useAuth();
  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
        Carregando painel do criador...
      </div>
    );
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  if (onboardingRequired) {
    return <Navigate to="/onboarding" replace />;
  }
  return <>{children}</>;
}

export function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Toaster richColors position="top-right" />
        <Routes>
          {/* Public Routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/sales/:id" element={<SalesPage />} />
          <Route path="/checkout/:id" element={<CheckoutPage />} />
          <Route path="/purchase/success" element={<PurchaseSuccessPage />} />
          <Route path="/purchase/download/:hashId" element={<DownloadPage />} />

          {/* Step 2 Creator Onboarding */}
          <Route
            path="/onboarding"
            element={
              <RequireAuth>
                <CreatorOnboardingPage />
              </RequireAuth>
            }
          />

          {/* Authenticated Creator Area */}
          <Route
            element={
              <RequireCreator>
                <CreatorLayout />
              </RequireCreator>
            }
          >
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/ebooks" element={<EbookListPage />} />
            <Route
              path="/ebook/preview/:id"
              element={<SalesPage preview={true} />}
            />
            <Route path="/files" element={<FileLibraryPage />} />
            <Route path="/sales-history" element={<SalesListPage />} />
            <Route path="/clients" element={<ClientListPage />} />
            <Route path="/stripe-connect" element={<StripeConnectPage />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>

          {/* Default Redirect */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}
