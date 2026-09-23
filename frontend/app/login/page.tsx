"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getErrorMessage, useAuth } from "@/components/auth-provider";
import { AuthShell } from "@/components/auth-shell";
import { FormField, inputClassName, primaryButtonClassName } from "@/components/ui";

export default function LoginPage() {
  const { login, user, ready } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (ready && user) router.replace("/dashboard");
  }, [ready, user, router]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setPending(true);
    try {
      await login(email.trim(), password);
    } catch (err) {
      setError(getErrorMessage(err, "Unable to sign in"));
    } finally {
      setPending(false);
    }
  }

  return (
    <AuthShell title="Welcome back" subtitle="Sign in to continue to FinTrack Coach.">
      <form onSubmit={onSubmit} className="space-y-4">
        <FormField label="Email" htmlFor="email">
          <input
            id="email"
            type="email"
            autoComplete="email"
            required
            className={inputClassName()}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </FormField>
        <FormField label="Password" htmlFor="password">
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            required
            className={inputClassName()}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </FormField>
        {error ? <p className="text-sm text-danger">{error}</p> : null}
        <button type="submit" className={`${primaryButtonClassName()} w-full`} disabled={pending}>
          {pending ? "Signing in…" : "Sign in"}
        </button>
      </form>
      <p className="mt-6 text-center text-sm text-muted">
        New here?{" "}
        <Link href="/register" className="font-medium text-primary hover:underline">
          Create an account
        </Link>
      </p>
    </AuthShell>
  );
}
