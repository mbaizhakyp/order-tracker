"use client";

import { useEffect, useState } from "react";
import { useSocket } from "@/hooks/useSocket";
import OfferCard from "@/components/shopper/OfferCard";
import { api } from "@/lib/api";
import { Loader2 } from "lucide-react";
import LiveMap from "@/components/maps/LiveMap";



export default function CourierPage() {
    const [shopperId, setShopperId] = useState("");
    const [connectedId, setConnectedId] = useState<string | null>(null);
    const [currentOffer, setCurrentOffer] = useState<any>(null);
    const [activeOrder, setActiveOrder] = useState<any>(null);
    const [shopperLocation, setShopperLocation] = useState<{ lat: number; lng: number } | undefined>(undefined);

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

    const handleLogin = (e: React.FormEvent) => {
        e.preventDefault();
        const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

        if (!uuidRegex.test(shopperId.trim())) {
            alert("Please enter a valid UUID (e.g. a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33)");
            return;
        }

        if (shopperId.trim()) {
            setConnectedId(shopperId);
        }
    };

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

    if (!connectedId) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-black p-4">
                <div className="w-full max-w-md bg-white dark:bg-zinc-900 p-8 rounded-xl shadow-lg border border-zinc-200 dark:border-zinc-800">
                    <h1 className="text-2xl font-bold mb-6 text-center">Shopper Login</h1>
                    <form onSubmit={handleLogin} className="space-y-4">
                        <div>
                            <label htmlFor="shopperId" className="block text-sm font-medium mb-1">
                                Enter Shopper UUID
                            </label>
                            <input
                                id="shopperId"
                                type="text"
                                value={shopperId}
                                onChange={(e) => setShopperId(e.target.value)}
                                placeholder="e.g. a0eebc99-..."
                                className="w-full px-4 py-2 rounded-lg border border-zinc-300 dark:border-zinc-700 bg-transparent"
                                required
                            />
                        </div>
                        <button
                            type="submit"
                            className="w-full py-2 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition"
                        >
                            Start Shift
                        </button>
                    </form>
                    <div className="mt-6 text-xs text-zinc-500 text-center">
                        <p>Test ID: a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33</p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-zinc-100 dark:bg-black relative">
            {/* Header */}
            <div className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 p-4 flex justify-between items-center shadow-sm z-10 relative">
                <h1 className="font-bold text-lg">OrderTracker Courier</h1>
                <div className="flex items-center gap-2 text-sm">
                    <div className={`w-2 h-2 rounded-full ${isConnected ? "bg-green-500" : "bg-red-500"}`} />
                    <span>{isConnected ? "Online" : "Connecting..."}</span>
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
