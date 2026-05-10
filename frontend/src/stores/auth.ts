import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import api from '../utils/api';

export interface UserProject {
  id: number;
  project_id: number;
  user_id: number;
  project: { id: number; name: string; slug: string };
}

export interface User {
  id: number;
  name: string;
  email: string;
  role: string[];
  is_active: boolean;
  email_verified_at?: string;
  projects?: UserProject[];
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(
    typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null
  );
  const loading = ref(false);
  const error = ref<string | null>(null);

  const isAuthenticated = computed(() => !!token.value && !!user.value);

  function hasRole(role: string): boolean {
    return user.value?.role?.includes(role) ?? false;
  }

  function hasAnyRole(roles: string[]): boolean {
    return roles.some((r) => hasRole(r));
  }

  function isInProject(slug: string): boolean {
    if (hasRole('admin')) return true;
    return user.value?.projects?.some((p) => p.project?.slug === slug) ?? false;
  }

  async function login(email: string, password: string): Promise<{ email_verified: boolean; must_change_password?: boolean; message?: string }> {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.post('/api/login', { email, password });
      const { token: newToken, user: userData, email_verified, must_change_password } = response.data;

      if (!email_verified) {
        return { email_verified: false, message: response.data.message };
      }

      token.value = newToken;
      user.value = userData;

      if (typeof window !== 'undefined') {
        localStorage.setItem('auth_token', newToken);
      }

      return { email_verified: true, must_change_password: !!must_change_password };
    } catch (e: any) {
      const msg = e.response?.data?.message || 'Login failed.';
      error.value = msg;
      const email_verified = e.response?.data?.email_verified ?? true;
      return { email_verified, message: msg };
    } finally {
      loading.value = false;
    }
  }

  async function logout(): Promise<void> {
    try {
      await api.post('/api/logout');
    } catch {
      // ignore errors on logout
    } finally {
      token.value = null;
      user.value = null;
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
        window.location.href = '/login';
      }
    }
  }

  async function fetchUser(): Promise<void> {
    if (!token.value) return;
    try {
      const response = await api.get('/api/me');
      user.value = response.data;
    } catch {
      token.value = null;
      user.value = null;
      if (typeof window !== 'undefined') {
        localStorage.removeItem('auth_token');
      }
    }
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    hasRole,
    hasAnyRole,
    isInProject,
    login,
    logout,
    fetchUser,
  };
});
