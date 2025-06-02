'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { authService, tokenManager, type User, type LoginRequest, type RegisterRequest, type AuthResponse } from '@/lib/api';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<boolean>;
  register: (userData: RegisterRequest) => Promise<boolean>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    initializeAuth();
  }, []);

  const initializeAuth = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      if (token) {
        tokenManager.setTokenForAllServices(token);
        await loadUserProfile();
      }
    } catch (error) {
      console.error('Failed to initialize auth:', error);
      tokenManager.clearAllTokens();
    } finally {
      setIsLoading(false);
    }
  };

  const loadUserProfile = async () => {
    try {
      const userProfile = await authService.getProfile();
      setUser(userProfile);
    } catch (error) {
      console.error('Failed to load user profile:', error);
      throw error;
    }
  };

  const login = async (credentials: LoginRequest): Promise<boolean> => {
    try {
      setIsLoading(true);
      const authResponse: AuthResponse = await authService.login(credentials);
      
      // Store tokens
      localStorage.setItem('auth_token', authResponse.token);
      localStorage.setItem('refresh_token', authResponse.refresh_token);
      localStorage.setItem('token_expires_at', authResponse.expires_at);
      
      // Set tokens for all services
      tokenManager.setTokenForAllServices(authResponse.token);
      
      // Load user profile
      await loadUserProfile();
      
      return true;
    } catch (error) {
      console.error('Login failed:', error);
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (userData: RegisterRequest): Promise<boolean> => {
    try {
      setIsLoading(true);
      const authResponse: AuthResponse = await authService.register(userData);
      
      // Store tokens
      localStorage.setItem('auth_token', authResponse.token);
      localStorage.setItem('refresh_token', authResponse.refresh_token);
      localStorage.setItem('token_expires_at', authResponse.expires_at);
      
      // Set tokens for all services
      tokenManager.setTokenForAllServices(authResponse.token);
      
      // Load user profile
      await loadUserProfile();
      
      return true;
    } catch (error) {
      console.error('Registration failed:', error);
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    try {
      authService.logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
    } finally {
      // Clear local state regardless of API call result
      setUser(null);
      localStorage.removeItem('auth_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('token_expires_at');
      tokenManager.clearAllTokens();
      router.push('/');
    }
  };

  const refreshUser = async () => {
    try {
      await loadUserProfile();
    } catch (error) {
      console.error('Failed to refresh user:', error);
      logout();
    }
  };

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    register,
    logout,
    refreshUser,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export function useRequireAuth() {
  const auth = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!auth.isLoading && !auth.isAuthenticated) {
      router.push('/login');
    }
  }, [auth.isAuthenticated, auth.isLoading, router]);

  return auth;
} 