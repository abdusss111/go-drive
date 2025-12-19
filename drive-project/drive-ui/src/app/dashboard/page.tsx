"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Plus, FolderOpen, FileText, Table, Presentation, Video } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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

export default function DashboardPage() {
    const router = useRouter();
    const { showToast } = useToast();
    const { isAuthenticated, hasHydrated } = useAuthStore();
    const [usage, setUsage] = useState<UsageData | null>(null);
    const [buckets, setBuckets] = useState<Bucket[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        // Wait for Zustand to hydrate from localStorage
        if (!hasHydrated) {
            return;
        }

        if (!isAuthenticated) {
            router.push("/login");
            return;
        }

        fetchData();
    }, [isAuthenticated, hasHydrated, router]);

    const fetchData = async () => {
        try {
            const [usageRes, bucketsRes] = await Promise.all([
                apiClient.get("/v1/usage"),
                apiClient.get("/v1/buckets"),
            ]);

            setUsage(usageRes.data);
            setBuckets(bucketsRes.data.buckets || []);
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to load data", "error");
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

    return (
        <DashboardLayout>
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
                    <Button onClick={() => router.push("/buckets")}>
                        <Plus className="mr-2 h-4 w-4" />
                        New Bucket
                    </Button>
                </div>

                {/* Category Cards */}
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                    <div className="group cursor-pointer rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 p-6 text-white shadow-lg transition-all hover:shadow-xl">
                        <div className="flex items-center justify-between">
                            <div>
                                <FileText className="mb-3 h-8 w-8" />
                                <h3 className="text-lg font-semibold">Documents</h3>
                                <p className="mt-1 text-sm text-purple-100">PDF, DOC, TXT</p>
                            </div>
                        </div>
                    </div>

                    <div className="group cursor-pointer rounded-xl bg-gradient-to-br from-teal-500 to-teal-600 p-6 text-white shadow-lg transition-all hover:shadow-xl">
                        <div className="flex items-center justify-between">
                            <div>
                                <Table className="mb-3 h-8 w-8" />
                                <h3 className="text-lg font-semibold">Spreadsheet</h3>
                                <p className="mt-1 text-sm text-teal-100">XLS, CSV</p>
                            </div>
                        </div>
                    </div>

                    <div className="group cursor-pointer rounded-xl bg-gradient-to-br from-pink-500 to-pink-600 p-6 text-white shadow-lg transition-all hover:shadow-xl">
                        <div className="flex items-center justify-between">
                            <div>
                                <Presentation className="mb-3 h-8 w-8" />
                                <h3 className="text-lg font-semibold">Presentation</h3>
                                <p className="mt-1 text-sm text-pink-100">PPT, KEY</p>
                            </div>
                        </div>
                    </div>

                    <div className="group cursor-pointer rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 p-6 text-white shadow-lg transition-all hover:shadow-xl">
                        <div className="flex items-center justify-between">
                            <div>
                                <Video className="mb-3 h-8 w-8" />
                                <h3 className="text-lg font-semibold">Video</h3>
                                <p className="mt-1 text-sm text-blue-100">MP4, AVI, MOV</p>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Usage Stats */}
                <div className="grid gap-6 md:grid-cols-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>Storage Usage</CardTitle>
                        </CardHeader>
                        <CardContent>
                            {usage && (
                                <>
                                    <div className="mb-2 flex items-center justify-between text-sm">
                                        <span className="text-gray-600">
                                            {formatBytes(usage.total_bytes)} / {formatBytes(usage.quota_bytes)}
                                        </span>
                                        <span className="font-medium">
                                            {usage.usage_percent_bytes.toFixed(1)}%
                                        </span>
                                    </div>
                                    <ProgressBar
                                        value={usage.total_bytes}
                                        max={usage.quota_bytes}
                                    />
                                </>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Files Count</CardTitle>
                        </CardHeader>
                        <CardContent>
                            {usage && (
                                <>
                                    <div className="mb-2 flex items-center justify-between text-sm">
                                        <span className="text-gray-600">
                                            {usage.total_files.toLocaleString()} / {usage.quota_files.toLocaleString()}
                                        </span>
                                        <span className="font-medium">
                                            {usage.usage_percent_files.toFixed(1)}%
                                        </span>
                                    </div>
                                    <ProgressBar
                                        value={usage.total_files}
                                        max={usage.quota_files}
                                    />
                                </>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* Buckets List */}
                <div>
                    <h2 className="mb-4 text-xl font-semibold text-gray-900">Your Buckets</h2>
                    {buckets.length === 0 ? (
                        <Card>
                            <CardContent className="flex flex-col items-center justify-center py-12">
                                <FolderOpen className="mb-4 h-12 w-12 text-gray-400" />
                                <h3 className="mb-2 text-lg font-medium text-gray-900">No buckets yet</h3>
                                <p className="mb-4 text-sm text-gray-500">
                                    Create your first bucket to start uploading files
                                </p>
                                <Button onClick={() => router.push("/buckets")}>
                                    <Plus className="mr-2 h-4 w-4" />
                                    Create Bucket
                                </Button>
                            </CardContent>
                        </Card>
                    ) : (
                        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                            {buckets.slice(0, 6).map((bucket) => (
                                <Card
                                    key={bucket.id}
                                    className="cursor-pointer transition-all hover:border-blue-300 hover:shadow-md"
                                    onClick={() => router.push(`/buckets/${bucket.id}`)}
                                >
                                    <CardHeader>
                                        <CardTitle className="flex items-center gap-2">
                                            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50">
                                                <FolderOpen className="h-5 w-5 text-blue-600" />
                                            </div>
                                            {bucket.name}
                                        </CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <div className="space-y-2 text-sm text-gray-600">
                                            <div className="flex justify-between">
                                                <span>Files:</span>
                                                <span className="font-medium text-gray-900">{bucket.usage.file_count}</span>
                                            </div>
                                            <div className="flex justify-between">
                                                <span>Size:</span>
                                                <span className="font-medium text-gray-900">{formatBytes(bucket.usage.total_bytes)}</span>
                                            </div>
                                        </div>
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    )}
                    {buckets.length > 6 && (
                        <div className="mt-4 text-center">
                            <Button variant="outline" onClick={() => router.push("/buckets")}>
                                View All Buckets
                            </Button>
                        </div>
                    )}
                </div>
            </div>
        </DashboardLayout>
    );
}
