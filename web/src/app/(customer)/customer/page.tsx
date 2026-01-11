"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { ShoppingBag, MapPin, LogOut } from "lucide-react";

interface Store {
    id: string;
    name: string;
    lat: number;
    lng: number;
}

export default function StoreSelectionPage() {
    const router = useRouter();
    const { user, logout } = useAuth();
    const [stores, setStores] = useState<Store[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        api.get<Store[]>("/stores")
            .then(res => setStores(res.data))
            .catch(err => console.error("Failed to load stores", err))
            .finally(() => setLoading(false));
    }, []);

    return (
        <div className="min-h-screen bg-zinc-50 dark:bg-black">
            <header className="bg-white dark:bg-zinc-900 border-b border-zinc-200 dark:border-zinc-800 p-4 sticky top-0 z-10">
                <div className="max-w-5xl mx-auto flex justify-between items-center">
                    <h1 className="font-bold text-xl flex items-center gap-2">
                        <ShoppingBag className="w-6 h-6 text-blue-600" />
                        Select Store
                    </h1>
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

            <main className="max-w-5xl mx-auto p-8">
                <h2 className="text-2xl font-bold mb-6">Available Stores</h2>

                {loading ? (
                    <div className="flex justify-center p-12">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {stores.map(store => (
                            <div
                                key={store.id}
                                onClick={() => router.push(`/customer/store/${store.id}`)}
                                className="bg-white dark:bg-zinc-900 p-6 rounded-xl border border-zinc-200 dark:border-zinc-800 hover:border-blue-500 cursor-pointer transition shadow-sm hover:shadow-md group"
                            >
                                <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center mb-4 group-hover:bg-blue-600 group-hover:text-white transition-colors text-blue-600">
                                    <MapPin className="w-6 h-6" />
                                </div>
                                <h3 className="font-bold text-lg mb-1">{store.name}</h3>
                                <p className="text-sm text-zinc-500">Open now • 0.5 mi</p>
                            </div>
                        ))}
                    </div>
                )}
            </main>
        </div>
    );
}
