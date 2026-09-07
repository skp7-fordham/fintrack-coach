"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  ArrowLeftRight,
  Wallet,
  Tags,
  Upload,
  Bot,
  LogOut,
  Menu,
  X,
} from "lucide-react";
import { useState } from "react";
import { useAuth } from "@/components/auth-provider";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/transactions", label: "Transactions", icon: ArrowLeftRight },
  { href: "/accounts", label: "Accounts", icon: Wallet },
  { href: "/categories", label: "Categories", icon: Tags },
  { href: "/imports", label: "Imports", icon: Upload },
  { href: "/coach", label: "AI Coach", icon: Bot },
];

function NavLinks({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();

  return (
    <nav className="flex flex-col gap-1" aria-label="Primary">
      {navItems.map((item) => {
        const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
        const Icon = item.icon;
        return (
          <Link
            key={item.href}
            href={item.href}
            onClick={onNavigate}
            className={`flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
              active
                ? "bg-teal-700/40 text-white"
                : "text-sidebar-muted hover:bg-white/5 hover:text-white"
            }`}
          >
            <Icon className="h-4 w-4 shrink-0" aria-hidden />
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}

export function AppSidebar() {
  const { user, logout } = useAuth();
  const [open, setOpen] = useState(false);

  return (
    <>
      <div className="sticky top-0 z-30 flex items-center justify-between border-b border-border bg-card px-4 py-3 lg:hidden">
        <div className="flex items-center gap-2">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-sm font-bold text-white">
            F
          </span>
          <span className="font-semibold text-foreground">FinTrack Coach</span>
        </div>
        <button
          type="button"
          aria-label={open ? "Close menu" : "Open menu"}
          className="rounded-lg border border-border p-2"
          onClick={() => setOpen((v) => !v)}
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {open ? (
        <div className="fixed inset-0 z-40 bg-black/40 lg:hidden" onClick={() => setOpen(false)}>
          <aside
            className="flex h-full w-72 flex-col bg-sidebar p-4 text-sidebar-text"
            onClick={(e) => e.stopPropagation()}
          >
            <Brand />
            <div className="mt-8 flex-1">
              <NavLinks onNavigate={() => setOpen(false)} />
            </div>
            <UserFooter email={user?.email} onLogout={logout} />
          </aside>
        </div>
      ) : null}

      <aside className="hidden w-64 shrink-0 flex-col bg-sidebar p-4 text-sidebar-text lg:flex">
        <Brand />
        <div className="mt-8 flex-1">
          <NavLinks />
        </div>
        <UserFooter email={user?.email} onLogout={logout} />
      </aside>
    </>
  );
}

function Brand() {
  return (
    <div className="flex items-center gap-3 px-2">
      <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-teal-500 text-sm font-bold text-white">
        F
      </span>
      <div>
        <p className="text-sm font-semibold text-white">FinTrack Coach</p>
        <p className="text-xs text-sidebar-muted">Personal finance</p>
      </div>
    </div>
  );
}

function UserFooter({
  email,
  onLogout,
}: {
  email?: string;
  onLogout: () => void;
}) {
  return (
    <div className="border-t border-white/10 pt-4">
      <p className="truncate px-2 text-xs text-sidebar-muted">{email ?? "Signed in"}</p>
      <button
        type="button"
        onClick={onLogout}
        className="mt-2 flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-sidebar-muted hover:bg-white/5 hover:text-white"
      >
        <LogOut className="h-4 w-4" aria-hidden />
        Log out
      </button>
    </div>
  );
}

export function AppHeader({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <header className="mb-6">
      <h1 className="text-2xl font-semibold tracking-tight text-foreground">{title}</h1>
      {subtitle ? <p className="mt-1 text-sm text-muted">{subtitle}</p> : null}
    </header>
  );
}
