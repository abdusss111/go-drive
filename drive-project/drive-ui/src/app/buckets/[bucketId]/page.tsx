"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { ArrowLeft, Trash2 } from "lucide-react";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { Input } from "@/components/ui/input";
import { useToast } from "@/components/ui/toast";
import { useAuthStore } from "@/store/authStore";
import { FileUploadZone } from "@/components/files/FileUploadZone";
import { FileList } from "@/components/files/FileList";
import apiClient from "@/lib/api";
import { formatBytes } from "@/lib/utils";

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

interface FileData {
    id: string;
    bucket_id: string;
    object_name: string;
    original_filename: string;
    size_bytes: number;
    content_type: string;
    checksum: string;
    created_at: string;
    updated_at: string;
}

export default function BucketDetailPage() {
    const router = useRouter();
    const params = useParams();
    const bucketId = params.bucketId as string;
    const { showToast } = useToast();
    const { isAuthenticated, hasHydrated } = useAuthStore();

    const [bucket, setBucket] = useState<Bucket | null>(null);
    const [files, setFiles] = useState<FileData[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [isUploading, setIsUploading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState<Record<string, number>>({});
    const [isShareModalOpen, setIsShareModalOpen] = useState(false);
    const [shareUrl, setShareUrl] = useState("");
    const [shareTTL, setShareTTL] = useState(3600);

    useEffect(() => {
        if (!hasHydrated) return;
        if (!isAuthenticated) {
            router.push("/login");
            return;
        }
        fetchBucketData();
    }, [isAuthenticated, hasHydrated, bucketId, router]);

    const fetchBucketData = async () => {
        try {
            const [bucketRes, filesRes] = await Promise.all([
                apiClient.get(`/v1/buckets/${bucketId}`),
                apiClient.get(`/v1/buckets/${bucketId}/files`),
            ]);

            setBucket(bucketRes.data);
            setFiles(filesRes.data.files || []);
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to load bucket", "error");
            if (error.response?.status === 404) {
                router.push("/buckets");
            }
        } finally {
            setIsLoading(false);
        }
    };

    const handleFilesSelected = async (selectedFiles: File[]) => {
        setIsUploading(true);

        for (const file of selectedFiles) {
            const formData = new FormData();
            formData.append("file", file);

            try {
                await apiClient.post(`/v1/buckets/${bucketId}/files`, formData, {
                    headers: {
                        "Content-Type": "multipart/form-data",
                    },
                    onUploadProgress: (progressEvent) => {
                        const progress = progressEvent.total
                            ? Math.round((progressEvent.loaded * 100) / progressEvent.total)
                            : 0;
                        setUploadProgress((prev) => ({ ...prev, [file.name]: progress }));
                    },
                });

                showToast(`${file.name} uploaded successfully`, "success");
            } catch (error: any) {
                const message = error.response?.data?.error || `Failed to upload ${file.name}`;
                showToast(message, "error");
            }
        }

        setIsUploading(false);
        setUploadProgress({});
        fetchBucketData();
    };

    const handleDownload = async (fileId: string, filename: string) => {
        try {
            const response = await apiClient.get(
                `/v1/buckets/${bucketId}/files/${fileId}/download`,
                { responseType: "blob" }
            );

            // Preserve the Content-Type from the backend response
            const contentType = response.headers['content-type'] || 'application/octet-stream';
            const blob = new Blob([response.data], { type: contentType });
            const url = window.URL.createObjectURL(blob);

            const link = document.createElement("a");
            link.href = url;
            link.setAttribute("download", filename);
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);

            showToast("File downloaded successfully", "success");
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to download file", "error");
        }
    };

    const handleDelete = async (fileId: string, filename: string) => {
        if (!confirm(`Are you sure you want to delete "${filename}"?`)) {
            return;
        }

        try {
            await apiClient.delete(`/v1/buckets/${bucketId}/files/${fileId}`);
            showToast("File deleted successfully", "success");
            fetchBucketData();
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to delete file", "error");
        }
    };

    const handleShare = async (fileId: string) => {
        try {
            const response = await apiClient.post(
                `/v1/buckets/${bucketId}/files/${fileId}/presigned-url`,
                {
                    method: "GET",
                    ttl: `${shareTTL}s`, // Convert seconds to Go duration format
                }
            );

            setShareUrl(response.data.url);
            setIsShareModalOpen(true);
        } catch (error: any) {
            showToast(error.response?.data?.error || "Failed to generate share link", "error");
        }
    };

    const handleCopyShareUrl = () => {
        navigator.clipboard.writeText(shareUrl);
        showToast("Link copied to clipboard", "success");
    };

    const handleDeleteBucket = async () => {
        if (!confirm(`Are you sure you want to delete "${bucket?.name}"? This will delete all files in the bucket.`)) {
            return;
        }

        try {
            await apiClient.delete(`/v1/buckets/${bucketId}`);
            showToast("Bucket deleted successfully", "success");
            router.push("/buckets");
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

    if (!bucket) {
        return null;
    }

    return (
        <DashboardLayout>
            <div className="space-y-6">
                {/* Header */}
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="icon" onClick={() => router.push("/buckets")}>
                            <ArrowLeft className="h-5 w-5" />
                        </Button>
                        <div>
                            <h1 className="text-3xl font-bold text-gray-900">{bucket.name}</h1>
                            {bucket.description && (
                                <p className="text-gray-600">{bucket.description}</p>
                            )}
                        </div>
                    </div>
                    <Button variant="destructive" onClick={handleDeleteBucket}>
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete Bucket
                    </Button>
                </div>

                {/* Stats */}
                <div className="grid gap-4 md:grid-cols-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>Files</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-3xl font-bold">{bucket.usage.file_count}</p>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle>Total Size</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <p className="text-3xl font-bold">{formatBytes(bucket.usage.total_bytes)}</p>
                        </CardContent>
                    </Card>
                </div>

                {/* Upload Zone */}
                <FileUploadZone onFilesSelected={handleFilesSelected} disabled={isUploading} />

                {/* Upload Progress */}
                {Object.keys(uploadProgress).length > 0 && (
                    <Card>
                        <CardHeader>
                            <CardTitle>Uploading...</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-2">
                            {Object.entries(uploadProgress).map(([filename, progress]) => (
                                <div key={filename}>
                                    <div className="mb-1 flex justify-between text-sm">
                                        <span>{filename}</span>
                                        <span>{progress}%</span>
                                    </div>
                                    <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                                        <div
                                            className="h-full bg-blue-600 transition-all"
                                            style={{ width: `${progress}%` }}
                                        />
                                    </div>
                                </div>
                            ))}
                        </CardContent>
                    </Card>
                )}

                {/* Files List */}
                <div>
                    <h2 className="mb-4 text-xl font-semibold text-gray-900">Files</h2>
                    <FileList
                        files={files}
                        onDownload={handleDownload}
                        onDelete={handleDelete}
                        onShare={handleShare}
                    />
                </div>
            </div>

            {/* Share Modal */}
            <Modal
                isOpen={isShareModalOpen}
                onClose={() => setIsShareModalOpen(false)}
                title="Share File"
            >
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700">
                            Shareable Link
                        </label>
                        <div className="mt-1 flex gap-2">
                            <Input value={shareUrl} readOnly className="flex-1" />
                            <Button onClick={handleCopyShareUrl}>Copy</Button>
                        </div>
                    </div>
                    <p className="text-sm text-gray-500">
                        This link will expire in {shareTTL / 3600} hour(s)
                    </p>
                </div>
            </Modal>
        </DashboardLayout>
    );
}
