"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
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
    shopper_name?: string;
}

export default function OrderTrackingPage() {
    const router = useRouter();
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




    // 5. Render: Receipt Mode (Delivered/Cancelled)
    if (["DELIVERED", "CANCELLED"].includes(order.status)) {
        return (
            <div className="min-h-screen bg-zinc-100 dark:bg-black flex items-center justify-center p-4">
                <div className="bg-white dark:bg-zinc-900 w-full max-w-md rounded-2xl shadow-xl border border-zinc-200 dark:border-zinc-800 overflow-hidden">
                    {/* Header */}
                    <div className="bg-zinc-50 dark:bg-zinc-800/50 p-8 text-center border-b border-zinc-100 dark:border-zinc-800">
                        <div className={`w-24 h-24 mx-auto rounded-full flex items-center justify-center mb-6 shadow-sm ${order.status === "DELIVERED" ? "bg-green-100 text-green-600" : "bg-red-100 text-red-600"
                            }`}>
                            {order.status === "DELIVERED" ? (
                                <svg xmlns="http://www.w3.org/2000/svg" className="w-12 h-12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                            ) : (
                                <svg xmlns="http://www.w3.org/2000/svg" className="w-12 h-12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                            )}
                        </div>
                        <h1 className="text-3xl font-bold mb-2 text-zinc-900 dark:text-white">
                            {order.status === "DELIVERED" ? "Order Delivered!" : "Order Cancelled"}
                        </h1>
                        <p className="text-zinc-500">
                            {order.status === "DELIVERED" ? "We hope you enjoy your items." : "This order was cancelled."}
                        </p>
                    </div>

                    {/* Receipt Details */}
                    <div className="p-8 space-y-8">
                        {/* Driver Info */}
                        {order.status === "DELIVERED" && order.shopper_name && (
                            <div className="flex items-center gap-4 p-4 bg-blue-50 dark:bg-blue-900/10 border border-blue-100 dark:border-blue-800 rounded-xl">
                                <div className="w-12 h-12 rounded-full bg-blue-200 text-blue-700 flex items-center justify-center font-bold text-lg">
                                    {order.shopper_name.charAt(0)}
                                </div>
                                <div>
                                    <p className="text-xs text-blue-500 dark:text-blue-400 font-bold uppercase tracking-wider">Delivered By</p>
                                    <p className="font-bold text-lg text-blue-900 dark:text-blue-100">{order.shopper_name}</p>
                                </div>
                            </div>
                        )}

                        {/* Item List */}
                        <div>
                            <h3 className="text-xs font-bold text-zinc-400 uppercase tracking-widest mb-4">Order Summary</h3>
                            <div className="space-y-3">
                                {order.items && (Array.isArray(order.items) ? order.items : JSON.parse(order.items as any)).map((item: any, i: number) => (
                                    <div key={i} className="flex justify-between items-center text-sm">
                                        <div className="flex items-center gap-2">
                                            <span className="font-bold text-zinc-900 dark:text-white">{item.quantity}x</span>
                                            <span className="text-zinc-600 dark:text-zinc-400">{item.name}</span>
                                        </div>
                                    </div>
                                ))}
                                <div className="pt-4 mt-4 border-t border-dashed border-zinc-200 dark:border-zinc-700 flex justify-between items-center">
                                    <span className="font-bold text-zinc-500">Total Paid</span>
                                    <span className="font-bold text-xl text-zinc-900 dark:text-white">${(order.total_amount / 100).toFixed(2)}</span>
                                </div>
                            </div>
                        </div>

                        {/* Action */}
                        <button
                            onClick={() => router.push("/customer/orders")}
                            className="w-full py-4 text-lg bg-zinc-900 dark:bg-white text-white dark:text-black rounded-xl font-bold hover:opacity-90 transition shadow-lg"
                        >
                            Return to Orders
                        </button>
                    </div>
                </div>
            </div>
        );
    }

    // 6. Render: Tracking Mode (Active)
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
                        <div className="w-3 h-3 rounded-full bg-blue-500 animate-pulse" />
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

                        {["CREATED", "OFFERED", "CLAIMED"].includes(order.status) && (
                            <button
                                onClick={async () => {
                                    if (!confirm("Are you sure you want to cancel this order?")) return;
                                    try {
                                        await api.post(`/orders/${orderId}/cancel`);
                                        alert("Order Cancelled");
                                        router.push(`/customer/store/${order.store_id}`);
                                    } catch (err) {
                                        console.error(err);
                                        alert("Failed to cancel order");
                                    }
                                }}
                                className="mt-6 w-full py-2 bg-red-50 text-red-600 border border-red-200 rounded-lg text-sm font-medium hover:bg-red-100 transition"
                            >
                                Cancel Order
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* Map Area */}
            <div className="flex-1 h-[60vh] md:h-screen relative bg-zinc-100 dark:bg-black flex items-center justify-center">
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
