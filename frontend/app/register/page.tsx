"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getErrorMessage, useAuth } from "@/components/auth-provider";
import { AuthShell } from "@/components/auth-shell";
import { FormField, inputClassName, primaryButtonClassName } from "@/components/ui";

export default function RegisterPage() {
  const { register, user, ready } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (ready && user) router.replace("/dashboard");
  }, [ready, user, router]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    if (password !== confirm) {
      setError("Passwords do not match");
      return;
    }
    setPending(true);
    try {
      await register(email.trim(), password);
    } catch (err) {
      setError(getErrorMessage(err, "Unable to register"));
    } finally {
      setPending(false);
    }
  }

  return (
    <AuthShell title="Create your account" subtitle="Start tracking with FinTrack Coach">
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
        <FormField
          label="Password"
          htmlFor="password"
          hint="At least 8 characters with upper, lower, and a digit"
        >
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            required
            className={inputClassName()}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </FormField>
        <FormField label="Confirm password" htmlFor="confirm">
          <input
            id="confirm"
            type="password"
            autoComplete="new-password"
            required
            className={inputClassName()}
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
          />
        </FormField>
        {error ? <p className="text-sm text-danger">{error}</p> : null}
        <button type="submit" className={`${primaryButtonClassName()} w-full`} disabled={pending}>
          {pending ? "Creating account…" : "Create account"}
        </button>
      </form>
      <p className="mt-6 text-center text-sm text-muted">
        Already have an account?{" "}
        <Link href="/login" className="font-medium text-primary hover:underline">
          Sign in
        </Link>
      </p>
    </AuthShell>
  );
}
