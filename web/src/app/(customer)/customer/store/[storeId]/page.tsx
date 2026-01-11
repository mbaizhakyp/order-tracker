"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { useGeolocation } from "@/hooks/useGeolocation";
import { api } from "@/lib/api";
import ProductCatalog, { PRODUCTS } from "@/components/customer/ProductCatalog";
import Checkout from "@/components/customer/Checkout";
import { ShoppingCart, LogOut, ArrowLeft } from "lucide-react";
import { useAuth } from "@/context/AuthContext";

export default function CustomerPage() {
    const router = useRouter();
    const params = useParams();
    const storeId = params.storeId as string;
    const { user, logout } = useAuth();

    // 1. Get User Location
    const { location } = useGeolocation();

    // 2. Init Demo on Location Found
    useEffect(() => {
        if (location) {
            console.log("Setting Demo Location:", location);
            api.post("/demo/location", location)
                .catch(err => console.error("Failed to set demo location", err));
        }
    }, [location]);

    const [cart, setCart] = useState<{ [key: string]: number }>({});

    const totalVal = Object.entries(cart).reduce((acc, [id, qty]) => {
        const product = PRODUCTS.find(p => p.id === id);
        return acc + (product ? product.priceCents * qty : 0);
    }, 0);

    const handleOrderSuccess = (orderId: string) => {
        setCart({});
        router.push(`/customer/track/${orderId}`);
    };

    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 p-4 sticky top-0 z-10">
                <div className="max-w-5xl mx-auto flex justify-between items-center">
                    <div className="flex items-center gap-4">
                        <button
                            onClick={() => router.push("/customer")}
                            className="p-2 -ml-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-full transition"
                            title="Back to Stores"
                        >
                            <ArrowLeft className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
                        </button>
                        <h1 className="font-bold text-xl flex items-center gap-2">
                            <ShoppingCart className="w-6 h-6 text-blue-600" />
                            FreshMart
                        </h1>
                    </div>
                    <div className="flex items-center gap-4 text-sm font-medium">
                        {user ? (
                            <>
                                <span className="text-zinc-600 dark:text-zinc-400">Hello, {user.name}</span>
                                <button
                                    onClick={() => router.push("/customer/orders")}
                                    className="text-zinc-600 hover:text-blue-600 transition"
                                >
                                    My Orders
                                </button>
                                <button
                                    onClick={() => { logout(); router.push("/login"); }}
                                    className="flex items-center gap-1 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 px-3 py-1.5 rounded-lg transition"
                                >
                                    <LogOut className="w-4 h-4" />
                                    Log Out
                                </button>
                            </>
                        ) : (
                            <span onClick={() => router.push("/login")} className="cursor-pointer hover:underline text-blue-600">
                                Log In
                            </span>
                        )}
                    </div>
                </div>
            </header>

            <main className="max-w-5xl mx-auto p-4 md:p-8 grid grid-cols-1 md:grid-cols-3 gap-8">
                <div className="md:col-span-2 space-y-6">
                    <section>
                        <h2 className="text-2xl font-bold mb-4">Shop Groceries</h2>
                        <ProductCatalog cart={cart} onUpdate={(id, delta) => {
                            setCart(prev => {
                                const current = prev[id] || 0;
                                const next = Math.max(0, current + delta);
                                return { ...prev, [id]: next };
                            });
                        }} />
                    </section>
                </div>

                <aside className="md:col-span-1">
                    <Checkout
                        cart={cart}
                        totalAmountCents={totalVal}
                        onSuccess={handleOrderSuccess}
                        location={location}
                        storeId={storeId}
                    />
                </aside>
            </main>
        </div>
    );
}
