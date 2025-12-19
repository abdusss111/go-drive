# GoDrive Frontend

A modern, responsive cloud storage frontend built with Next.js 14, TypeScript, and Tailwind CSS.

## Features

- 🔐 **Authentication**: Login and registration with JWT tokens
- 📦 **Bucket Management**: Create, list, view, and delete storage buckets
- 📁 **File Operations**: Upload, download, delete, and share files
- 📊 **Usage Statistics**: Real-time quota tracking and usage visualization
- 🔗 **File Sharing**: Generate presigned URLs with configurable expiration
- 📱 **Responsive Design**: Mobile-first UI with Tailwind CSS
- ⚡ **Modern Stack**: Next.js 14 App Router, TypeScript, Zustand

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- GoDrive backend API running (default: `http://localhost:8080`)

### Installation

```bash
# Install dependencies
npm install

# Copy environment variables
cp .env.example .env.local

# Update .env.local with your backend API URL
# NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

### Development

```bash
# Run development server
npm run dev

# Open http://localhost:3000
```

### Production

```bash
# Build for production
npm run build

# Start production server
npm start
```

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
│   ├── login/             # Login page
│   ├── register/          # Registration page
│   ├── dashboard/         # Main dashboard
│   ├── buckets/           # Bucket management
│   │   └── [bucketId]/   # Bucket detail page
│   └── usage/             # Usage statistics
├── components/
│   ├── ui/                # Reusable UI components
│   ├── layout/            # Layout components
│   └── files/             # File-related components
├── lib/
│   ├── api.ts            # Axios API client
│   └── utils.ts          # Utility functions
└── store/
    └── authStore.ts      # Authentication state
```

## Environment Variables

See `.env.example` for all available configuration options:

- `NEXT_PUBLIC_API_BASE_URL`: Backend API endpoint
- `NEXT_PUBLIC_APP_NAME`: Application name
- `NEXT_PUBLIC_MAX_FILE_SIZE`: Maximum file upload size
- Feature flags for preview and sharing

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Forms**: React Hook Form + Zod
- **HTTP Client**: Axios
- **File Upload**: React Dropzone
- **Icons**: Lucide React

## API Integration

The frontend integrates with the GoDrive backend API:

- Authentication (login, register)
- Bucket CRUD operations
- File upload/download/delete
- Presigned URL generation
- Usage statistics

## License

MIT
