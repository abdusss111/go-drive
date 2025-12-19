"use client";

import {
    Download,
    Trash2,
    Share2,
    FileIcon,
    Image,
    Video,
    Music,
    FileText,
    FileArchive,
    Folder,
    File
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { formatBytes, formatDate, getFileIconType, type FileIconType } from "@/lib/utils";

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

interface FileListProps {
    files: FileData[];
    onDownload: (fileId: string, filename: string) => void;
    onDelete: (fileId: string, filename: string) => void;
    onShare: (fileId: string) => void;
}

const iconMap: Record<FileIconType, React.ComponentType<{ className?: string }>> = {
    file: File,
    image: Image,
    video: Video,
    music: Music,
    "file-text": FileText,
    "file-archive": FileArchive,
    folder: Folder,
};

export function FileList({ files, onDownload, onDelete, onShare }: FileListProps) {
    if (files.length === 0) {
        return (
            <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                    <FileIcon className="mb-4 h-12 w-12 text-gray-400" />
                    <h3 className="mb-2 text-lg font-medium text-gray-900">No files yet</h3>
                    <p className="text-sm text-gray-500">Upload your first file to get started</p>
                </CardContent>
            </Card>
        );
    }

    const getIconComponent = (mimeType: string) => {
        const iconType = getFileIconType(mimeType);
        return iconMap[iconType];
    };

    return (
        <div className="space-y-3">
            {files.map((file) => {
                const IconComponent = getIconComponent(file.content_type);
                return (
                    <div
                        key={file.id}
                        className="group flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4 transition-all hover:border-blue-300 hover:shadow-sm"
                    >
                        <div className="flex items-center gap-4">
                            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-50">
                                <IconComponent className="h-6 w-6 text-blue-600" />
                            </div>
                            <div>
                                <p className="font-medium text-gray-900">{file.original_filename}</p>
                                <p className="text-sm text-gray-500">
                                    {formatBytes(file.size_bytes)} • {formatDate(file.created_at)}
                                </p>
                            </div>
                        </div>

                        <div className="flex items-center gap-1">
                            <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => onDownload(file.id, file.original_filename)}
                                className="h-9 w-9 p-0 text-gray-600 hover:bg-blue-50 hover:text-blue-600"
                                title="Download"
                            >
                                <Download className="h-4 w-4" />
                            </Button>
                            <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => onShare(file.id)}
                                className="h-9 w-9 p-0 text-gray-600 hover:bg-blue-50 hover:text-blue-600"
                                title="Share"
                            >
                                <Share2 className="h-4 w-4" />
                            </Button>
                            <Button
                                size="sm"
                                variant="ghost"
                                onClick={() => onDelete(file.id, file.original_filename)}
                                className="h-9 w-9 p-0 text-gray-600 hover:bg-red-50 hover:text-red-600"
                                title="Delete"
                            >
                                <Trash2 className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                );
            })}
        </div>
    );
}
