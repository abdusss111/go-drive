"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, FolderOpen, BarChart3, LogOut, User } from "lucide-react";
import { useAuthStore } from "@/store/authStore";
import { cn } from "@/lib/utils";

const navigation = [
    { name: "Dashboard", href: "/dashboard", icon: Home },
    { name: "Buckets", href: "/buckets", icon: FolderOpen },
    { name: "Usage", href: "/usage", icon: BarChart3 },
];

export function Sidebar() {
    const pathname = usePathname();
    const { logout } = useAuthStore();

    const handleLogout = () => {
        logout();
        window.location.href = "/login";
    };

    return (
        <div className="flex h-full w-20 flex-col bg-[#2C5282] text-white">
            {/* Logo Icon */}
            <div className="flex h-20 items-center justify-center border-b border-white/10">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-white/10">
                    <span className="text-xl font-bold">G</span>
                </div>
            </div>

            {/* Navigation - Icon Only */}
            <nav className="flex-1 space-y-2 px-4 py-8">
                {navigation.map((item) => {
                    const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
                    return (
                        <Link
                            key={item.name}
                            href={item.href}
                            className={cn(
                                "flex h-12 w-12 items-center justify-center rounded-lg transition-colors",
                                isActive
                                    ? "bg-white/20 text-white"
                                    : "text-white/70 hover:bg-white/10 hover:text-white"
                            )}
                            title={item.name}
                        >
                            <item.icon className="h-6 w-6" />
                        </Link>
                    );
                })}
            </nav>

            {/* User & Logout */}
            <div className="space-y-2 border-t border-white/10 px-4 py-4">
                <button
                    className="flex h-12 w-12 items-center justify-center rounded-lg text-white/70 transition-colors hover:bg-white/10 hover:text-white"
                    title="Profile"
                >
                    <User className="h-6 w-6" />
                </button>
                <button
                    onClick={handleLogout}
                    className="flex h-12 w-12 items-center justify-center rounded-lg text-white/70 transition-colors hover:bg-white/10 hover:text-white"
                    title="Logout"
                >
                    <LogOut className="h-6 w-6" />
                </button>
            </div>
        </div>
    );
}
