import { useEffect, useRef, useState, useCallback } from "react";

export interface WebSocketMessage {
    type: string;
    payload: any;
}

export function useSocket(url: string | null) {
    const [isConnected, setIsConnected] = useState(false);
    const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
    const socketRef = useRef<WebSocket | null>(null);

    useEffect(() => {
        if (!url) {
            setIsConnected(false);
            return;
        }

        const ws = new WebSocket(url);
        socketRef.current = ws;

        ws.onopen = () => {
            console.log("WebSocket Connected");
            setIsConnected(true);
        };

        ws.onclose = () => {
            console.log("WebSocket Disconnected");
            setIsConnected(false);
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                setLastMessage(data);
            } catch (e) {
                console.error("Failed to parse WS message", event.data);
            }
        };

        return () => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.close();
            }
        };
    }, [url]);

    return { isConnected, lastMessage };
}
