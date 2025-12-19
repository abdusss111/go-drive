import * as React from "react";
import { cn } from "@/lib/utils";

interface ProgressBarProps {
    value: number;
    max?: number;
    className?: string;
    showLabel?: boolean;
}

export function ProgressBar({
    value,
    max = 100,
    className,
    showLabel = false,
}: ProgressBarProps) {
    const percentage = Math.min((value / max) * 100, 100);

    return (
        <div className={cn("w-full", className)}>
            <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
                <div
                    className={cn(
                        "h-full transition-all duration-300",
                        percentage < 50
                            ? "bg-green-600"
                            : percentage < 80
                                ? "bg-yellow-600"
                                : "bg-red-600"
                    )}
                    style={{ width: `${percentage}%` }}
                />
            </div>
            {showLabel && (
                <p className="mt-1 text-sm text-gray-600">
                    {percentage.toFixed(0)}%
                </p>
            )}
        </div>
    );
}
