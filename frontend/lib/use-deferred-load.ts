"use client";

import { useEffect } from "react";

/** Defer mount/load work so setState is not synchronous inside the effect body. */
export function useDeferredLoad(load: () => void | Promise<void>, deps: unknown[]) {
  useEffect(() => {
    let cancelled = false;
    const timer = window.setTimeout(() => {
      if (!cancelled) {
        void load();
      }
    }, 0);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- caller controls deps explicitly
  }, deps);
}
