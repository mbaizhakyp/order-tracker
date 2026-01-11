"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { ArrowLeft, Clock, ShoppingBag, MapPin, ChevronRight, Package, Loader2 } from "lucide-react";

interface Order {
    id: string;
    store_name: string;
    status: string;
    total_amount: number;
    created_at: string;
    items: any;
}

export default function OrderHistoryPage() {
    const router = useRouter();
    const { user } = useAuth();
    const [orders, setOrders] = useState<Order[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchOrders = async () => {
            try {
                const res = await api.get("/orders");
                setOrders(res.data || []);
            } catch (err) {
                console.error("Failed to fetch orders", err);
            } finally {
                setLoading(false);
            }
        };
        fetchOrders();
    }, []);

    const getStatusColor = (status: string) => {
        switch (status) {
            case "DELIVERED": return "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400";
            case "CANCELLED": return "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400";
            case "CREATED": return "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-400";
            default: return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400";
        }
    };

    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 p-4 sticky top-0 z-10">
                <div className="max-w-3xl mx-auto flex items-center gap-4">
                    <button
                        onClick={() => router.push("/customer")}
                        className="p-2 -ml-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-full transition"
                    >
                        <ArrowLeft className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                    </button>
                    <h1 className="font-bold text-xl">My Orders</h1>
                </div>
            </header>

            <main className="max-w-3xl mx-auto p-4 md:p-8">
                {loading ? (
                    <div className="flex justify-center p-12">
                        <Loader2 className="animate-spin text-zinc-400 w-8 h-8" />
                    </div>
                ) : orders.length === 0 ? (
                    <div className="text-center py-20 text-zinc-500">
                        <ShoppingBag className="w-12 h-12 mx-auto mb-4 opacity-20" />
                        <p>No orders found.</p>
                        <button
                            onClick={() => router.push("/customer")}
                            className="mt-4 text-blue-600 hover:underline"
                        >
                            Start Shopping
                        </button>
                    </div>
                ) : (
                    <div className="space-y-4">
                        {orders.map((order) => {
                            const date = new Date(order.created_at).toLocaleDateString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
                            const items = order.items ? (Array.isArray(order.items) ? order.items : JSON.parse(order.items as any)) : [];
                            const itemsSummary = items.map((i: any) => `${i.quantity}x ${i.name}`).join(", ");
                            const isActive = !["DELIVERED", "CANCELLED"].includes(order.status);

                            return (
                                <div
                                    key={order.id}
                                    onClick={() => router.push(`/customer/track/${order.id}`)}
                                    className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-xl p-4 hover:border-blue-500 dark:hover:border-blue-500 cursor-pointer transition group shadow-sm hover:shadow-md"
                                >
                                    <div className="flex justify-between items-start mb-3">
                                        <div className="flex items-center gap-2">
                                            <div className="bg-zinc-100 dark:bg-zinc-800 p-2 rounded-lg">
                                                <MapPin className="w-5 h-5 text-zinc-500" />
                                            </div>
                                            <div>
                                                <h3 className="font-bold text-zinc-900 dark:text-white">{order.store_name}</h3>
                                                <p className="text-xs text-zinc-500 flex items-center gap-1">
                                                    <Clock className="w-3 h-3" />
                                                    {date}
                                                </p>
                                            </div>
                                        </div>
                                        <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold tracking-wide ${getStatusColor(order.status)}`}>
                                            {order.status.replace(/_/g, " ")}
                                        </span>
                                    </div>

                                    <div className="bg-zinc-50 dark:bg-zinc-800/50 rounded-lg p-3 mb-3">
                                        <p className="text-sm text-zinc-600 dark:text-zinc-300 line-clamp-1">
                                            {itemsSummary || "No items"}
                                        </p>
                                    </div>

                                    <div className="flex justify-between items-center text-sm">
                                        <span className="font-bold text-zinc-900 dark:text-zinc-100">
                                            ${(order.total_amount / 100).toFixed(2)}
                                        </span>
                                        <div className="flex items-center gap-1 text-blue-600 font-medium group-hover:translate-x-1 transition-transform">
                                            {isActive ? "Track Order" : "View Details"}
                                            <ChevronRight className="w-4 h-4" />
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </main>
        </div>
    );
}
