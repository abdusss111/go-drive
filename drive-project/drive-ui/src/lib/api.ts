import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import { useAuthStore } from "@/store/authStore";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

export const apiClient = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        "Content-Type": "application/json",
    },
});

// Request interceptor to add auth token
apiClient.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
        // Get token from Zustand store (synchronous, no race condition)
        const token = useAuthStore.getState().accessToken;
        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// Response interceptor for error handling
apiClient.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
        if (error.response?.status === 401) {
            const { isAuthenticated, logout } = useAuthStore.getState();

            // Only redirect if user was authenticated (token expired/invalid)
            // Don't redirect on login failures or if already logged out
            if (isAuthenticated) {
                logout();

                // Show a toast or log before redirecting
                console.warn('Session expired. Redirecting to login...');

                // Add a small delay so user can see the error
                setTimeout(() => {
                    if (typeof window !== "undefined" && !window.location.pathname.includes("/login")) {
                        window.location.href = "/login";
                    }
                }, 1000);
            }
        }
        return Promise.reject(error);
    }
);

export default apiClient;
