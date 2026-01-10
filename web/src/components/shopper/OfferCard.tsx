"use client";

import { useEffect, useState } from "react";
import { Timer, MapPin, DollarSign, Package } from "lucide-react";
import { api } from "@/lib/api";
import { cn } from "@/lib/utils";

interface OfferCardProps {
    offer: any; // Ideally typed from backend entity
    onAccept: () => void;
    onDecline: () => void;
}

export default function OfferCard({ offer, onAccept, onDecline }: OfferCardProps) {
    const [timeLeft, setTimeLeft] = useState(60);
    const [isClaiming, setIsClaiming] = useState(false);

    useEffect(() => {
        const timer = setInterval(() => {
            setTimeLeft((prev) => {
                if (prev <= 1) {
                    clearInterval(timer);
                    onDecline(); // Auto-decline
                    return 0;
                }
                return prev - 1;
            });
        }, 1000);
        return () => clearInterval(timer);
    }, [onDecline]);

    const handleAccept = async () => {
        setIsClaiming(true);
        try {
            // Logic for accepting is handled by parent or here. 
            // Usually the API call is: POST /orders/:id/claim
            // But waiting for the parent to pass the "claim" function is cleaner.
            // However, for speed, let's just do it here or assume onAccept does the API call.
            // Let's assume onAccept initiates the API call.
            await onAccept();
        } catch (error) {
            console.error("Failed to claim", error);
            setIsClaiming(false);
        }
    };

    return (
        <div className="fixed bottom-4 left-4 right-4 md:left-auto md:right-8 md:bottom-8 md:w-96 bg-white dark:bg-zinc-900 rounded-xl shadow-2xl border border-zinc-200 dark:border-zinc-800 p-6 z-50 animate-in slide-in-from-bottom duration-300">
            <div className="flex justify-between items-start mb-4">
                <div>
                    <h3 className="text-lg font-bold text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
                        <Package className="w-5 h-5 text-blue-500" />
                        New Delivery Offer
                    </h3>
                    <p className="text-sm text-zinc-500">Order #{offer.payload.order_id.slice(0, 8)}</p>
                </div>
                <div className="flex items-center gap-1 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 px-3 py-1 rounded-full text-sm font-medium">
                    <Timer className="w-4 h-4" />
                    <span>{timeLeft}s</span>
                </div>
            </div>

            <div className="space-y-4 mb-6">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400">
                        <DollarSign className="w-5 h-5" />
                    </div>
                    <div>
                        <p className="text-sm text-zinc-500">Estimated Earnings</p>
                        <p className="font-bold text-lg">$15.50</p>
                    </div>
                </div>

                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400">
                        <MapPin className="w-5 h-5" />
                    </div>
                    <div>
                        <p className="text-sm text-zinc-500">Store Location</p>
                        <p className="font-medium text-sm">Downtown Grocery</p>
                        <p className="text-xs text-zinc-400">1.2 mi away</p>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
                <button
                    onClick={onDecline}
                    disabled={isClaiming}
                    className="px-4 py-3 rounded-lg font-medium text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
                >
                    Decline
                </button>
                <button
                    onClick={handleAccept}
                    disabled={isClaiming}
                    className={cn(
                        "px-4 py-3 rounded-lg font-medium text-white bg-blue-600 hover:bg-blue-700 transition-colors flex items-center justify-center gap-2",
                        isClaiming && "opacity-70 cursor-wait"
                    )}
                >
                    {isClaiming ? "Claiming..." : "Accept Order"}
                </button>
            </div>
        </div>
    );
}
