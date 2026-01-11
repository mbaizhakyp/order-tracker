"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { Loader2 } from "lucide-react";

export default function RegisterPage() {
    const router = useRouter();
    const { login } = useAuth(); // We can force login after register if we want, or just redirect to login

    const [form, setForm] = useState({
        name: "",
        email: "",
        password: "",
        role: "CUSTOMER"
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setForm({ ...form, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        try {
            const res = await api.post("/auth/register", form);
            // Registration success.
            // Option A: Automatically login (needs backend to return token on register, or frontend to call login immediately)
            // Option B: Redirect to Login.
            // Let's do Option B for simplicity and clarity.
            router.push("/login?registered=true");
        } catch (err: any) {
            console.error("Registration failed", err);
            setError(err.response?.data?.error || "Registration failed. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-black p-4">
            <div className="w-full max-w-md bg-white dark:bg-zinc-900 p-8 rounded-xl shadow-lg border border-zinc-200 dark:border-zinc-800">
                <div className="text-center mb-8">
                    <h1 className="text-2xl font-bold">Create Account</h1>
                    <p className="text-sm text-zinc-500 mt-2">Join OrderTracker today</p>
                </div>

                {error && (
                    <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 text-sm rounded-lg">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium mb-1">Full Name</label>
                        <input
                            name="name"
                            value={form.name}
                            onChange={handleChange}
                            className="w-full px-4 py-2 rounded-lg border border-zinc-300 dark:border-zinc-700 bg-transparent"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1">Email</label>
                        <input
                            name="email"
                            type="email"
                            value={form.email}
                            onChange={handleChange}
                            className="w-full px-4 py-2 rounded-lg border border-zinc-300 dark:border-zinc-700 bg-transparent"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1">Password</label>
                        <input
                            name="password"
                            type="password"
                            value={form.password}
                            onChange={handleChange}
                            className="w-full px-4 py-2 rounded-lg border border-zinc-300 dark:border-zinc-700 bg-transparent"
                            required
                            minLength={6}
                        />
                    </div>

                    <div className="pt-2">
                        <label className="block text-sm font-medium mb-2">I am a...</label>
                        <div className="grid grid-cols-2 gap-4">
                            <button
                                type="button"
                                onClick={() => setForm({ ...form, role: "CUSTOMER" })}
                                className={`py-2 rounded-lg border font-medium transition ${form.role === "CUSTOMER"
                                        ? "bg-blue-50 dark:bg-blue-900/20 border-blue-500 text-blue-600"
                                        : "border-zinc-300 dark:border-zinc-700 text-zinc-500 hover:bg-zinc-50 dark:hover:bg-zinc-800"
                                    }`}
                            >
                                Customer
                            </button>
                            <button
                                type="button"
                                onClick={() => setForm({ ...form, role: "SHOPPER" })}
                                className={`py-2 rounded-lg border font-medium transition ${form.role === "SHOPPER"
                                        ? "bg-blue-50 dark:bg-blue-900/20 border-blue-500 text-blue-600"
                                        : "border-zinc-300 dark:border-zinc-700 text-zinc-500 hover:bg-zinc-50 dark:hover:bg-zinc-800"
                                    }`}
                            >
                                Shopper
                            </button>
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full py-3 bg-blue-600 text-white rounded-lg font-bold hover:bg-blue-700 transition disabled:opacity-50 flex items-center justify-center gap-2"
                    >
                        {loading && <Loader2 className="w-5 h-5 animate-spin" />}
                        Sign Up
                    </button>
                </form>

                <p className="mt-6 text-center text-sm text-zinc-500">
                    Already have an account?{" "}
                    <Link href="/login" className="text-blue-600 hover:underline">
                        Log in
                    </Link>
                </p>
            </div>
        </div>
    );
}
