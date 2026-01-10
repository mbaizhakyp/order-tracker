"use client";

import { useState } from "react";
import { Plus, Minus, ShoppingCart } from "lucide-react";

export interface Product {
    id: string;
    name: string;
    priceCents: number;
    image?: string;
}

export const PRODUCTS: Product[] = [
    { id: "1", name: "Organic Milk", priceCents: 550 },
    { id: "2", name: "Sourdough Bread", priceCents: 450 },
    { id: "3", name: "Apples (Bag)", priceCents: 600 },
    { id: "4", name: "Eggs (Dozen)", priceCents: 350 },
];

// ... imports

interface ProductCatalogProps {
    cart: { [key: string]: number };
    onUpdate: (id: string, delta: number) => void;
}

export default function ProductCatalog({ cart, onUpdate }: ProductCatalogProps) {
    return (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {PRODUCTS.map(product => {
                const qty = cart[product.id] || 0;
                return (
                    <div key={product.id} className="bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 p-4 rounded-xl shadow-sm flex justify-between items-center">
                        <div>
                            <h3 className="font-semibold text-lg">{product.name}</h3>
                            <p className="text-sm text-zinc-500">${(product.priceCents / 100).toFixed(2)}</p>
                        </div>
                        <div className="flex items-center gap-3 bg-zinc-100 dark:bg-zinc-800 rounded-lg p-1">
                            <button
                                onClick={() => onUpdate(product.id, -1)}
                                className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-white dark:hover:bg-zinc-700 transition"
                                disabled={qty === 0}
                            >
                                <Minus className="w-4 h-4" />
                            </button>
                            <span className="font-medium w-4 text-center">{qty}</span>
                            <button
                                onClick={() => onUpdate(product.id, 1)}
                                className="w-8 h-8 flex items-center justify-center rounded-md hover:bg-white dark:hover:bg-zinc-700 transition"
                            >
                                <Plus className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
