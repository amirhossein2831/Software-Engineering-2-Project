"use client";

import { useEffect, useRef } from "react";

function wsBase() {
  if (process.env.NEXT_PUBLIC_WS_URL) return process.env.NEXT_PUBLIC_WS_URL;
  if (typeof window === "undefined") return "";
  const proto = window.location.protocol === "https:" ? "wss" : "ws";
  return `${proto}://${window.location.host}/api`;
}

export interface RealtimeMessage {
  type: string;
  [key: string]: unknown;
}

export function useRealtime(
  channels: string[],
  onMessage: (msg: RealtimeMessage) => void,
) {
  const handlerRef = useRef(onMessage);
  handlerRef.current = onMessage;

  const key = channels.join(",");

  useEffect(() => {
    if (!key) return;
    const base = wsBase();
    if (!base) return;
    const url = `${base}/ws?channels=${encodeURIComponent(key)}`;
    let ws: WebSocket | null = null;
    let closed = false;
    let retry: ReturnType<typeof setTimeout>;

    const connect = () => {
      ws = new WebSocket(url);
      ws.onmessage = (ev) => {
        try {
          handlerRef.current(JSON.parse(ev.data) as RealtimeMessage);
        } catch {
          // ignore malformed frames
        }
      };
      ws.onclose = () => {
        if (!closed) retry = setTimeout(connect, 2000);
      };
    };
    connect();

    return () => {
      closed = true;
      clearTimeout(retry);
      ws?.close();
    };
  }, [key]);
}
