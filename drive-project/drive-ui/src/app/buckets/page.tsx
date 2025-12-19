"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Plus, FolderOpen, Trash2, Search } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Modal } from "@/components/ui/modal";
import { useToast } from "@/components/ui/toast";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/api";
import { formatBytes, formatDate } from "@/lib/utils";

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

export default function BucketsPage() {
    const router = useRouter();
    const { showToast } = useToast();
    const { isAuthenticated, hasHydrated } = useAuthStore();
    const [buckets, setBuckets] = useState<Bucket[]>([]);
    const [filteredBuckets, setFilteredBuckets] = useState<Bucket[]>([]);
    const [searchQuery, setSearchQuery] = useState("");
    const [isLoading, setIsLoading] = useState(true);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [newBucketName, setNewBucketName] = useState("");
    const [newBucketDescription, setNewBucketDescription] = useState("");
    const [isCreating, setIsCreating] = useState(false);

    useEffect(() => {
        if (!hasHydrated) return;
        if (!isAuthenticated) {
            router.push("/login");
            return;
        }
        fetchBuckets();
    }, [isAuthenticated, hasHydrated, router]);

    useEffect(() => {
        if (searchQuery) {
            setFilteredBuckets(
                buckets.filter((bucket) =>
                    bucket.name.toLowerCase().includes(searchQuery.toLowerCase())
                )
            );
        } else {
            setFilteredBuckets(buckets);
        }
    }, [searchQuery, buckets]);

    const fetchBuckets = async () => {
        try {
            const response = await apiClient.get("/v1/buckets");
            setBuckets(response.data.buckets || []);
            setFilteredBuckets(response.data.buckets || []);
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to load buckets", "error");
        } finally {
            setIsLoading(false);
        }
    };

    const handleCreateBucket = async () => {
        if (!newBucketName.trim()) {
            showToast("Bucket name is required", "error");
            return;
        }

        setIsCreating(true);
        try {
            await apiClient.post("/v1/buckets", {
                name: newBucketName,
                description: newBucketDescription || undefined,
            });

            showToast("Bucket created successfully", "success");
            setIsCreateModalOpen(false);
            setNewBucketName("");
            setNewBucketDescription("");
            fetchBuckets();
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to create bucket", "error");
        } finally {
            setIsCreating(false);
        }
    };

    const handleDeleteBucket = async (bucketId: string, bucketName: string) => {
        if (!confirm(`Are you sure you want to delete "${bucketName}"? This action cannot be undone.`)) {
            return;
        }

        try {
            await apiClient.delete(`/v1/buckets/${bucketId}`);
            showToast("Bucket deleted successfully", "success");
            fetchBuckets();
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to delete bucket", "error");
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
                    <h1 className="text-3xl font-bold text-gray-900">Buckets</h1>
                    <Button onClick={() => setIsCreateModalOpen(true)}>
                        <Plus className="mr-2 h-4 w-4" />
                        Create Bucket
                    </Button>
                </div>

                {/* Search */}
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                    <Input
                        type="search"
                        placeholder="Search buckets..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-10"
                    />
                </div>

                {/* Buckets Grid */}
                {filteredBuckets.length === 0 ? (
                    <Card>
                        <CardContent className="flex flex-col items-center justify-center py-12">
                            <FolderOpen className="mb-4 h-12 w-12 text-gray-400" />
                            <h3 className="mb-2 text-lg font-medium text-gray-900">
                                {searchQuery ? "No buckets found" : "No buckets yet"}
                            </h3>
                            <p className="mb-4 text-sm text-gray-500">
                                {searchQuery
                                    ? "Try a different search term"
                                    : "Create your first bucket to start uploading files"}
                            </p>
                            {!searchQuery && (
                                <Button onClick={() => setIsCreateModalOpen(true)}>
                                    <Plus className="mr-2 h-4 w-4" />
                                    Create Bucket
                                </Button>
                            )}
                        </CardContent>
                    </Card>
                ) : (
                    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                        {filteredBuckets.map((bucket) => (
                            <Card
                                key={bucket.id}
                                className="transition-all hover:border-blue-300 hover:shadow-md"
                            >
                                <CardHeader>
                                    <CardTitle className="flex items-center justify-between">
                                        <div
                                            className="flex cursor-pointer items-center gap-2"
                                            onClick={() => router.push(`/buckets/${bucket.id}`)}
                                        >
                                            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50">
                                                <FolderOpen className="h-5 w-5 text-blue-600" />
                                            </div>
                                            <span className="truncate">{bucket.name}</span>
                                        </div>
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDeleteBucket(bucket.id, bucket.name);
                                            }}
                                            className="rounded-md p-1.5 hover:bg-red-50"
                                        >
                                            <Trash2 className="h-4 w-4 text-red-600" />
                                        </button>
                                    </CardTitle>
                                </CardHeader>
                                <CardContent
                                    className="cursor-pointer"
                                    onClick={() => router.push(`/buckets/${bucket.id}`)}
                                >
                                    {bucket.description && (
                                        <p className="mb-3 text-sm text-gray-600">{bucket.description}</p>
                                    )}
                                    <div className="space-y-2 text-sm text-gray-600">
                                        <div className="flex justify-between">
                                            <span>Files:</span>
                                            <span className="font-medium text-gray-900">{bucket.usage.file_count}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Size:</span>
                                            <span className="font-medium text-gray-900">{formatBytes(bucket.usage.total_bytes)}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span>Created:</span>
                                            <span className="font-medium text-gray-900">{formatDate(bucket.created_at)}</span>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                )}
            </div>

            {/* Create Bucket Modal */}
            <Modal
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
                title="Create New Bucket"
            >
                <div className="space-y-4">
                    <div>
                        <label htmlFor="bucketName" className="block text-sm font-medium text-gray-700">
                            Bucket Name *
                        </label>
                        <Input
                            id="bucketName"
                            type="text"
                            placeholder="my-bucket"
                            value={newBucketName}
                            onChange={(e) => setNewBucketName(e.target.value)}
                            className="mt-1"
                        />
                    </div>

                    <div>
                        <label htmlFor="bucketDescription" className="block text-sm font-medium text-gray-700">
                            Description (Optional)
                        </label>
                        <Input
                            id="bucketDescription"
                            type="text"
                            placeholder="Bucket description"
                            value={newBucketDescription}
                            onChange={(e) => setNewBucketDescription(e.target.value)}
                            maxLength={255}
                            className="mt-1"
                        />
                        <p className="mt-1 text-xs text-gray-500">
                            {newBucketDescription.length}/255 characters
                        </p>
                    </div>

                    <div className="flex gap-2">
                        <Button
                            onClick={handleCreateBucket}
                            disabled={isCreating || !newBucketName.trim()}
                            className="flex-1"
                        >
                            {isCreating ? "Creating..." : "Create"}
                        </Button>
                        <Button
                            variant="outline"
                            onClick={() => setIsCreateModalOpen(false)}
                            className="flex-1"
                        >
                            Cancel
                        </Button>
                    </div>
                </div>
            </Modal>
        </DashboardLayout>
    );
}
