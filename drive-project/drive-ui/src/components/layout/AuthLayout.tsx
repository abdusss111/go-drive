import { ReactNode } from "react";

interface AuthLayoutProps {
    children: ReactNode;
}

export function AuthLayout({ children }: AuthLayoutProps) {
    return (
        <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
            <div className="w-full max-w-md">
                <div className="mb-8 text-center">
                    <h1 className="text-4xl font-bold text-gray-900">GoDrive</h1>
                    <p className="mt-2 text-gray-600">Your secure cloud storage</p>
                </div>
                {children}
            </div>
        </div>
    );
}
