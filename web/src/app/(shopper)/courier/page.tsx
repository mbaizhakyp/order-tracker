"use client";

import { useEffect, useState } from "react";
import { useSocket } from "@/hooks/useSocket";
import OfferCard from "@/components/shopper/OfferCard";
import { api } from "@/lib/api";
import { Loader2 } from "lucide-react";
import LiveMap from "@/components/maps/LiveMap";
import { useAuth } from "@/context/AuthContext";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";

export default function CourierPage() {
    const { user, logout } = useAuth();
    const router = useRouter();
    const [connectedId, setConnectedId] = useState<string | null>(null);
    const [currentOffer, setCurrentOffer] = useState<any>(null);
    const [activeOrder, setActiveOrder] = useState<any>(null);
    const [shopperLocation, setShopperLocation] = useState<{ lat: number; lng: number } | undefined>(undefined);

    // Auto-connect if logged in user is a SHOPPER
    useEffect(() => {
        if (!user) {
            router.push("/login");
            return;
        }

        if (user.role === "SHOPPER" && !connectedId) {
            setConnectedId(user.id);
        }
    }, [user, connectedId, router]);

    // Connect to WS only when we have a valid connectedId
    const wsUrl = connectedId
        ? `ws://localhost:8080/ws?user_id=${connectedId}`
        : null;

    const { lastMessage, isConnected } = useSocket(wsUrl);

    // Listen for Messages
    useEffect(() => {
        if (!lastMessage) return;
        console.log("WS Message Received:", lastMessage);

        if (lastMessage.type === "NEW_OFFER") {
            setCurrentOffer(lastMessage);
        } else if (lastMessage.type === "SHOPPER_MOVED") {
            const payload = lastMessage.payload;
            if (payload.shopper_id === connectedId) {
                setShopperLocation({ lat: payload.lat, lng: payload.lng });
            }
        } else if (["ORDER_ARRIVED_AT_STORE", "ORDER_PICKED_UP", "ORDER_ARRIVED_AT_CUSTOMER", "ORDER_DELIVERED"].includes(lastMessage.type)) {
            // If this update relates to my active order, update state
            const order = lastMessage.data; // EventEnvelope structure: type, data

            setActiveOrder((prev: any) => {
                if (prev && order.id === prev.id) {
                    return { ...prev, status: order.status };
                }
                return prev;
            });
        }
    }, [lastMessage, connectedId]);

    const handleAcceptOffer = async () => {
        if (!currentOffer) return;
        const payload = currentOffer.payload;
        const orderId = payload.order_id;

        try {
            // API call to claim
            await api.post(`/orders/${orderId}/claim`, {
                shopper_id: connectedId,
            });

            // Success path
            setupActiveOrder(payload);

        } catch (error: any) {
            // If 409, check if we actually own it (idempotency/recovery)
            if (error.response?.status === 409) {
                try {
                    const res = await api.get(`/orders/${orderId}`);
                    const order = res.data;
                    if (order.shopper_id === connectedId) {
                        // We own it! Recover state.
                        setupActiveOrder(payload); // payload has location data we need
                    } else {
                        alert("This order was already taken by another shopper.");
                    }
                } catch (fetchErr) {
                    console.error("Failed to check order status:", fetchErr);
                    alert("Failed to claim order.");
                }
            } else {
                console.error("Claim failed:", error);
                alert("An error occurred while accepting the order.");
            }
        }

        setCurrentOffer(null); // Clear offer popup always
    };

    const setupActiveOrder = (payload: any) => {
        setActiveOrder({
            id: payload.order_id,
            store_lat: payload.store_lat,
            store_lng: payload.store_lng,
            delivery_lat: payload.delivery_lat,
            delivery_lng: payload.delivery_lng,
            status: "CLAIMED"
        });
    };

    const handleDeclineOffer = () => {
        setCurrentOffer(null);
    };

    // Render loading state while connecting
    if (!connectedId) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-black">
                <div className="text-center">
                    <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-blue-600" />
                    <p className="text-zinc-500">Connecting to dispatch...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-zinc-100 dark:bg-black relative">
            {/* Header */}
            <div className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 p-4 flex justify-between items-center shadow-sm z-10 relative">
                <h1 className="font-bold text-lg">OrderTracker Courier</h1>
                <div className="flex items-center gap-4 text-sm">
                    <div className="flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${isConnected ? "bg-green-500" : "bg-red-500"}`} />
                        <span>{isConnected ? "Online" : "Connecting..."}</span>
                    </div>
                    {user && (
                        <div className="flex items-center gap-4 border-l pl-4 border-zinc-200 dark:border-zinc-700">
                            <span className="font-medium text-zinc-600 dark:text-zinc-400">Hello, {user.name}</span>
                            <button onClick={() => { logout(); router.push("/login"); }} className="text-red-500 hover:text-red-600" title="Log Out">
                                <LogOut className="w-4 h-4" />
                            </button>
                        </div>
                    )}
                </div>
            </div>

            {/* Main Content Area (Map) */}
            <div className="absolute inset-0 top-14">
                <LiveMap
                    pickup={
                        (currentOffer?.payload?.store_lat && {
                            lat: currentOffer.payload.store_lat,
                            lng: currentOffer.payload.store_lng
                        }) ||
                        (activeOrder?.store_lat && {
                            lat: activeOrder.store_lat,
                            lng: activeOrder.store_lng
                        }) || undefined
                    }
                    dropoff={
                        (currentOffer?.payload?.delivery_lat && {
                            lat: currentOffer.payload.delivery_lat,
                            lng: currentOffer.payload.delivery_lng
                        }) ||
                        (activeOrder?.delivery_lat && {
                            lat: activeOrder.delivery_lat,
                            lng: activeOrder.delivery_lng
                        }) || undefined
                    }
                    shopper={shopperLocation}
                />
            </div>

            {/* Offer Popup */}
            {currentOffer && (
                <OfferCard
                    offer={currentOffer}
                    onAccept={handleAcceptOffer}
                    onDecline={handleDeclineOffer}
                />
            )}

            {/* Active Order Overlay */}
            {activeOrder && !currentOffer && (
                <div className="absolute bottom-8 left-4 right-4 z-50">
                    <div className="bg-white dark:bg-zinc-900 p-4 rounded-xl shadow-xl border border-blue-500/20 max-w-md mx-auto">
                        <div className="flex justify-between items-center mb-2">
                            <h3 className="font-bold text-lg">Active Order</h3>
                            <span className="bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded-full font-medium">
                                {activeOrder.status?.replace(/_/g, " ") || "IN PROGRESS"}
                            </span>
                        </div>
                        <p className="text-sm text-zinc-500 mb-4">
                            {activeOrder.status === "CLAIMED" && "Drive to the store."}
                            {activeOrder.status === "ARRIVED_AT_STORE" && "You have arrived. Pick up the order."}
                            {activeOrder.status === "PICKED_UP" && "Drive to the customer."}
                            {activeOrder.status === "DELIVERED" && "Great job!"}
                        </p>

                        {(activeOrder.status === "CLAIMED" || activeOrder.status === "ARRIVED_AT_STORE") && (
                            <button
                                disabled={activeOrder.status !== "ARRIVED_AT_STORE"}
                                onClick={async () => {
                                    await api.post(`/orders/${activeOrder.id}/pickup`);
                                    setActiveOrder((prev: any) => ({ ...prev, status: "PICKED_UP" }));
                                }}
                                className={`mt-4 w-full py-3 text-white rounded-lg font-bold transition shadow-lg ${activeOrder.status === "ARRIVED_AT_STORE"
                                    ? "bg-blue-600 hover:bg-blue-700"
                                    : "bg-zinc-300 dark:bg-zinc-700 cursor-not-allowed text-zinc-500"
                                    }`}
                            >
                                {activeOrder.status === "CLAIMED" ? "Driving to Store..." : "PICK UP ORDER"}
                            </button>
                        )}

                        {(activeOrder.status === "PICKED_UP" || activeOrder.status === "ARRIVED_AT_CUSTOMER") && (
                            <button
                                disabled={activeOrder.status !== "ARRIVED_AT_CUSTOMER"}
                                onClick={async () => {
                                    await api.post(`/orders/${activeOrder.id}/deliver`);
                                    setActiveOrder((prev: any) => ({ ...prev, status: "DELIVERED" }));
                                    setTimeout(() => setActiveOrder(null), 3000); // Clear after 3s
                                }}
                                className={`mt-4 w-full py-3 text-white rounded-lg font-bold transition shadow-lg ${activeOrder.status === "ARRIVED_AT_CUSTOMER"
                                    ? "bg-green-600 hover:bg-green-700"
                                    : "bg-zinc-300 dark:bg-zinc-700 cursor-not-allowed text-zinc-500"
                                    }`}
                            >
                                {activeOrder.status === "PICKED_UP" ? "Driving to Customer..." : "MARK AS DELIVERED"}
                            </button>
                        )}

                        {activeOrder.status === "DELIVERED" && (
                            <div className="mt-4 w-full py-3 bg-green-50 text-green-700 text-center rounded-lg font-bold">
                                Order Completed!
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}
