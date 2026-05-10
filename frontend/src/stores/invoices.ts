import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../utils/api';

export interface Invoice {
  id: number;
  invoice_number: string;
  client_id: number;
  job_request_id: number;
  assigned_sales_id?: number;
  assigned_cs_id?: number;
  amount: number;
  sales_commission: number;
  cs_commission: number;
  billing_month: string;
  status: 'pending' | 'paid' | 'overdue';
  notes?: string;
  file_path?: string;
  file_uploaded_at?: string;
  paid_at?: string;
  paid_by?: number;
  created_by: number;
  created_at: string;
  client?: { id: number; company_name: string };
  assigned_sales?: { id: number; name: string; email: string };
  assigned_cs?: { id: number; name: string; email: string };
}

export const useInvoicesStore = defineStore('invoices', () => {
  const invoices = ref<Invoice[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchInvoices(params?: { month?: string; status?: string }) {
    loading.value = true;
    try {
      const response = await api.get('/api/invoices', { params });
      invoices.value = response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch invoices.';
    } finally {
      loading.value = false;
    }
  }

  async function fetchActiveClients(month?: string) {
    const response = await api.get('/api/invoices/active-clients', { params: { month } });
    return response.data;
  }

  async function fetchAdminOverview(month?: string) {
    const response = await api.get('/api/invoices/admin-overview', { params: { month } });
    return response.data;
  }

  async function fetchCommissions(month?: string) {
    const response = await api.get('/api/invoices/commissions', { params: { month } });
    return response.data;
  }

  async function createInvoice(data: {
    client_id: number;
    job_request_id: number;
    amount: number;
    billing_month: string;
    notes?: string;
    file?: File;
  }) {
    const formData = new FormData();
    formData.append('client_id', String(data.client_id));
    formData.append('job_request_id', String(data.job_request_id));
    formData.append('amount', String(data.amount));
    formData.append('billing_month', data.billing_month);
    if (data.notes) formData.append('notes', data.notes);
    if (data.file) formData.append('file', data.file);

    const response = await api.post('/api/invoices', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    invoices.value.unshift(response.data);
    return response.data;
  }

  async function uploadFile(invoiceId: number, file: File) {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post(`/api/invoices/${invoiceId}/upload-file`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
  }

  async function markPaid(invoiceId: number) {
    const response = await api.post(`/api/invoices/${invoiceId}/pay`);
    const idx = invoices.value.findIndex((i) => i.id === invoiceId);
    if (idx !== -1) invoices.value[idx] = response.data.invoice;
    return response.data;
  }

  async function updateInvoice(invoiceId: number, data: { status?: string; notes?: string; amount?: number }) {
    const response = await api.patch(`/api/invoices/${invoiceId}`, data);
    const idx = invoices.value.findIndex((i) => i.id === invoiceId);
    if (idx !== -1) invoices.value[idx] = response.data;
    return response.data;
  }

  return {
    invoices,
    loading,
    error,
    fetchInvoices,
    fetchActiveClients,
    fetchAdminOverview,
    fetchCommissions,
    createInvoice,
    uploadFile,
    markPaid,
    updateInvoice,
  };
});
