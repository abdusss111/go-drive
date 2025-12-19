# UI Requirements - GoDrive Frontend

## Overview

This document describes all user interfaces, pages, and components required for the GoDrive frontend application built with **React** and **Next.js**.

**Tech Stack:**
- React 18+
- Next.js 14+ (App Router)
- TypeScript
- Tailwind CSS (recommended)
- React Query / SWR (for data fetching)
- Zustand / Redux Toolkit (for state management)
- React Hook Form (for form handling)
- Axios / Fetch API (for HTTP requests)

---

## 1. Authentication Pages

### 1.1 Login Page (`/login`)

**Route:** `/login`  
**Access:** Public (redirect to dashboard if authenticated)

**Components:**
- `LoginForm` - Main login form component
- `AuthLayout` - Shared layout for auth pages

**Fields:**
- Email (required, email validation)
- Password (required, min 8 characters)
- "Remember me" checkbox (optional)
- "Forgot password?" link (future feature)

**Actions:**
- Submit login form → `POST /v1/auth/login`
- Redirect to dashboard on success
- Display error messages for invalid credentials
- Show loading state during submission

**UI Requirements:**
- Responsive design (mobile-first)
- Password visibility toggle
- Form validation with real-time feedback
- Error message display
- Link to registration page

**State Management:**
- Store JWT tokens (access + refresh) in secure storage (httpOnly cookies recommended)
- Store user info in global state

---

### 1.2 Registration Page (`/register`)

**Route:** `/register`  
**Access:** Public (redirect to dashboard if authenticated)

**Components:**
- `RegisterForm` - Main registration form component
- `AuthLayout` - Shared layout for auth pages

**Fields:**
- Email (required, email validation, unique check)
- Password (required, min 8, max 72 characters)
- Confirm Password (required, must match)
- Display Name (optional, max 128 characters)

**Actions:**
- Submit registration form → `POST /v1/auth/register`
- Auto-login after successful registration
- Redirect to dashboard on success
- Display error messages (email already exists, validation errors)

**UI Requirements:**
- Password strength indicator
- Password confirmation matching validation
- Real-time email validation
- Terms & conditions checkbox (optional)
- Link to login page

---

## 2. Main Dashboard (`/dashboard`)

**Route:** `/dashboard` or `/` (protected)  
**Access:** Authenticated users only

**Layout:**
- `DashboardLayout` - Main app layout with sidebar/navbar
- `Sidebar` - Navigation sidebar
- `TopBar` - Top navigation bar with user menu

**Components:**
- `UsageStatsCard` - Display total usage and quota
- `BucketsList` - List of user's buckets (grid/list view toggle)
- `QuickActions` - Quick action buttons (create bucket, upload file)
- `RecentFiles` - Recently uploaded files widget

**Data Fetching:**
- `GET /v1/usage` - User usage statistics
- `GET /v1/buckets` - List all buckets

**UI Requirements:**
- Responsive grid layout
- Search/filter buckets
- Sort buckets (by name, date, size)
- Empty state when no buckets exist
- Loading skeletons
- Error states with retry

**Features:**
- Create new bucket (modal or dedicated page)
- View bucket details on click
- Delete bucket (with confirmation)
- Usage progress bars (bytes and file count)

---

## 3. Bucket Management

### 3.1 Bucket List View (`/buckets`)

**Route:** `/buckets`  
**Access:** Authenticated users only

**Components:**
- `BucketCard` - Individual bucket card component
- `BucketGrid` - Grid view container
- `BucketList` - List view container
- `BucketFilters` - Filter and sort controls
- `CreateBucketButton` - Floating action button or header button

**Bucket Card Display:**
- Bucket name
- Description (if available)
- File count
- Total size (formatted: KB, MB, GB)
- Created date
- Usage percentage (visual progress bar)
- Actions menu (view, edit, delete)

**Actions:**
- Click bucket → Navigate to bucket detail page
- Create bucket → Open create bucket modal
- Delete bucket → Confirmation dialog → `DELETE /v1/buckets/:bucketID`
- View usage → Show usage details

**UI Requirements:**
- Grid/List view toggle
- Search by bucket name
- Sort by: name, date created, size, file count
- Filter by: date range, size range
- Pagination (if many buckets)
- Empty state with CTA to create first bucket

---

### 3.2 Bucket Detail Page (`/buckets/[bucketId]`)

**Route:** `/buckets/[bucketId]`  
**Access:** Authenticated users only (bucket owner)

**Components:**
- `BucketHeader` - Bucket name, description, actions
- `BucketUsageStats` - Usage statistics for this bucket
- `FilesList` - List of files in bucket
- `UploadZone` - Drag & drop file upload area
- `BucketActionsMenu` - Edit, delete, share options

