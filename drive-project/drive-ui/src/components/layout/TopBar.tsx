"use client";

import { Bell, Search } from "lucide-react";
import { Input } from "@/components/ui/input";

export function TopBar() {
    return (
        <div className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
            {/* Search */}
            <div className="flex flex-1 items-center gap-4">
                <div className="relative w-96">
                    <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <Input
                        type="search"
                        placeholder="Search files and buckets..."
                        className="pl-10"
                    />
                </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-4">
                <button className="rounded-full p-2 hover:bg-gray-100">
                    <Bell className="h-5 w-5 text-gray-600" />
                </button>
            </div>
        </div>
    );
}
