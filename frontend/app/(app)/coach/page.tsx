"use client";

import { FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { AppHeader } from "@/components/app-shell";
import { getErrorMessage } from "@/components/auth-provider";
import {
  CoachMessageBubble,
  EmptyState,
  ErrorState,
  LoadingSkeleton,
  inputClassName,
  primaryButtonClassName,
  secondaryButtonClassName,
} from "@/components/ui";
import { api, ApiError } from "@/lib/api";
import { useDeferredLoad } from "@/lib/use-deferred-load";
import type { CoachConversation, CoachMessage } from "@/lib/types";

const suggestions = [
  "What is my current total balance?",
  "Where did most of my spending go this month?",
  "Compare income and expenses over the last six months.",
  "Did my latest CSV import succeed?",
];

export default function CoachPage() {
  const [conversations, setConversations] = useState<CoachConversation[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [messages, setMessages] = useState<CoachMessage[]>([]);
  const [input, setInput] = useState("");
  const [loadingList, setLoadingList] = useState(true);
  const [loadingThread, setLoadingThread] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const bottomRef = useRef<HTMLDivElement | null>(null);

  const loadConversations = useCallback(async () => {
    setLoadingList(true);
    setError("");
    try {
      const res = await api.listConversations(1, 50);
      setConversations(res.data ?? []);
    } catch (err) {
      setError(getErrorMessage(err, "Failed to load conversations"));
    } finally {
      setLoadingList(false);
    }
  }, []);

  useDeferredLoad(loadConversations, [loadConversations]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, sending]);

  async function selectConversation(id: string) {
    setSelectedId(id);
    setLoadingThread(true);
    setError("");
    try {
      const res = await api.getConversation(id);
      setMessages(res.data.messages ?? []);
    } catch (err) {
      setError(getErrorMessage(err, "Failed to load conversation"));
      setMessages([]);
    } finally {
      setLoadingThread(false);
    }
  }

  function startNewConversation() {
    setSelectedId(null);
    setMessages([]);
    setInput("");
    setError("");
  }

  async function sendMessage(raw: string) {
    const message = raw.trim();
    if (!message || sending) return;

    setSending(true);
    setError("");
    const optimisticId = `local-${crypto.randomUUID()}`;
    const optimistic: CoachMessage = {
      id: optimisticId,
      role: "user",
      content: message,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, optimistic]);
    setInput("");

    try {
      const res = await api.coachChat(message, selectedId ?? undefined);
      setSelectedId(res.data.conversation_id);
      const detail = await api.getConversation(res.data.conversation_id);
      setMessages(detail.data.messages ?? []);
      await loadConversations();
    } catch (err) {
      setMessages((prev) => prev.filter((item) => item.id !== optimisticId));
      if (err instanceof ApiError && (err.status === 503 || err.status === 502)) {
        setError("Financial coach is temporarily unavailable. Check AI configuration and try again.");
      } else {
        setError(getErrorMessage(err, "Unable to reach FinTrack Coach"));
      }
    } finally {
      setSending(false);
    }
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    await sendMessage(input);
  }

  return (
    <div className="flex h-[calc(100vh-3rem)] flex-col lg:h-[calc(100vh-4rem)]">
      <AppHeader
        title="AI Coach"
        subtitle="Ask grounded questions about your FinTrack data"
      />

      <div className="grid min-h-0 flex-1 gap-4 lg:grid-cols-[280px_1fr]">
        <aside className="flex min-h-0 flex-col rounded-xl border border-border bg-card shadow-sm">
          <div className="border-b border-border p-3">
            <button type="button" className={`${primaryButtonClassName()} w-full`} onClick={startNewConversation}>
              New conversation
            </button>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto p-2">
            {loadingList ? <LoadingSkeleton rows={4} /> : null}
            {!loadingList && conversations.length === 0 ? (
              <p className="p-3 text-sm text-muted">No conversations yet.</p>
            ) : null}
            {conversations.map((conversation) => (
              <button
                key={conversation.id}
                type="button"
                onClick={() => void selectConversation(conversation.id)}
                className={`mb-1 w-full rounded-lg px-3 py-2 text-left text-sm ${
                  selectedId === conversation.id ? "bg-primary-soft text-primary" : "hover:bg-slate-50"
                }`}
              >
                <span className="line-clamp-2 font-medium">
                  {conversation.title || "Untitled conversation"}
                </span>
              </button>
            ))}
          </div>
        </aside>

        <section className="flex min-h-0 flex-col rounded-xl border border-border bg-card shadow-sm">
          <div className="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
            {error ? <ErrorState message={error} /> : null}
            {loadingThread ? <LoadingSkeleton rows={4} /> : null}

            {!loadingThread && messages.length === 0 ? (
              <EmptyState
                title="Ask FinTrack Coach a question"
                description="The coach uses your accounts, transactions, dashboard summaries, and imports — it will not invent balances."
                action={
                  <div className="flex flex-wrap gap-2">
                    {suggestions.map((prompt) => (
                      <button
                        key={prompt}
                        type="button"
                        className={secondaryButtonClassName()}
                        onClick={() => void sendMessage(prompt)}
                        disabled={sending}
                      >
                        {prompt}
                      </button>
                    ))}
                  </div>
                }
              />
            ) : null}

            {messages.map((message) => (
              <CoachMessageBubble key={message.id} role={message.role} content={message.content} />
            ))}
            {sending ? (
              <p className="text-sm text-muted" aria-live="polite">
                Coach is thinking…
              </p>
            ) : null}
            <div ref={bottomRef} />
          </div>

          <form onSubmit={onSubmit} className="border-t border-border p-3">
            <div className="flex gap-2">
              <label className="sr-only" htmlFor="coach_input">
                Message
              </label>
              <textarea
                id="coach_input"
                rows={2}
                className={inputClassName()}
                placeholder="Ask about balances, spending, trends, or imports…"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                disabled={sending}
              />
              <button type="submit" className={primaryButtonClassName()} disabled={sending || !input.trim()}>
                Send
              </button>
            </div>
          </form>
        </section>
      </div>
    </div>
  );
}