**Data Fetching:**
- `GET /v1/buckets/:bucketID` - Bucket details
- `GET /v1/buckets/:bucketID/files` - Files in bucket
- `GET /v1/buckets/:bucketID/usage` - Bucket usage stats

**Actions:**
- Upload files → `POST /v1/buckets/:bucketID/files`
- Download file → `GET /v1/buckets/:bucketID/files/:fileID/download`
- Delete file → `DELETE /v1/buckets/:bucketID/files/:fileID`
- Generate presigned URL → `POST /v1/buckets/:bucketID/files/:fileID/presigned-url`
- Edit bucket (name, description) - Future feature
- Delete bucket → Confirmation → `DELETE /v1/buckets/:bucketID`

**UI Requirements:**
- Breadcrumb navigation
- File upload with progress indicator
- Drag & drop file upload
- File list with thumbnails/icons
- File preview (for images, PDFs)
- File actions menu (download, delete, share, get link)
- File search and filter
- Sort files (by name, size, date)
- Pagination for large file lists
- Empty state when bucket is empty

---

### 3.3 Create Bucket Modal/Page

**Route:** Modal overlay or `/buckets/new`  
**Access:** Authenticated users only

**Components:**
- `CreateBucketForm` - Form component
- `BucketFormModal` - Modal wrapper (if modal)

**Fields:**
- Bucket Name (required, unique per user)
- Description (optional, max 255 characters)

**Actions:**
- Submit → `POST /v1/buckets`
- Close/Cancel → Close modal or navigate back
- Validation before submission

**UI Requirements:**
- Real-time name validation
- Character counter for description
- Error handling (name already exists)
- Success notification
- Auto-redirect to new bucket on success

---

## 4. File Management

### 4.1 File Upload Component

**Components:**
- `FileUploadZone` - Drag & drop area
- `FileUploadProgress` - Upload progress indicator
- `FileUploadList` - List of files being uploaded
- `FileUploadItem` - Individual file upload item with progress

**Features:**
- Drag & drop multiple files
- Click to select files
- File type validation
- File size validation (check quota)
- Upload progress for each file
- Cancel upload
- Retry failed uploads
- Show upload errors

**API Integration:**
- `POST /v1/buckets/:bucketID/files` (multipart/form-data)
- Handle quota errors (413, 429 status codes)
- Show appropriate error messages

**UI Requirements:**
- Visual feedback during drag & drop
- File preview thumbnails
- Upload progress bars
- Error messages per file
- Success indicators

---

### 4.2 File List Component

**Components:**
- `FileItem` - Individual file row/card
- `FileList` - Container for file items
- `FileGrid` - Grid view of files
- `FileActionsMenu` - Context menu for file actions
- `FilePreview` - Preview modal/panel

**File Item Display:**
- File name
- File icon (based on file type)
- File size (formatted)
- Upload date
- File type/MIME type
- Actions button (download, delete, share, get link)

**Actions:**
- Click file → Preview or download
- Download → `GET /v1/buckets/:bucketID/files/:fileID/download`
- Delete → Confirmation → `DELETE /v1/buckets/:bucketID/files/:fileID`
- Get shareable link → Generate presigned URL
- Copy link to clipboard

**UI Requirements:**
- List/Grid view toggle
- File type icons
- File size formatting (KB, MB, GB)
- Date formatting (relative or absolute)
- Hover effects
- Selection (multi-select for bulk actions)
- Sort options
- Filter by file type, date, size
- Search by filename

---

### 4.3 File Preview Component

**Components:**
- `FilePreviewModal` - Modal for file preview
- `ImageViewer` - Image preview component
- `PDFViewer` - PDF preview component
- `TextPreview` - Text file preview
- `VideoPlayer` - Video preview component
- `UnsupportedFileView` - Fallback for unsupported types

**Supported Previews:**
- Images (JPEG, PNG, GIF, WebP)
- PDF documents
- Text files
- Videos (MP4, WebM)
- Audio files (MP3, WAV)

**Actions:**
- Download file
- Get shareable link
- Close preview

**UI Requirements:**
- Full-screen or modal view
- Zoom controls for images
- Navigation for multi-page documents
- Download button
- Share button
- Responsive design

---

## 5. Usage & Statistics

### 5.1 Usage Dashboard (`/usage`)

**Route:** `/usage`  
**Access:** Authenticated users only

**Components:**
- `UsageOverview` - Total usage statistics
- `UsageChart` - Visual chart of usage over time (future)
- `QuotaProgressBars` - Progress bars for bytes and file count
- `BucketUsageList` - List of buckets with their usage

**Data Fetching:**
- `GET /v1/usage` - User total usage
- `GET /v1/buckets` - All buckets with usage stats

