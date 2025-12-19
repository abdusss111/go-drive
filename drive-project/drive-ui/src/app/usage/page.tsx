"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { FolderOpen } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ProgressBar } from "@/components/ui/progress-bar";
import { useToast } from "@/components/ui/toast";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/api";
import { formatBytes } from "@/lib/utils";

interface UsageData {
    total_bytes: number;
    quota_bytes: number;
    total_files: number;
    quota_files: number;
    usage_percent_bytes: number;
    usage_percent_files: number;
}

interface Bucket {
    id: string;
    owner_id: string;
    name: string;
    description?: string;
    created_at: string;
    updated_at: string;
    usage: {
        total_bytes: number;
        file_count: number;
    };
}

export default function UsagePage() {
    const router = useRouter();
    const { showToast } = useToast();
    const { isAuthenticated, hasHydrated } = useAuthStore();
    const [usage, setUsage] = useState<UsageData | null>(null);
    const [buckets, setBuckets] = useState<Bucket[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        if (!hasHydrated) return;
        if (!isAuthenticated) {
            router.push("/login");
            return;
        }
        fetchUsageData();
    }, [isAuthenticated, hasHydrated, router]);

    const fetchUsageData = async () => {
        try {
            const [usageRes, bucketsRes] = await Promise.all([
                apiClient.get("/v1/usage"),
                apiClient.get("/v1/buckets"),
            ]);

            setUsage(usageRes.data);
            setBuckets(bucketsRes.data.buckets || []);
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to load usage data", "error");
        } finally {
            setIsLoading(false);
        }
    };

    if (isLoading) {
        return (
            <DashboardLayout>
                <div className="flex h-full items-center justify-center">
                    <p className="text-gray-500">Loading...</p>
                </div>
            </DashboardLayout>
        );
    }

    const bytesPercentage = usage ? usage.usage_percent_bytes : 0;
    const filesPercentage = usage ? usage.usage_percent_files : 0;

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <h1 className="text-3xl font-bold text-gray-900">Usage Statistics</h1>

                {/* Overall Usage */}
                <div className="grid gap-6 md:grid-cols-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>Storage Usage</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {usage && (
                                <>
                                    <div className="flex items-center justify-between">
                                        <div>
                                            <p className="text-3xl font-bold text-gray-900">
                                                {formatBytes(usage.total_bytes)}
                                            </p>
                                            <p className="text-sm text-gray-500">
                                                of {formatBytes(usage.quota_bytes)}
                                            </p>
                                        </div>
                                        <div className="text-right">
                                            <p className="text-2xl font-bold text-gray-900">
                                                {bytesPercentage.toFixed(1)}%
                                            </p>
                                            <p className="text-sm text-gray-500">used</p>
                                        </div>
                                    </div>
                                    <ProgressBar
                                        value={usage.total_bytes}
                                        max={usage.quota_bytes}
                                    />
                                    <div className="flex justify-between text-sm text-gray-600">
                                        <span>
                                            {formatBytes(usage.quota_bytes - usage.total_bytes)} remaining
                                        </span>
                                    </div>
                                </>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Files Count</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {usage && (
                                <>
                                    <div className="flex items-center justify-between">
                                        <div>
                                            <p className="text-3xl font-bold text-gray-900">
                                                {usage.total_files.toLocaleString()}
                                            </p>
                                            <p className="text-sm text-gray-500">
                                                of {usage.quota_files.toLocaleString()} files
                                            </p>
                                        </div>
                                        <div className="text-right">
                                            <p className="text-2xl font-bold text-gray-900">
                                                {filesPercentage.toFixed(1)}%
                                            </p>
                                            <p className="text-sm text-gray-500">used</p>
                                        </div>
                                    </div>
                                    <ProgressBar
                                        value={usage.total_files}
                                        max={usage.quota_files}
                                    />
                                    <div className="flex justify-between text-sm text-gray-600">
                                        <span>
                                            {(usage.quota_files - usage.total_files).toLocaleString()} remaining
                                        </span>
                                    </div>
                                </>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* Bucket Breakdown */}
                <Card>
                    <CardHeader>
                        <CardTitle>Usage by Bucket</CardTitle>
                    </CardHeader>
                    <CardContent>
                        {buckets.length === 0 ? (
                            <p className="py-8 text-center text-gray-500">No buckets yet</p>
                        ) : (
                            <div className="space-y-3">
                                {buckets.map((bucket) => (
                                    <div
                                        key={bucket.id}
                                        className="group cursor-pointer rounded-lg border border-gray-200 bg-white p-4 transition-all hover:border-blue-300 hover:shadow-sm"
                                        onClick={() => router.push(`/buckets/${bucket.id}`)}
                                    >
                                        <div className="mb-3 flex items-center justify-between">
                                            <div className="flex items-center gap-3">
                                                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50">
                                                    <FolderOpen className="h-5 w-5 text-blue-600" />
                                                </div>
                                                <h3 className="font-semibold text-gray-900">{bucket.name}</h3>
                                            </div>
                                            <div className="text-right">
                                                <p className="font-medium text-gray-900">
                                                    {formatBytes(bucket.usage.total_bytes)}
                                                </p>
                                                <p className="text-sm text-gray-500">
                                                    {bucket.usage.file_count} {bucket.usage.file_count === 1 ? "file" : "files"}
                                                </p>
                                            </div>
                                        </div>
                                        {usage && (
                                            <ProgressBar
                                                value={bucket.usage.total_bytes}
                                                max={usage.quota_bytes}
                                                showLabel={false}
                                            />
                                        )}
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>

                {/* Quota Information */}
                <Card>
                    <CardHeader>
                        <CardTitle>Quota Information</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-3 text-sm">
                            <div className="flex justify-between">
                                <span className="text-gray-600">Storage Quota:</span>
                                <span className="font-medium text-gray-900">
                                    {usage && formatBytes(usage.quota_bytes)}
                                </span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-600">Files Quota:</span>
                                <span className="font-medium text-gray-900">
                                    {usage && usage.quota_files.toLocaleString()} files
                                </span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-600">Total Buckets:</span>
                                <span className="font-medium text-gray-900">{buckets.length}</span>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </DashboardLayout>
    );
}
