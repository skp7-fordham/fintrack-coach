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
import { useEffect, useState } from "react";
import { useAuth } from "@/components/auth-provider";
import { BrandLockup } from "@/components/brand";

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
    <nav className="flex flex-col gap-1.5" aria-label="Primary">
      {navItems.map((item) => {
        const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
        const Icon = item.icon;
        return (
          <Link
            key={item.href}
            href={item.href}
            onClick={onNavigate}
            className={`group flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors ${
              active
                ? "bg-teal-500/15 text-white ring-1 ring-inset ring-teal-300/10"
                : "text-sidebar-muted hover:bg-white/[0.06] hover:text-white"
            }`}
          >
            <Icon
              className={`h-4 w-4 shrink-0 transition-colors ${
                active ? "text-teal-300" : "text-sidebar-muted group-hover:text-white"
              }`}
              strokeWidth={1.8}
              aria-hidden
            />
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

  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", closeOnEscape);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  return (
    <>
      <div className="sticky top-0 z-30 flex items-center justify-between border-b border-border bg-card px-4 py-3 lg:hidden">
        <BrandLockup compact />
        <button
          type="button"
          aria-label={open ? "Close menu" : "Open menu"}
          className="rounded-md border border-border p-2 text-muted transition-colors hover:bg-slate-50 hover:text-foreground"
          onClick={() => setOpen((v) => !v)}
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {open ? (
        <div className="fixed inset-0 z-40 bg-black/40 lg:hidden" onClick={() => setOpen(false)}>
          <aside
            className="flex h-full w-72 flex-col bg-sidebar p-5 text-sidebar-text"
            onClick={(e) => e.stopPropagation()}
          >
            <Brand />
            <div className="mt-9 flex-1">
              <NavLinks onNavigate={() => setOpen(false)} />
            </div>
            <UserFooter email={user?.email} onLogout={logout} />
          </aside>
        </div>
      ) : null}

      <aside className="hidden h-screen w-64 shrink-0 flex-col bg-sidebar p-5 text-sidebar-text lg:sticky lg:top-0 lg:flex">
        <Brand />
        <div className="mt-9 flex-1">
          <NavLinks />
        </div>
        <UserFooter email={user?.email} onLogout={logout} />
      </aside>
    </>
  );
}

function Brand() {
  return (
    <div className="px-2">
      <BrandLockup inverse />
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
        className="mt-2 flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-sidebar-muted transition-colors hover:bg-white/[0.06] hover:text-white"
      >
        <LogOut className="h-4 w-4" aria-hidden />
        Log out
      </button>
    </div>
  );
}

export function AppHeader({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <header className="mb-6 max-w-3xl">
      <h1 className="text-2xl font-semibold tracking-[-0.025em] text-foreground">{title}</h1>
      {subtitle ? <p className="mt-1.5 text-sm leading-6 text-muted">{subtitle}</p> : null}
    </header>
  );
}