**Display:**
- Total bytes used / quota (formatted: GB, TB)
- Total files / quota
- Usage percentage for bytes
- Usage percentage for files
- Visual progress bars
- Breakdown by bucket
- Usage history chart (future feature)

**UI Requirements:**
- Color-coded progress bars (green/yellow/red based on usage)
- Percentage indicators
- Formatted numbers (1.5 GB instead of bytes)
- Responsive charts
- Empty state if no usage

---

### 5.2 Bucket Usage Component

**Components:**
- `BucketUsageCard` - Usage card for individual bucket
- `BucketUsageChart` - Chart showing bucket usage (future)

**Data Fetching:**
- `GET /v1/buckets/:bucketID/usage` - Bucket-specific usage

**Display:**
- Bucket name
- Total bytes used
- File count
- Visual representation (progress bar or chart)

**UI Requirements:**
- Compact design for sidebar/widget
- Detailed view on bucket detail page
- Color coding based on usage percentage

---

## 6. Presigned URLs / Sharing

### 6.1 Share File Modal

**Components:**
- `ShareFileModal` - Modal for sharing files
- `PresignedURLGenerator` - Component to generate presigned URLs
- `LinkCopyButton` - Copy to clipboard button

**Features:**
- Generate presigned URL for GET (download)
- Set expiration time (TTL)
- Copy link to clipboard
- Share via email/social (future)
- QR code generation (future)

**API Integration:**
- `POST /v1/buckets/:bucketID/files/:fileID/presigned-url`
  - Method: "GET"
  - TTL: Optional (default 1 hour)

**UI Requirements:**
- Link display with copy button
- Expiration time display
- Success notification on copy
- Error handling

---

## 7. Common Components

### 7.1 Layout Components

**Components:**
- `AppLayout` - Main application layout
- `Sidebar` - Navigation sidebar
  - Dashboard link
  - Buckets link
  - Usage link
  - Settings link (future)
  - Logout button
- `TopBar` - Top navigation bar
  - Search bar (global search)
  - Notifications (future)
  - User menu (profile, settings, logout)
- `Footer` - Application footer (optional)

**UI Requirements:**
- Responsive sidebar (collapsible on mobile)
- Active route highlighting
- User avatar/name display
- Logout confirmation

---

### 7.2 Navigation Components

**Components:**
- `Breadcrumbs` - Breadcrumb navigation
- `Pagination` - Pagination controls
- `Tabs` - Tab navigation (if needed)

---

### 7.3 Form Components

**Components:**
- `Input` - Text input with validation
- `Textarea` - Textarea input
- `Select` - Dropdown select
- `Checkbox` - Checkbox input
- `Radio` - Radio button group
- `FileInput` - File input with preview
- `FormField` - Wrapper for form fields with label and error

**Features:**
- Validation messages
- Required field indicators
- Error states
- Success states
- Loading states
- Disabled states

---

### 7.4 Feedback Components

**Components:**
- `Toast` / `Notification` - Toast notifications
- `Alert` - Alert messages (error, warning, success, info)
- `LoadingSpinner` - Loading indicator
- `Skeleton` - Loading skeleton screens
- `EmptyState` - Empty state component
- `ErrorBoundary` - Error boundary component

**Use Cases:**
- Success notifications (file uploaded, bucket created)
- Error notifications (upload failed, quota exceeded)
- Loading states during API calls
- Empty states (no buckets, no files)
- Error boundaries for unexpected errors

---

### 7.5 Modal Components

**Components:**
- `Modal` - Base modal component
- `ConfirmDialog` - Confirmation dialog
- `DeleteConfirmDialog` - Delete confirmation dialog

**Features:**
- Backdrop/overlay
- Close on backdrop click
- Close on Escape key
- Focus trap
- Animation/transitions

---

### 7.6 Data Display Components

**Components:**
- `Table` - Data table component
- `Card` - Card component
- `Badge` - Badge/tag component
- `Avatar` - User avatar
- `ProgressBar` - Progress bar
- `Tooltip` - Tooltip component

---

## 8. State Management

### 8.1 Global State

**State Stores:**
- `AuthStore` - Authentication state
  - User info
  - Access token
  - Refresh token
  - Is authenticated
  - Login/logout actions
- `BucketStore` - Buckets state (optional, can use React Query cache)
- `FileStore` - Files state (optional, can use React Query cache)
- `UIStore` - UI state
  - Theme (light/dark)
  - Sidebar collapsed state
  - Modals open/closed
  - Notifications

### 8.2 Server State (React Query / SWR)

**Queries:**
- User usage query
- Buckets list query
- Bucket detail query
- Files list query
- Bucket usage query

**Mutations:**
- Login mutation
- Register mutation
- Create bucket mutation
- Delete bucket mutation
- Upload file mutation
- Delete file mutation
- Generate presigned URL mutation

