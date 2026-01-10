"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import { useSocket } from "@/hooks/useSocket";
import LiveMap from "@/components/maps/LiveMap";
import { Loader2, Package } from "lucide-react";

interface Order {
    id: string;
    status: string;
    delivery_lat: number;
    delivery_lng: number;
    store_id: string; // Ideally we'd get store lat/lng too, but for demo we know it's Target
    shopper_id?: string;
}

export default function OrderTrackingPage() {
    const params = useParams();
    const orderId = params.id as string;
    const [order, setOrder] = useState<Order | null>(null);
    const [loading, setLoading] = useState(true);

    // Target, Tuscaloosa (Hardcoded for Demo consistency with Backend)
    const storeLocation = { lat: 33.1956, lng: -87.5268 };

    // 3. Connect to WebSocket
    const socketUrl = orderId ? `ws://localhost:8080/ws?user_id=customer_tracker_${orderId}` : null;
    const { lastMessage } = useSocket(socketUrl);

    const [shopperLocation, setShopperLocation] = useState<{ lat: number; lng: number } | undefined>(undefined);

    useEffect(() => {
        if (!orderId) return;

        const fetchOrder = async () => {
            try {
                const res = await api.get(`/orders/${orderId}`);
                console.log("Fetched Order:", res.data);
                setOrder(res.data);
            } catch (err) {
                console.error("Failed to fetch order", err);
            } finally {
                setLoading(false);
            }
        };

        fetchOrder();

        // Poll for updates every 5 seconds (Simple "Real-time" for now)
        const interval = setInterval(fetchOrder, 5000);
        return () => clearInterval(interval);
    }, [orderId]);

    // 4. Handle Real-Time Updates
    useEffect(() => {
        if (!lastMessage) return;

        if (lastMessage.type === "SHOPPER_MOVED") {
            const payload = lastMessage.payload;
            if (order && order.shopper_id && payload.shopper_id === order.shopper_id) {
                setShopperLocation({ lat: payload.lat, lng: payload.lng });
            }
        } else if (["ORDER_CLAIMED", "ORDER_ARRIVED_AT_STORE", "ORDER_PICKED_UP", "ORDER_ARRIVED_AT_CUSTOMER", "ORDER_DELIVERED"].includes(lastMessage.type)) {
            // Real-time status update
            const updatedOrder = lastMessage.data;
            if (updatedOrder.id === orderId) {
                setOrder(updatedOrder);
            }
        }
    }, [lastMessage, order]);

    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-black">
                <Loader2 className="w-8 h-8 animate-spin text-zinc-400" />
            </div>
        );
    }

    if (!order) {
        return (
            <div className="min-h-screen flex items-center justify-center text-red-500">
                Order not found.
            </div>
        );
    }




    return (
        <div className="min-h-screen bg-zinc-100 dark:bg-black relative">
            {/* Header / Status Bar */}
            <div className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 p-4 shadow-sm z-20 absolute top-0 left-0 right-0 h-20 flex items-center">
                <div className="max-w-5xl mx-auto w-full flex justify-between items-center">
                    <div>
                        <h1 className="font-bold text-lg flex items-center gap-2">
                            <Package className="w-5 h-5 text-blue-600" />
                            Tracking Order
                        </h1>
                        <p className="text-xs text-zinc-500">#{order.id.slice(0, 8)}</p>
                    </div>

                    <div className="flex items-center gap-3">
                        <div className="text-right">
                            <div className="text-sm font-bold text-zinc-900 dark:text-zinc-100">
                                {order.status === "CREATED" || order.status === "OFFERED"
                                    ? "Searching for Shopper..."
                                    : order.status}
                            </div>
                            <div className="text-xs text-zinc-500">
                                {order.status === "CREATED" || order.status === "OFFERED"
                                    ? "Hang tight!"
                                    : order.status === "CLAIMED"
                                        ? "Shopper is driving to store"
                                        : order.status === "ARRIVED_AT_STORE"
                                            ? "Shopper is at the store"
                                            : order.status === "PICKED_UP"
                                                ? "On the way to you!"
                                                : "Delivered"}
                            </div>
                        </div>
                        {/* Status Indicator */}
                        <div className={`w-3 h-3 rounded-full ${order.status === "CREATED" ? "bg-yellow-400 animate-pulse" :
                            order.status === "OFFERED" ? "bg-orange-400 animate-pulse" :
                                order.status === "CLAIMED" ? "bg-blue-500" :
                                    "bg-green-500"
                            }`} />
                    </div>
                </div>
            </div>

            {/* Map Area */}
            <div className="absolute inset-x-0 bottom-0 top-20 z-10 transition-all">
                <LiveMap
                    pickup={storeLocation}
                    dropoff={{ lat: order.delivery_lat, lng: order.delivery_lng }}
                    shopper={shopperLocation}
                />
            </div>
        </div>
    );
}
