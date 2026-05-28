import type { WSEvent } from "@/types";

type EventHandler = (event: WSEvent) => void;

class WSClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<EventHandler>> = new Map();
  private reconnectDelay = 1000;
  private maxReconnectDelay = 30000;

  connect() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const url = `${protocol}//${window.location.host}/api/v1/ws`;

    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      console.log("WebSocket connected");
      this.reconnectDelay = 1000;
    };

    this.ws.onmessage = (msg) => {
      try {
        const event: WSEvent = JSON.parse(msg.data);
        // Notify all handlers
        this.handlers.get(event.type)?.forEach((h) => h(event));
        // Notify wildcard handlers
        this.handlers.get("*")?.forEach((h) => h(event));
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      console.log("WebSocket disconnected, reconnecting...");
      setTimeout(() => this.connect(), this.reconnectDelay);
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
    };
  }

  on(type: string, handler: EventHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => this.handlers.get(type)?.delete(handler);
  }

  disconnect() {
    this.ws?.close();
  }
}

export const wsClient = new WSClient();
