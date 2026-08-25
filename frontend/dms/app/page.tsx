'use client';

import { useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';

export default function HomePage() {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading) {
      router.push(user ? '/dashboard' : '/login');
    }
  }, [user, isLoading, router]);

  return <div className="flex items-center justify-center min-h-screen">Loading...</div>;
}