# VPSentinel Web

The operations UI is a React and TypeScript application built with Vite,
Tailwind CSS, DaisyUI, Radix UI primitives, Lucide icons, TanStack Router, and
TanStack Query.

## Structure

- `src/components/ui`: reusable buttons, dialogs, panels, status, and empty states.
- `src/features`: fleet overview, node operations, and incident views.
- `src/hooks`: admin session and fleet data orchestration.
- `src/lib`: API client and display formatting.
- `src/types.ts`: API response contracts shared by features.

The access JWT is held only in React memory. On reload or expiry, the app calls
the refresh endpoint using the HttpOnly cookie and retries the interrupted API
request once. The refresh credential is never accessible to JavaScript.

```bash
npm ci
npm run dev
npm run lint
npm run build
```