---

## 9. Routing Structure

### 9.1 Public Routes

```
/login          - Login page
/register       - Registration page
```

### 9.2 Protected Routes

```
/                    - Dashboard (redirect from /)
/dashboard           - Dashboard
/buckets             - Buckets list
/buckets/new         - Create bucket (optional, can be modal)
/buckets/[bucketId]  - Bucket detail page
/usage               - Usage statistics
/settings            - User settings (future)
```

### 9.3 Route Protection

- Protected routes require authentication
- Redirect to `/login` if not authenticated
- Store intended destination for redirect after login
- Refresh token handling for expired access tokens

---

## 10. API Integration

### 10.1 API Client Setup

**Components:**
- `apiClient` - Axios instance with interceptors
- Request interceptor: Add JWT token to headers
- Response interceptor: Handle 401 errors, refresh token

**Base URL:** `http://localhost:8080` (configurable via env)

**Headers:**
- `Authorization: Bearer <access_token>` for protected routes
- `Content-Type: application/json` for JSON requests
- `Content-Type: multipart/form-data` for file uploads

### 10.2 Error Handling

**Error Types:**
- Network errors
- 400 Bad Request - Validation errors
- 401 Unauthorized - Token expired, redirect to login
- 404 Not Found - Resource not found
- 409 Conflict - Resource already exists
- 413 Request Entity Too Large - Quota exceeded (bytes)
- 429 Too Many Requests - Quota exceeded (files)
- 500 Internal Server Error - Server error

**Error Display:**
- Toast notifications for user-facing errors
- Inline form errors for validation errors
- Error pages for critical errors
- Retry mechanisms for transient errors

---

## 11. Responsive Design

### 11.1 Breakpoints

- Mobile: < 640px
- Tablet: 640px - 1024px
- Desktop: > 1024px

### 11.2 Mobile Considerations

- Collapsible sidebar
- Bottom navigation bar (optional)
- Touch-friendly buttons and targets
- Swipe gestures for file actions
- Optimized file upload for mobile
- Responsive tables (cards on mobile)

---

## 12. Accessibility (a11y)

**Requirements:**
- Keyboard navigation support
- ARIA labels for screen readers
- Focus management
- Color contrast compliance (WCAG AA)
- Alt text for images
- Semantic HTML
- Form labels and error associations

---

## 13. Performance Optimizations

**Techniques:**
- Code splitting (Next.js automatic)
- Lazy loading for images
- Virtual scrolling for long lists
- Debounced search inputs
- Optimistic UI updates
- Caching with React Query
- Image optimization (Next.js Image component)

---

## 14. Security Considerations

**Requirements:**
- Store tokens in httpOnly cookies (recommended) or secure localStorage
- CSRF protection (if using cookies)
- XSS prevention (sanitize user inputs)
- Secure file upload validation
- Rate limiting on client side (visual feedback)
- HTTPS in production

---

## 15. Testing Requirements

**Test Types:**
- Unit tests for components
- Integration tests for API calls
- E2E tests for critical flows (login, upload, download)
- Visual regression tests (optional)

**Critical Flows to Test:**
- User registration and login
- Bucket creation and deletion
- File upload and download
- File deletion
- Presigned URL generation
- Quota enforcement UI feedback

---

## 16. Future Enhancements

**Planned Features:**
- File versioning UI
- Advanced search
- File sharing with permissions
- Folder organization
- Bulk operations
- Drag & drop file organization
- Real-time collaboration (future)
- Mobile app (React Native)

---

## 17. Design System

### 17.1 Color Palette

- Primary: Brand color for CTAs
- Secondary: Secondary actions
- Success: Green for success states
- Error: Red for errors
- Warning: Yellow/Orange for warnings
- Info: Blue for informational messages
- Neutral: Grays for text and backgrounds

### 17.2 Typography

- Headings: Bold, clear hierarchy
- Body: Readable font size and line height
- Code: Monospace for technical content

### 17.3 Spacing

- Consistent spacing scale (4px, 8px, 16px, 24px, 32px, etc.)

### 17.4 Icons

- Icon library (Heroicons, Lucide, or custom)
- Consistent icon sizes
- Icon + text combinations

---

## Summary

This document outlines **15+ pages/routes** and **50+ components** required for the GoDrive frontend application. The application follows a modern React/Next.js architecture with:

- **Authentication flow** (login/register)
- **Dashboard** with usage overview
- **Bucket management** (list, create, view, delete)
- **File management** (upload, download, delete, preview)
- **Usage statistics** and quota tracking
- **File sharing** via presigned URLs
- **Responsive design** for all devices
- **Accessibility** compliance
- **Performance optimizations**

All components should be built with TypeScript for type safety and maintainability.

