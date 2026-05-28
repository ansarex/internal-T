<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5">
    <h3 class="font-semibold text-gray-800 mb-4">Agreements</h3>

    <div v-for="type in ['service_agreement', 'nda']" :key="type" class="mb-6">
      <div class="flex items-center justify-between mb-2">
        <h4 class="text-sm font-semibold text-gray-700 uppercase tracking-wide">
          {{ type === 'service_agreement' ? 'Service Agreement (SA)' : 'Non-Disclosure Agreement (NDA)' }}
        </h4>
        <button
          v-if="canUpload(type)"
          @click="openUpload(type)"
          class="text-xs bg-blue-600 text-white px-3 py-1 rounded-full hover:bg-blue-700"
        >
          Upload
        </button>
      </div>

      <div v-if="byType(type).length === 0" class="text-sm text-gray-400 italic">No documents uploaded.</div>
      <div v-else class="space-y-2">
        <div
          v-for="ag in byType(type)"
          :key="ag.id"
          class="flex items-start justify-between p-3 rounded-lg border"
          :class="agreementBorderClass(ag.status)"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-sm font-medium text-gray-800">v{{ ag.version }}</span>
              <span :class="statusBadge(ag.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">
                {{ formatStatus(ag.status) }}
              </span>
              <span class="text-xs text-gray-400">by {{ ag.uploader?.name }}</span>
            </div>
            <div v-if="ag.notes" class="text-xs text-red-600 mt-1">Rejection: {{ ag.notes }}</div>
            <div v-if="ag.owner_remarks" class="text-xs text-blue-600 mt-1">Remarks: {{ ag.owner_remarks }}</div>
            <!-- Admin remarks textarea -->
            <div v-if="auth.hasRole('admin') && editRemarksId === ag.id" class="mt-2">
              <textarea
                v-model="remarksText"
                rows="2"
                class="w-full border rounded px-2 py-1 text-xs focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="Add remarks..."
              ></textarea>
              <div class="flex gap-2 mt-1">
                <button @click="saveRemarks(ag.id)" class="text-xs bg-blue-600 text-white px-2 py-0.5 rounded">Save</button>
                <button @click="editRemarksId = null" class="text-xs text-gray-500">Cancel</button>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 ml-3 flex-shrink-0">
            <a
              :href="`${apiBase}/api/agreements/${ag.id}/download`"
              target="_blank"
              class="text-xs text-blue-600 hover:underline"
              @click.prevent="downloadFile(`/api/agreements/${ag.id}/download`)"
            >
              Download
            </a>
            <button
              v-if="auth.hasRole('admin') && editRemarksId !== ag.id"
              @click="openRemarks(ag)"
              class="text-xs text-gray-500 hover:text-gray-700"
            >
              Remarks
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Upload Modal -->
    <div v-if="uploadModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm mx-4">
        <h3 class="text-base font-semibold mb-3">Upload {{ uploadType === 'service_agreement' ? 'Service Agreement' : 'NDA' }}</h3>
        <input type="file" ref="fileInput" accept=".pdf,.doc,.docx" class="w-full text-sm mb-3" />
        <textarea v-model="uploadNotes" rows="2" class="w-full border rounded px-2 py-1 text-sm mb-3" placeholder="Notes (optional)"></textarea>
        <div v-if="uploadError" class="text-red-600 text-xs mb-2">{{ uploadError }}</div>
        <div class="flex gap-3">
          <button @click="doUpload" :disabled="uploading" class="flex-1 bg-blue-600 text-white py-1.5 rounded text-sm disabled:opacity-50">
            {{ uploading ? 'Uploading...' : 'Upload' }}
          </button>
          <button @click="uploadModal = false" class="flex-1 bg-gray-100 text-gray-700 py-1.5 rounded text-sm">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Reject Modal -->
    <div v-if="rejectModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm mx-4">
        <h3 class="text-base font-semibold mb-3">Reject Agreement</h3>
        <textarea v-model="rejectNotes" rows="3" class="w-full border rounded px-2 py-1 text-sm mb-3" placeholder="Rejection reason (optional)"></textarea>
        <div class="flex gap-3">
          <button @click="doReject" class="flex-1 bg-red-600 text-white py-1.5 rounded text-sm">Reject</button>
          <button @click="rejectModal = false" class="flex-1 bg-gray-100 text-gray-700 py-1.5 rounded text-sm">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useJobRequestsStore } from '../../stores/jobRequests';
import { useAuthStore } from '../../stores/auth';
import type { Agreement, JobRequest } from '../../stores/jobRequests';
import api from '../../utils/api';

const props = defineProps<{ job: JobRequest; agreements: Agreement[] }>();
const emit = defineEmits<{ refresh: [] }>();

const jobStore = useJobRequestsStore();
const auth = useAuthStore();

const apiBase = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

const uploadModal = ref(false);
const uploadType = ref('service_agreement');
const uploadNotes = ref('');
const uploading = ref(false);
const uploadError = ref('');
const fileInput = ref<HTMLInputElement | null>(null);

const rejectModal = ref(false);
const rejectTargetId = ref<number | null>(null);
const rejectNotes = ref('');

const editRemarksId = ref<number | null>(null);
const remarksText = ref('');

function byType(type: string) {
  return props.agreements
    .filter((a) => a.type === type)
    .sort((a, b) => b.version - a.version);
}

function canUpload(_type: string) {
  if (auth.hasRole('admin')) return true;
  return auth.hasAnyRole(['sales']) && props.job.current_stage === 1;
}

function openUpload(type: string) {
  uploadType.value = type;
  uploadNotes.value = '';
  uploadError.value = '';
  uploadModal.value = true;
}

async function doUpload() {
  const file = fileInput.value?.files?.[0];
  if (!file) { uploadError.value = 'Please select a file.'; return; }
  uploading.value = true;
  uploadError.value = '';
  try {
    await jobStore.uploadAgreement(props.job.id, uploadType.value, file, uploadNotes.value || undefined);
    uploadModal.value = false;
    emit('refresh');
  } catch (e: any) {
    uploadError.value = e.response?.data?.message || 'Upload failed.';
  } finally {
    uploading.value = false;
  }
}

function openRemarks(ag: Agreement) {
  editRemarksId.value = ag.id;
  remarksText.value = ag.owner_remarks || '';
}

async function saveRemarks(id: number) {
  await jobStore.addRemarks(id, remarksText.value);
  editRemarksId.value = null;
  emit('refresh');
}

async function downloadFile(path: string) {
  const response = await api.get(path, { responseType: 'blob' });
  const url = window.URL.createObjectURL(response.data);
  const a = document.createElement('a');
  a.href = url;
  a.download = path.split('/').pop() || 'file';
  a.click();
  window.URL.revokeObjectURL(url);
}

function statusBadge(status: string) {
  return {
    'bg-yellow-100 text-yellow-800': status === 'pending_approval',
    'bg-green-100 text-green-800': status === 'approved',
    'bg-red-100 text-red-700': status === 'rejected',
    'bg-gray-100 text-gray-600': status === 'draft',
  };
}

function formatStatus(status: string) {
  return {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    approved: 'Approved',
    rejected: 'Rejected',
  }[status] ?? status;
}

function agreementBorderClass(status: string) {
  return {
    'border-yellow-200 bg-yellow-50': status === 'pending_approval',
    'border-green-200 bg-green-50': status === 'approved',
    'border-red-200 bg-red-50': status === 'rejected',
    'border-gray-200': status === 'draft',
  };
}
</script>
