import { useState, useEffect } from 'react';

interface WebSocketMessage {
  type: 'measurement' | 'alert' | 'incident' | 'probe_status' | 'target_status';
  data: any;
  timestamp: string;
}

type MessageHandler = (message: WebSocketMessage) => void;

class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectInterval: number = 5000;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private messageHandlers: Set<MessageHandler> = new Set();
  private isIntentionallyClosed: boolean = false;

  constructor(url: string) {
    this.url = url;
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    this.isIntentionallyClosed = false;

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
          this.reconnectTimer = null;
        }
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.messageHandlers.forEach(handler => handler(message));
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        if (!this.isIntentionallyClosed) {
          this.scheduleReconnect();
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  disconnect() {
    this.isIntentionallyClosed = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimer || this.isIntentionallyClosed) {
      return;
    }

    this.reconnectTimer = setTimeout(() => {
      console.log('Attempting to reconnect WebSocket...');
      this.reconnectTimer = null;
      this.connect();
    }, this.reconnectInterval);
  }

  subscribe(handler: MessageHandler) {
    this.messageHandlers.add(handler);
    return () => {
      this.messageHandlers.delete(handler);
    };
  }

  send(data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  getState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }
}

let wsClient: WebSocketClient | null = null;

export const getWebSocketClient = (): WebSocketClient => {
  if (!wsClient) {
    const wsUrl = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';
    wsClient = new WebSocketClient(wsUrl);
  }
  return wsClient;
};

export const useWebSocket = (onMessage: MessageHandler) => {
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    const client = getWebSocketClient();
    client.connect();

    const unsubscribe = client.subscribe(onMessage);

    // Check connection status periodically
    const checkConnection = setInterval(() => {
      setIsConnected(client.getState() === WebSocket.OPEN);
    }, 1000);

    return () => {
      clearInterval(checkConnection);
      unsubscribe();
    };
  }, [onMessage]);

  return { isConnected };
};
