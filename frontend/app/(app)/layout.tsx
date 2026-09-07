"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth-provider";
import { AppSidebar } from "@/components/app-shell";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const { user, ready } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!ready) return;
    if (!user) router.replace("/login");
  }, [ready, user, router]);

  if (!ready || !user) {
    return (
      <div className="flex min-h-screen items-center justify-center text-sm text-muted">
        Checking session…
      </div>
    );
  }

  return (
    <div className="min-h-screen lg:flex">
      <AppSidebar />
      <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">{children}</main>
    </div>
  );
}
