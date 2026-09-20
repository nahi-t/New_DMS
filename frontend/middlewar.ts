// import { NextResponse } from 'next/server';
// import type { NextRequest } from 'next/server';

// export function middleware(request: NextRequest) {
//   const token = request.cookies.get('token')?.value;
//   const { pathname } = request.nextUrl;

//   // 1. Root route redirect
//   if (pathname === '/') {
//     const target = token ? '/dashboard' : '/login';
//     return NextResponse.redirect(new URL(target, request.url));
//   }

//   // 2. Protect /dashboard routes from unauthenticated users
//   if (pathname.startsWith('/dashboard') && !token) {
//     return NextResponse.redirect(new URL('/login', request.url));
//   }

//   // 3. Prevent authenticated users from visiting /login
//   if (pathname === '/login' && token) {
//     return NextResponse.redirect(new URL('/dashboard', request.url));
//   }

//   return NextResponse.next();
// }

// export const config = {
//   // Ignore static assets, images, and internal Next.js requests
//   matcher: [
//     '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
//   ],
// };