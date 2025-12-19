"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { Eye, EyeOff } from "lucide-react";
import { AuthLayout } from "@/components/layout/AuthLayout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/components/ui/toast";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/api";

interface RegisterFormData {
    email: string;
    password: string;
    confirmPassword: string;
    displayName?: string;
}

export default function RegisterPage() {
    const router = useRouter();
    const { showToast } = useToast();
    const { setAuth } = useAuthStore();
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [isLoading, setIsLoading] = useState(false);

    const {
        register,
        handleSubmit,
        watch,
        formState: { errors },
    } = useForm<RegisterFormData>();

    const password = watch("password");

    const getPasswordStrength = (pwd: string): string => {
        if (!pwd) return "";
        if (pwd.length < 8) return "Weak";
        if (pwd.length < 12) return "Medium";
        return "Strong";
    };

    const passwordStrength = getPasswordStrength(password);

    const onSubmit = async (data: RegisterFormData) => {
        setIsLoading(true);
        try {
            const response = await apiClient.post("/v1/auth/register", {
                email: data.email,
                password: data.password,
                display_name: data.displayName,
            });

            // Backend returns: { user: {...}, tokens: { access_token, refresh_token } }
            const { user, tokens } = response.data;
            setAuth(user, tokens.access_token, tokens.refresh_token);

            showToast("Registration successful!", "success");
            router.push("/dashboard");
        } catch (error: any) {
            const message = error.response?.data?.error || "Registration failed";
            showToast(message, "error");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <AuthLayout>
            <Card>
                <CardHeader>
                    <CardTitle>Create an account</CardTitle>
                    <CardDescription>Get started with your free account</CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                        <div>
                            <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                                Email
                            </label>
                            <Input
                                id="email"
                                type="email"
                                placeholder="you@example.com"
                                {...register("email", {
                                    required: "Email is required",
                                    pattern: {
                                        value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                                        message: "Invalid email address",
                                    },
                                })}
                                className="mt-1"
                            />
                            {errors.email && (
                                <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
                            )}
                        </div>

                        <div>
                            <label htmlFor="displayName" className="block text-sm font-medium text-gray-700">
                                Display Name (Optional)
                            </label>
                            <Input
                                id="displayName"
                                type="text"
                                placeholder="John Doe"
                                {...register("displayName", {
                                    maxLength: {
                                        value: 128,
                                        message: "Display name must be less than 128 characters",
                                    },
                                })}
                                className="mt-1"
                            />
                            {errors.displayName && (
                                <p className="mt-1 text-sm text-red-600">{errors.displayName.message}</p>
                            )}
                        </div>

                        <div>
                            <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                                Password
                            </label>
                            <div className="relative mt-1">
                                <Input
                                    id="password"
                                    type={showPassword ? "text" : "password"}
                                    placeholder="••••••••"
                                    {...register("password", {
                                        required: "Password is required",
                                        minLength: {
                                            value: 8,
                                            message: "Password must be at least 8 characters",
                                        },
                                        maxLength: {
                                            value: 72,
                                            message: "Password must be less than 72 characters",
                                        },
                                    })}
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500"
                                >
                                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            {password && (
                                <p className={`mt-1 text-sm ${passwordStrength === "Strong" ? "text-green-600" :
                                        passwordStrength === "Medium" ? "text-yellow-600" :
                                            "text-red-600"
                                    }`}>
                                    Strength: {passwordStrength}
                                </p>
                            )}
                            {errors.password && (
                                <p className="mt-1 text-sm text-red-600">{errors.password.message}</p>
                            )}
                        </div>

                        <div>
                            <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700">
                                Confirm Password
                            </label>
                            <div className="relative mt-1">
                                <Input
                                    id="confirmPassword"
                                    type={showConfirmPassword ? "text" : "password"}
                                    placeholder="••••••••"
                                    {...register("confirmPassword", {
                                        required: "Please confirm your password",
                                        validate: (value) =>
                                            value === password || "Passwords do not match",
                                    })}
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500"
                                >
                                    {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            {errors.confirmPassword && (
                                <p className="mt-1 text-sm text-red-600">{errors.confirmPassword.message}</p>
                            )}
                        </div>

                        <Button type="submit" className="w-full" disabled={isLoading}>
                            {isLoading ? "Creating account..." : "Create account"}
                        </Button>

                        <p className="text-center text-sm text-gray-600">
                            Already have an account?{" "}
                            <Link href="/login" className="font-medium text-blue-600 hover:text-blue-500">
                                Sign in
                            </Link>
                        </p>
                    </form>
                </CardContent>
            </Card>
        </AuthLayout>
    );
}
