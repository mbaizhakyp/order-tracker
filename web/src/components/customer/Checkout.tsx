"use client";

import { useState } from "react";
import { api } from "@/lib/api";
import { Loader2 } from "lucide-react";

interface CheckoutProps {
    cart: { [key: string]: number };
    totalAmountCents: number;
    onSuccess: (orderId: string) => void;
    onSuccess: (orderId: string) => void;
    location: { lat: number; lng: number } | null;
    storeId: string;
}

export default function Checkout({ cart, totalAmountCents, onSuccess, location, storeId }: CheckoutProps) {
    const [isLoading, setIsLoading] = useState(false);
    const [customerId, setCustomerId] = useState("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"); // Default ID for demo

    const handleCheckout = async () => {
        setIsLoading(true);
        try {
            const items = Object.entries(cart).map(([id, qty]) => {
                // In real app, we look up name from ID. Here we just hardcode or pass name.
                // For simplicity, let's just send basic items.
                return { name: `Product ${id}`, qty };
            });

            // Use passed location or fallback (e.g. store location so it's a 0 distance delivery for safety)
            // But ideally we want user location.
            const deliveryLoc = location || { lat: 0, lng: 0 };

            const res = await api.post("/orders", {
                customer_id: customerId,
                store_id: storeId,
                total_cents: totalAmountCents,
                items: items,
                delivery_lat: deliveryLoc.lat,
                delivery_lng: deliveryLoc.lng,
            });

            onSuccess(res.data.id);
        } catch (error) {
            console.error("Checkout failed", error);
            alert("Checkout Failed");
        } finally {
            setIsLoading(false);
        }
    };

    if (totalAmountCents === 0) {
        return (
            <div className="p-6 bg-zinc-50 dark:bg-zinc-800 rounded-xl text-center text-zinc-500">
                Your cart is empty.
            </div>
        );
    }

    return (
        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 p-6 rounded-xl shadow-lg sticky top-6">
            <h2 className="text-xl font-bold mb-4">Order Summary</h2>

            <div className="space-y-2 mb-6 text-sm text-zinc-600 dark:text-zinc-400">
                {Object.entries(cart).map(([id, qty]) => (
                    <div key={id} className="flex justify-between">
                        <span>Product {id} x {qty}</span>
                        {/* <span>...</span> */}
                    </div>
                ))}
                <div className="border-t border-zinc-200 dark:border-zinc-700 pt-2 flex justify-between font-bold text-zinc-900 dark:text-zinc-100 text-lg">
                    <span>Total</span>
                    <span>${(totalAmountCents / 100).toFixed(2)}</span>
                </div>
            </div>

            <div className="space-y-3 mb-6">
                <div>
                    <label className="text-xs text-zinc-500 block mb-1">Customer ID (Demo)</label>
                    <input
                        type="text"
                        value={customerId}
                        onChange={e => setCustomerId(e.target.value)}
                        className="w-full text-xs p-2 bg-zinc-100 dark:bg-zinc-800 rounded mb-2 border border-zinc-200 dark:border-zinc-700"
                    />
                </div>
            </div>

            <button
                onClick={handleCheckout}
                disabled={isLoading}
                className="w-full py-3 bg-black dark:bg-white text-white dark:text-black rounded-lg font-bold hover:opacity-90 transition flex items-center justify-center gap-2"
            >
                {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                {isLoading ? "Processing..." : "Place Order"}
            </button>
        </div>
    );
}
