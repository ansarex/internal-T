import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../utils/api';

export interface Client {
  id: number;
  company_name: string;
  todo_list?: string;
  account_status: 'inactive' | 'active' | 'paused';
  pending_account_status?: string;
  pending_status_requested_by?: number;
  pending_status_requested_at?: string;
  created_by: number;
  created_at: string;
  updated_at: string;
  creator?: { id: number; name: string; email: string };
  job_requests?: any[];
}

export const useClientsStore = defineStore('clients', () => {
  const clients = ref<Client[]>([]);
  const currentClient = ref<Client | null>(null);
  const totalRecurring = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchClients() {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.get('/api/clients');
      clients.value = response.data.clients;
      totalRecurring.value = response.data.total_recurring;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch clients.';
    } finally {
      loading.value = false;
    }
  }

  async function fetchClient(id: number) {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.get(`/api/clients/${id}`);
      currentClient.value = response.data;
      return response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch client.';
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function createClient(data: {
    company_name: string;
    todo_list?: string;
    assigned_sales_id?: number;
    assigned_cs_id?: number;
  }) {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.post('/api/clients', data);
      clients.value.unshift(response.data);
      return response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to create client.';
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function updateClient(id: number, data: { company_name?: string; todo_list?: string }) {
    loading.value = true;
    error.value = null;
    try {
      const response = await api.put(`/api/clients/${id}`, data);
      const idx = clients.value.findIndex((c) => c.id === id);
      if (idx !== -1) clients.value[idx] = response.data;
      if (currentClient.value?.id === id) currentClient.value = response.data;
      return response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to update client.';
      throw e;
    } finally {
      loading.value = false;
    }
  }

  async function activateClient(id: number) {
    const response = await api.post(`/api/clients/${id}/activate`);
    await fetchClients();
    return response.data;
  }

  async function pauseClient(id: number) {
    const response = await api.post(`/api/clients/${id}/pause`);
    await fetchClients();
    return response.data;
  }

  async function requestDeactivate(id: number) {
    const response = await api.post(`/api/clients/${id}/request-deactivate`);
    await fetchClients();
    return response.data;
  }

  async function approveDeactivate(id: number) {
    const response = await api.post(`/api/clients/${id}/approve-deactivate`);
    await fetchClients();
    return response.data;
  }

  async function rejectDeactivate(id: number) {
    const response = await api.post(`/api/clients/${id}/reject-deactivate`);
    await fetchClients();
    return response.data;
  }

  return {
    clients,
    currentClient,
    totalRecurring,
    loading,
    error,
    fetchClients,
    fetchClient,
    createClient,
    updateClient,
    activateClient,
    pauseClient,
    requestDeactivate,
    approveDeactivate,
    rejectDeactivate,
  };
});
