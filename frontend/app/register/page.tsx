"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Eye, EyeOff } from "lucide-react";
import { getErrorMessage, useAuth } from "@/components/auth-provider";
import { AuthShell } from "@/components/auth-shell";
import { FormField, inputClassName, primaryButtonClassName } from "@/components/ui";

export default function RegisterPage() {
  const { register, user, ready } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
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
    <AuthShell title="Create your account" subtitle="Start tracking and understanding your finances.">
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
          <div className="relative">
            <input
              id="password"
              type={showPassword ? "text" : "password"}
              autoComplete="new-password"
              required
              className={`${inputClassName()} pr-10`}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <button
              type="button"
              aria-label={showPassword ? "Hide password" : "Show password"}
              aria-pressed={showPassword}
              onClick={() => setShowPassword((visible) => !visible)}
              className="absolute right-1 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-muted transition-colors hover:bg-teal-50 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30"
            >
              {showPassword ? (
                <EyeOff className="h-4 w-4" aria-hidden />
              ) : (
                <Eye className="h-4 w-4" aria-hidden />
              )}
            </button>
          </div>
        </FormField>
        <FormField label="Confirm password" htmlFor="confirm">
          <div className="relative">
            <input
              id="confirm"
              type={showConfirm ? "text" : "password"}
              autoComplete="new-password"
              required
              className={`${inputClassName()} pr-10`}
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
            />
            <button
              type="button"
              aria-label={showConfirm ? "Hide password" : "Show password"}
              aria-pressed={showConfirm}
              onClick={() => setShowConfirm((visible) => !visible)}
              className="absolute right-1 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-muted transition-colors hover:bg-teal-50 hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/30"
            >
              {showConfirm ? (
                <EyeOff className="h-4 w-4" aria-hidden />
              ) : (
                <Eye className="h-4 w-4" aria-hidden />
              )}
            </button>
          </div>
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
