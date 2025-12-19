"use client";

import { useCallback } from "react";
import { useDropzone } from "react-dropzone";
import { Upload } from "lucide-react";

interface FileUploadZoneProps {
    onFilesSelected: (files: File[]) => void;
    disabled?: boolean;
}

export function FileUploadZone({ onFilesSelected, disabled }: FileUploadZoneProps) {
    const onDrop = useCallback(
        (acceptedFiles: File[]) => {
            onFilesSelected(acceptedFiles);
        },
        [onFilesSelected]
    );

    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,
        disabled,
    });

    return (
        <div
            {...getRootProps()}
            className={`flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-12 transition-colors ${isDragActive
                    ? "border-blue-500 bg-blue-50"
                    : "border-gray-300 bg-gray-50 hover:border-gray-400"
                } ${disabled ? "cursor-not-allowed opacity-50" : ""}`}
        >
            <input {...getInputProps()} />
            <Upload className="mb-4 h-12 w-12 text-gray-400" />
            {isDragActive ? (
                <p className="text-lg font-medium text-blue-600">Drop files here...</p>
            ) : (
                <>
                    <p className="mb-2 text-lg font-medium text-gray-700">
                        Drag & drop files here
                    </p>
                    <p className="text-sm text-gray-500">or click to browse</p>
                </>
            )}
        </div>
    );
}
