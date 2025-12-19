import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone', // For server-side deployment with Nginx

  images: {
    unoptimized: true, // Required for standalone mode
  },

  // Environment variables
  env: {
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8081',
  },
};

export default nextConfig;
