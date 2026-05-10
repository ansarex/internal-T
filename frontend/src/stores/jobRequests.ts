import { defineStore } from 'pinia';
import { ref } from 'vue';
import api from '../utils/api';

export interface SLAStatus {
  indicator: string;
  sla_deadline?: string;
  days_remaining: number;
  sla_overdue: boolean;
  days_since_update: number;
  stale: boolean;
}

export interface Task {
  id: number;
  job_request_id: number;
  task_type: string;
  status: 'pending' | 'in_progress' | 'completed';
  remarks?: string;
  updated_by?: number;
  updated_by_user?: { id: number; name: string };
  completed_at?: string;
}

export interface Agreement {
  id: number;
  job_request_id: number;
  type: 'service_agreement' | 'nda';
  version: number;
  file_path: string;
  status: 'draft' | 'pending_approval' | 'approved' | 'rejected';
  uploaded_by: number;
  uploader?: { id: number; name: string };
  approved_by?: number;
  approver?: { id: number; name: string };
  approved_at?: string;
  notes?: string;
  owner_remarks?: string;
  created_at: string;
}

export interface JobRequest {
  id: number;
  client_id: number;
  status: string;
  current_stage: number;
  indicator: string;
  customer_pic?: string;
  monthly_recurring?: number;
  account_type?: string;
  recurring_start_date?: string;
  assigned_sales_id?: number;
  assigned_cs_id?: number;
  assigned_sales?: { id: number; name: string; email: string };
  assigned_cs?: { id: number; name: string; email: string };
  signed_file_path?: string;
  signed_uploaded_at?: string;
  sla_started_at?: string;
  sla_deadline?: string;
  last_activity_at?: string;
  stage1_approved_at?: string;
  created_at: string;
  client?: { id: number; company_name: string };
  tasks?: Task[];
  agreements?: Agreement[];
  sla?: SLAStatus;
}

export const useJobRequestsStore = defineStore('jobRequests', () => {
  const jobRequests = ref<JobRequest[]>([]);
  const currentJob = ref<JobRequest | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchJobRequests() {
    loading.value = true;
    try {
      const response = await api.get('/api/job-requests');
      jobRequests.value = response.data;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch job requests.';
    } finally {
      loading.value = false;
    }
  }

  async function fetchJobRequest(id: number) {
    loading.value = true;
    try {
      const response = await api.get(`/api/job-requests/${id}`);
      currentJob.value = { ...response.data.job_request, sla: response.data.sla };
      return currentJob.value;
    } catch (e: any) {
      error.value = e.response?.data?.message || 'Failed to fetch job request.';
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function updateStage1(id: number, data: { customer_pic?: string; monthly_recurring?: number; account_type?: string; recurring_start_date?: string }) {
    const response = await api.patch(`/api/job-requests/${id}/stage1`, data);
    if (currentJob.value?.id === id) {
      Object.assign(currentJob.value, response.data);
    }
    return response.data;
  }

  async function uploadSignedCopy(id: number, file: File) {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post(`/api/job-requests/${id}/signed-copy`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
  }

  async function uploadAgreement(jobId: number, type: string, file: File, notes?: string) {
    const formData = new FormData();
    formData.append('type', type);
    formData.append('file', file);
    if (notes) formData.append('notes', notes);
    const response = await api.post(`/api/job-requests/${jobId}/agreements`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data;
  }

  async function fetchAgreements(jobId: number) {
    const response = await api.get(`/api/job-requests/${jobId}/agreements`);
    return response.data;
  }

  async function approveAgreement(agreementId: number) {
    const response = await api.post(`/api/agreements/${agreementId}/approve`);
    return response.data;
  }

  async function rejectAgreement(agreementId: number, notes?: string) {
    const response = await api.post(`/api/agreements/${agreementId}/reject`, { notes });
    return response.data;
  }

  async function addRemarks(agreementId: number, ownerRemarks: string) {
    const response = await api.post(`/api/agreements/${agreementId}/remarks`, { owner_remarks: ownerRemarks });
    return response.data;
  }

  async function fetchTasks(jobId: number) {
    const response = await api.get(`/api/job-requests/${jobId}/tasks`);
    return response.data;
  }

  async function updateTask(taskId: number, data: { status: string; remarks?: string }) {
    const response = await api.patch(`/api/tasks/${taskId}`, data);
    return response.data;
  }

  async function fetchSLA(jobId: number) {
    const response = await api.get(`/api/job-requests/${jobId}/sla`);
    return response.data;
  }

  return {
    jobRequests,
    currentJob,
    loading,
    error,
    fetchJobRequests,
    fetchJobRequest,
    updateStage1,
    uploadSignedCopy,
    uploadAgreement,
    fetchAgreements,
    approveAgreement,
    rejectAgreement,
    addRemarks,
    fetchTasks,
    updateTask,
    fetchSLA,
  };
});
