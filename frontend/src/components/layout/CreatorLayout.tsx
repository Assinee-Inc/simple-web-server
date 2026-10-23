import React from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  BookOpen,
  CreditCard,
  FileText,
  LayoutDashboard,
  LogOut,
  Settings,
  ShieldCheck,
  ShoppingBag,
  Users,
} from "lucide-react";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/ebooks", label: "E-books", icon: BookOpen },
  { to: "/files", label: "Arquivos (S3)", icon: FileText },
  { to: "/sales-history", label: "Vendas & Transações", icon: ShoppingBag },
  { to: "/clients", label: "Clientes", icon: Users },
  { to: "/stripe-connect", label: "Recebimentos Stripe", icon: CreditCard },
  { to: "/settings", label: "Configurações", icon: Settings },
];

export function CreatorLayout() {
  const { user, creator, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate("/login");
  };

  return (
    <div className="flex min-h-screen bg-muted/20">
      {/* Sidebar */}
      <aside className="hidden w-64 flex-col border-r bg-card md:flex">
        <div className="flex h-16 items-center gap-2 border-b px-6">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <Link to="/dashboard" className="text-lg font-bold tracking-tight">
            Docffy
          </Link>
          <Badge variant="secondary" className="ml-auto text-[10px]">
            API v1
          </Badge>
        </div>

        <nav className="flex-1 space-y-1 p-4" aria-label="Menu do Criador">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-primary text-primary-foreground shadow-xs"
                      : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                  )
                }
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            );
          })}
        </nav>

        <div className="border-t p-4">
          <div className="mb-3 rounded-lg bg-muted/60 p-3 text-xs">
            <p className="font-semibold text-foreground truncate">
              {creator?.social_name || creator?.name || user?.username}
            </p>
            <p className="text-muted-foreground truncate">{user?.email}</p>
            {user?.days_left !== undefined && (
              <div className="mt-2 flex items-center justify-between">
                <span className="text-muted-foreground">Assinatura:</span>
                <Badge
                  variant={
                    user.subscription_status === "active" ||
                    user.subscription_status === "trialing"
                      ? "success"
                      : "warning"
                  }
                >
                  {user.subscription_status || "trial"} ({user.days_left}d)
                </Badge>
              </div>
            )}
          </div>
          <Button
            variant="outline"
            size="sm"
            className="w-full justify-start text-muted-foreground"
            onClick={handleLogout}
          >
            <LogOut className="h-4 w-4" />
            Encerrar sessão
          </Button>
        </div>
      </aside>

      {/* Main Content */}
      <div className="flex flex-1 flex-col">
        <header className="flex h-16 items-center justify-between border-b bg-card px-6">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium text-muted-foreground">
              Painel do Criador
            </span>
            {creator && !creator.onboarding_completed && (
              <Link to="/stripe-connect">
                <Badge variant="warning" className="cursor-pointer">
                  Stripe Connect Pendente — Clique para ativar recebimentos
                </Badge>
              </Link>
            )}
          </div>
          <div className="flex items-center gap-3 md:hidden">
            <Button variant="outline" size="sm" onClick={handleLogout}>
              <LogOut className="h-4 w-4" />
              Sair
            </Button>
          </div>
        </header>

        <main className="flex-1 p-6 md:p-8 max-w-7xl w-full mx-auto">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
