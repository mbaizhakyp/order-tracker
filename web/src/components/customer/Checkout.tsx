"use client";

import { useState } from "react";
import { api } from "@/lib/api";
import { Loader2 } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { useRouter } from "next/navigation";

interface CheckoutProps {
    cart: { [key: string]: number };
    totalAmountCents: number;
    onSuccess: (orderId: string) => void;
    onSuccess: (orderId: string) => void;
    location: { lat: number; lng: number } | null;
    storeId: string;
}

export default function Checkout({ cart, totalAmountCents, onSuccess, location, storeId }: CheckoutProps) {
    const { user } = useAuth();
    const router = useRouter();
    const [isLoading, setIsLoading] = useState(false);

    const handleCheckout = async () => {
        if (!user) {
            router.push("/login?redirect=/customer"); // Simple redirect logic
            return;
        }

        setIsLoading(true);
        try {
            const items = Object.entries(cart).map(([id, qty]) => {
                return { name: `Product ${id}`, qty };
            });

            const deliveryLoc = location || { lat: 0, lng: 0 };

            // customer_id is now handled by backend from token
            const res = await api.post("/orders", {
                store_id: storeId,
                total_cents: totalAmountCents,
                items: items,
                delivery_lat: deliveryLoc.lat,
                delivery_lng: deliveryLoc.lng,
            });

            onSuccess(res.data.id);
        } catch (error) {
            console.error("Checkout failed", error);
            alert("Checkout Failed: Please ensure you are logged in.");
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

    if (!user) {
        return (
            <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 p-6 rounded-xl shadow-lg sticky top-6 text-center">
                <h2 className="text-xl font-bold mb-4">Ready to Order?</h2>
                <p className="text-zinc-500 mb-6">Please log in to place your order.</p>
                <button
                    onClick={() => router.push("/login")}
                    className="w-full py-3 bg-blue-600 text-white rounded-lg font-bold hover:bg-blue-700 transition"
                >
                    Log In / Register
                </button>
            </div>
        )
    }

    return (
        <div className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 p-6 rounded-xl shadow-lg sticky top-6">
            <h2 className="text-xl font-bold mb-4">Order Summary</h2>

            <div className="space-y-2 mb-6 text-sm text-zinc-600 dark:text-zinc-400">
                {Object.entries(cart).map(([id, qty]) => (
                    <div key={id} className="flex justify-between">
                        <span>Product {id} x {qty}</span>
                    </div>
                ))}
                <div className="border-t border-zinc-200 dark:border-zinc-700 pt-2 flex justify-between font-bold text-zinc-900 dark:text-zinc-100 text-lg">
                    <span>Total</span>
                    <span>${(totalAmountCents / 100).toFixed(2)}</span>
                </div>
            </div>

            <div className="mb-6 p-3 bg-zinc-50 dark:bg-zinc-800 rounded text-sm text-zinc-600 dark:text-zinc-400">
                Ordering as <span className="font-bold text-zinc-900 dark:text-zinc-100">{user.email}</span>
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
