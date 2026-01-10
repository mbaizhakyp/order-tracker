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
        <div className="min-h-screen bg-zinc-100 dark:bg-black relative flex flex-col md:flex-row">
            {/* Sidebar / Info Panel */}
            <div className="w-full md:w-96 bg-white dark:bg-zinc-900 h-[40vh] md:h-screen overflow-y-auto border-r border-zinc-200 dark:border-zinc-800 shadow-lg z-20 flex flex-col">
                <div className="p-6 border-b border-zinc-200 dark:border-zinc-800">
                    <h1 className="font-bold text-xl flex items-center gap-2 mb-1">
                        <Package className="w-6 h-6 text-blue-600" />
                        Tracking Order
                    </h1>
                    <p className="text-sm text-zinc-500">ID: #{order.id.slice(0, 8)}</p>
                </div>

                <div className="p-6 flex-1">
                    {/* Status Card */}
                    <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-xl mb-8 flex justify-between items-center border border-blue-100 dark:border-blue-800">
                        <div>
                            <p className="text-xs text-blue-600 dark:text-blue-400 font-bold uppercase tracking-wider mb-1">Current Status</p>
                            <p className="font-bold text-lg text-blue-900 dark:text-blue-100">
                                {order.status.replace(/_/g, " ")}
                            </p>
                        </div>
                        <div className={`w-3 h-3 rounded-full ${order.status === "DELIVERED" ? "bg-green-500" : "bg-blue-500 animate-pulse"}`} />
                    </div>

                    {/* Timeline */}
                    <OrderTimeline orderId={orderId} status={order.status} />

                    {/* Items (Mock) */}
                    <div className="border-t border-zinc-100 dark:border-zinc-800 pt-6">
                        <h3 className="font-bold mb-4">Order Summary</h3>
                        {order.items && (Array.isArray(order.items) ? order.items : JSON.parse(order.items as any)).map((item: any, i: number) => (
                            <div key={i} className="flex justify-between text-sm py-1">
                                <span className="text-zinc-600 dark:text-zinc-400">{item.quantity}x {item.name}</span>
                            </div>
                        ))}
                        <div className="mt-4 pt-4 border-t border-dashed border-zinc-200 dark:border-zinc-700 flex justify-between font-bold">
                            <span>Total</span>
                            <span>${(order.total_amount / 100).toFixed(2)}</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Map Area */}
            <div className="flex-1 h-[60vh] md:h-screen relative">
                <LiveMap
                    pickup={storeLocation}
                    dropoff={{ lat: order.delivery_lat, lng: order.delivery_lng }}
                    shopper={shopperLocation}
                />
            </div>
        </div>
    );
}

function OrderTimeline({ orderId, status }: { orderId: string, status: string }) {
    const [events, setEvents] = useState<any[]>([]);

    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const res = await api.get(`/orders/${orderId}/history`);
                console.log("History:", res.data);
                setEvents(res.data || []);
            } catch (err) {
                console.error("Failed to fetch history", err);
            }
        };
        fetchHistory();
    }, [orderId, status]);

    return (
        <div className="mb-8">
            <h4 className="text-xs font-bold text-zinc-400 uppercase tracking-widest mb-4">Tracking History</h4>
            <div className="space-y-0 relative pl-4 border-l-2 border-zinc-100 dark:border-zinc-800 ml-2">
                {events && events.map((event, i) => (
                    <div key={event.id} className="relative pl-6 pb-6 last:pb-0">
                        {/* Dot */}
                        <div className={`absolute -left-[9px] top-0 w-4 h-4 rounded-full border-2 border-white dark:border-zinc-900 ${i === events.length - 1 ? "bg-blue-600" : "bg-zinc-300 dark:bg-zinc-700"
                            }`} />

                        <div>
                            <p className={`text-sm font-medium leading-none ${i === events.length - 1 ? "text-zinc-900 dark:text-white" : "text-zinc-500"
                                }`}>
                                {event.status.replace(/_/g, " ")}
                            </p>
                            <p className="text-xs text-zinc-400 mt-1">
                                {new Date(event.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                            </p>
                        </div>
                    </div>
                ))}
                {(!events || events.length === 0) && (
                    <p className="text-sm text-zinc-400 italic pl-6">No updates yet...</p>
                )}
            </div>
        </div>
    );
}
