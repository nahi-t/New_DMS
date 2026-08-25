'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User } from '@/type';
import { login as apiLogin, register as apiRegister } from '@/lib/api';
import {
  setTokenCookie,
  removeTokenCookie,
  setUserStorage,
  removeUserStorage,
  getUserStorage,
  getTokenCookie,
} from '@/lib/utils';
import toast from 'react-hot-toast';
import { useRouter } from 'next/navigation';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  register: (username: string, email: string, password: string, role: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // Load session from storage on mount
  useEffect(() => {
    const storedUser = getUserStorage();
    const storedToken = getTokenCookie();
    if (storedUser && storedToken) {
      setUser(storedUser);
      setToken(storedToken);
    }
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    try {
      const data = await apiLogin(email, password);
      setToken(data.token);
      setUser(data.user);
      setTokenCookie(data.token);
      setUserStorage(data.user);
      toast.success('Welcome back!');
      router.push('/dashboard');
    } catch (error: any) {
      toast.error(error.message || 'Login failed');
      throw error;
    }
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    removeTokenCookie();
    removeUserStorage();
    toast.success('Logged out');
    router.push('/login');
  };

  const register = async (username: string, email: string, password: string, role: string) => {
    try {
      await apiRegister(username, email, password, role);
      toast.success('Registration successful! Please log in.');
    } catch (error: any) {
      toast.error(error.message || 'Registration failed');
      throw error;
    }
  };

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, logout, register }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used within AuthProvider');
  return context;
};