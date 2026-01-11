"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { Loader2 } from "lucide-react";

export default function LoginPage() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const { login } = useAuth();

    // Check if redirected from register
    const justRegistered = searchParams.get("registered") === "true";

    const [form, setForm] = useState({
        email: "",
        password: "",
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
            const res = await api.post("/auth/login", form);
            const { token, user } = res.data;

            // Update Context & Storage
            login(token, user);

            // Redirect based on role
            if (user.role === "SHOPPER") {
                router.push("/courier");
            } else {
                router.push("/customer");
            }

        } catch (err: any) {
            console.error("Login failed", err);
            setError(err.response?.data?.error || "Invalid email or password.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-black p-4">
            <div className="w-full max-w-md bg-white dark:bg-zinc-900 p-8 rounded-xl shadow-lg border border-zinc-200 dark:border-zinc-800">
                <div className="text-center mb-8">
                    <h1 className="text-2xl font-bold">Welcome Back</h1>
                    <p className="text-sm text-zinc-500 mt-2">Log in to your account</p>
                </div>

                {justRegistered && (
                    <div className="mb-6 p-3 bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400 text-sm rounded-lg text-center">
                        Account created! Please log in.
                    </div>
                )}

                {error && (
                    <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 text-sm rounded-lg">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
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
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full py-3 bg-blue-600 text-white rounded-lg font-bold hover:bg-blue-700 transition disabled:opacity-50 flex items-center justify-center gap-2"
                    >
                        {loading && <Loader2 className="w-5 h-5 animate-spin" />}
                        Log In
                    </button>

                    {/* Demo Credentials Hint */}
                    <div className="mt-4 p-3 bg-zinc-100 dark:bg-zinc-800 rounded text-xs text-zinc-500">
                        <p className="font-bold mb-1">Demo Credentials:</p>
                        <p>User: user@example.com / password123</p>
                        <p>Shopper: shopper@example.com / password123</p>
                    </div>
                </form>

                <p className="mt-6 text-center text-sm text-zinc-500">
                    Don't have an account?{" "}
                    <Link href="/register" className="text-blue-600 hover:underline">
                        Sign up
                    </Link>
                </p>
            </div>
        </div>
    );
}
