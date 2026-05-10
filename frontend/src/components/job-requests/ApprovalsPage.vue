<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Pending Approvals</h1>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else-if="!auth.hasRole('admin')" class="bg-red-50 text-red-600 p-4 rounded-lg">
      Access denied. Admin only.
    </div>

    <div v-else-if="agreements.length === 0" class="bg-white rounded-xl border border-gray-200 p-8 text-center text-gray-400">
      No pending agreements.
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="ag in agreements"
        :key="ag.id"
        class="bg-white rounded-xl border border-yellow-200 p-5 flex items-start justify-between gap-4"
      >
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 flex-wrap mb-1">
            <span class="font-semibold text-gray-800">{{ ag.job_request?.client?.company_name }}</span>
            <span class="text-xs text-gray-500">Job #{{ ag.job_request_id }}</span>
          </div>
          <div class="flex items-center gap-2 flex-wrap text-sm text-gray-600">
            <span class="font-medium">{{ ag.type === 'service_agreement' ? 'Service Agreement' : 'NDA' }}</span>
            <span>v{{ ag.version }}</span>
            <span>•</span>
            <span>Uploaded by {{ ag.uploader?.name }}</span>
            <span>•</span>
            <span>{{ formatDate(ag.created_at) }}</span>
          </div>
          <div v-if="ag.notes" class="text-xs text-gray-500 mt-1">Notes: {{ ag.notes }}</div>

          <!-- Remarks for this agreement -->
          <div v-if="editRemarksId === ag.id" class="mt-2">
            <textarea
              v-model="remarksText"
              rows="2"
              class="w-full border rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="Add owner remarks..."
            ></textarea>
            <div class="flex gap-2 mt-1">
              <button @click="saveRemarks(ag.id)" class="text-xs bg-blue-600 text-white px-3 py-1 rounded">Save</button>
              <button @click="editRemarksId = null" class="text-xs text-gray-500">Cancel</button>
            </div>
          </div>
          <div v-else-if="ag.owner_remarks" class="text-xs text-blue-600 mt-1">Remarks: {{ ag.owner_remarks }}</div>
        </div>

        <div class="flex items-center gap-2 flex-shrink-0 flex-wrap">
          <button
            @click="downloadFile(`/api/agreements/${ag.id}/download`)"
            class="text-xs text-blue-600 hover:underline"
          >
            Download
          </button>
          <button
            v-if="editRemarksId !== ag.id"
            @click="openRemarks(ag)"
            class="text-xs text-gray-500 hover:text-gray-700"
          >
            Remarks
          </button>
          <button
            @click="approve(ag.id)"
            :disabled="actionId === ag.id"
            class="text-xs bg-green-600 text-white px-3 py-1.5 rounded-lg hover:bg-green-700 disabled:opacity-50"
          >
            Approve
          </button>
          <button
            @click="openReject(ag.id)"
            :disabled="actionId === ag.id"
            class="text-xs bg-red-600 text-white px-3 py-1.5 rounded-lg hover:bg-red-700 disabled:opacity-50"
          >
            Reject
          </button>
        </div>
      </div>
    </div>

    <!-- Reject Modal -->
    <div v-if="rejectModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm mx-4">
        <h3 class="text-base font-semibold mb-3">Reject Agreement</h3>
        <textarea v-model="rejectNotes" rows="3" class="w-full border rounded px-2 py-1 text-sm mb-3" placeholder="Reason (optional)"></textarea>
        <div class="flex gap-3">
          <button @click="doReject" class="flex-1 bg-red-600 text-white py-2 rounded text-sm">Reject</button>
          <button @click="rejectModal = false" class="flex-1 bg-gray-100 text-gray-700 py-2 rounded text-sm">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '../../stores/auth';
import { useJobRequestsStore } from '../../stores/jobRequests';
import api from '../../utils/api';

const auth = useAuthStore();
const jobStore = useJobRequestsStore();

const loading = ref(true);
const agreements = ref<any[]>([]);
const actionId = ref<number | null>(null);
const editRemarksId = ref<number | null>(null);
const remarksText = ref('');
const rejectModal = ref(false);
const rejectTargetId = ref<number | null>(null);
const rejectNotes = ref('');

onMounted(async () => {
  await auth.fetchUser();
  await loadApprovals();
});

async function loadApprovals() {
  loading.value = true;
  try {
    const res = await api.get('/api/dashboard');
    agreements.value = res.data.pending_approvals || [];
  } finally {
    loading.value = false;
  }
}

async function approve(id: number) {
  actionId.value = id;
  try {
    await jobStore.approveAgreement(id);
    await loadApprovals();
  } finally {
    actionId.value = null;
  }
}

function openReject(id: number) {
  rejectTargetId.value = id;
  rejectNotes.value = '';
  rejectModal.value = true;
}

async function doReject() {
  if (!rejectTargetId.value) return;
  await jobStore.rejectAgreement(rejectTargetId.value, rejectNotes.value || undefined);
  rejectModal.value = false;
  await loadApprovals();
}

function openRemarks(ag: any) {
  editRemarksId.value = ag.id;
  remarksText.value = ag.owner_remarks || '';
}

async function saveRemarks(id: number) {
  await jobStore.addRemarks(id, remarksText.value);
  editRemarksId.value = null;
  await loadApprovals();
}

async function downloadFile(path: string) {
  const response = await api.get(path, { responseType: 'blob' });
  const url = window.URL.createObjectURL(response.data);
  const a = document.createElement('a');
  a.href = url;
  a.download = path.split('/').pop() || 'agreement';
  a.click();
  window.URL.revokeObjectURL(url);
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('en-MY', { day: '2-digit', month: 'short', year: 'numeric' });
}
</script>
