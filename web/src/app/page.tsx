import Link from "next/link";

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-8 p-24 bg-zinc-50 dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100">
      <h1 className="text-4xl font-bold tracking-tight">Order Tracker</h1>
      <div className="flex gap-4">
        <Link
          href="/customer"
          className="rounded-lg bg-blue-600 px-6 py-3 font-medium text-white transition hover:bg-blue-700 shadow-lg shadow-blue-600/20"
        >
          Customer App
        </Link>
        <Link
          href="/courier"
          className="rounded-lg bg-emerald-600 px-6 py-3 font-medium text-white transition hover:bg-emerald-700 shadow-lg shadow-emerald-600/20"
        >
          Shopper App
        </Link>
      </div>
    </main>
  );
}
