import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../utils/api';

export interface StaffUser {
  id: number;
  name: string;
  email: string;
  role: string[];
  is_active: boolean;
  email_verified_at?: string;
  created_at: string;
}

export interface Project {
  id: number;
  name: string;
  slug: string;
  description: string;
  is_active: boolean;
}

export interface ProjectMember {
  id: number;
  project_id: number;
  user_id: number;
  user: StaffUser;
}

export const useStaffStore = defineStore('staff', () => {
  const staff = ref<StaffUser[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchStaff(roleFilter?: string) {
    loading.value = true;
    try {
      const params = roleFilter ? { role: roleFilter } : {};
      const response = await api.get('/api/staff', { params });
      staff.value = response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch staff.';
    } finally {
      loading.value = false;
    }
  }

  async function createStaff(data: { name: string; email: string; password: string; role?: string[] }) {
    loading.value = true;
    try {
      const response = await api.post('/api/staff', data);
      staff.value.push(response.data);
      return response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to create staff.';
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function updateRole(staffId: number, role: string[]) {
    const response = await api.patch(`/api/staff/${staffId}/role`, { role });
    const idx = staff.value.findIndex((s) => s.id === staffId);
    if (idx !== -1) staff.value[idx] = response.data;
    return response.data;
  }

  async function deactivateStaff(staffId: number) {
    const response = await api.post(`/api/staff/${staffId}/deactivate`);
    const idx = staff.value.findIndex((s) => s.id === staffId);
    if (idx !== -1) staff.value[idx] = response.data.user;
    return response.data;
  }

  async function activateStaff(staffId: number) {
    const response = await api.post(`/api/staff/${staffId}/activate`);
    const idx = staff.value.findIndex((s) => s.id === staffId);
    if (idx !== -1) staff.value[idx] = response.data.user;
    return response.data;
  }

  async function deleteStaff(staffId: number) {
    await api.delete(`/api/staff/${staffId}`);
    staff.value = staff.value.filter((s) => s.id !== staffId);
  }

  const projects = ref<Project[]>([]);

  async function fetchProjects() {
    const res = await api.get('/api/projects');
    projects.value = res.data;
    return res.data as Project[];
  }

  async function fetchProjectStaff(projectId: number) {
    const res = await api.get(`/api/projects/${projectId}/staff`);
    return res.data as ProjectMember[];
  }

  async function addToProject(projectId: number, userId: number) {
    const res = await api.post(`/api/projects/${projectId}/staff`, { user_id: userId });
    return res.data as ProjectMember;
  }

  async function removeFromProject(projectId: number, userId: number) {
    await api.delete(`/api/projects/${projectId}/staff/${userId}`);
  }

  return {
    staff,
    loading,
    error,
    projects,
    fetchStaff,
    createStaff,
    updateRole,
    deactivateStaff,
    activateStaff,
    deleteStaff,
    fetchProjects,
    fetchProjectStaff,
    addToProject,
    removeFromProject,
  };
});
