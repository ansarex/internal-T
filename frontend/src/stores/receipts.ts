import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../utils/api';

export interface Receipt {
  id: number;
  receipt_number: string;
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

export const useReceiptsStore = defineStore('receipts', () => {
  const receipts = ref<Receipt[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchReceipts(params?: { month?: string; status?: string }) {
    loading.value = true;
    try {
      const response = await api.get('/api/receipts', { params });
      receipts.value = response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch receipts.';
    } finally {
      loading.value = false;
    }
  }

  async function fetchActiveClients(month?: string) {
    const response = await api.get('/api/receipts/active-clients', { params: { month } });
    return response.data;
  }

  async function fetchAdminOverview(month?: string) {
    const response = await api.get('/api/receipts/admin-overview', { params: { month } });
    return response.data;
  }

  async function fetchCommissions(month?: string) {
    const response = await api.get('/api/receipts/commissions', { params: { month } });
    return response.data;
  }

  async function createReceipt(data: {
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

    const response = await api.post('/api/receipts', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    receipts.value.unshift(response.data);
    return response.data;
  }

  async function uploadFile(receiptId: number, file: File) {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post(`/api/receipts/${receiptId}/upload-file`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
  }

  async function markPaid(receiptId: number) {
    const response = await api.post(`/api/receipts/${receiptId}/pay`);
    const idx = receipts.value.findIndex((r) => r.id === receiptId);
    if (idx !== -1) receipts.value[idx] = response.data.receipt;
    return response.data;
  }

  async function updateReceipt(receiptId: number, data: { status?: string; notes?: string; amount?: number }) {
    const response = await api.patch(`/api/receipts/${receiptId}`, data);
    const idx = receipts.value.findIndex((r) => r.id === receiptId);
    if (idx !== -1) receipts.value[idx] = response.data;
    return response.data;
  }

  return {
    receipts,
    loading,
    error,
    fetchReceipts,
    fetchActiveClients,
    fetchAdminOverview,
    fetchCommissions,
    createReceipt,
    uploadFile,
    markPaid,
    updateReceipt,
  };
});
