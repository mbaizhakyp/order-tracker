"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useGeolocation } from "@/hooks/useGeolocation";
import { api } from "@/lib/api";
import ProductCatalog, { PRODUCTS } from "@/components/customer/ProductCatalog";
import Checkout from "@/components/customer/Checkout";
import { ShoppingCart } from "lucide-react";

export default function CustomerPage() {
    const router = useRouter();

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
                    <h1 className="font-bold text-xl flex items-center gap-2">
                        <ShoppingCart className="w-6 h-6 text-blue-600" />
                        FreshMart
                    </h1>
                    <div className="text-sm font-medium">
                        Customer App
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
                    />
                </aside>
            </main>
        </div>
    );
}
