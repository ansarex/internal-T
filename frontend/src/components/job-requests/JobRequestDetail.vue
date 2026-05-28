<template>
  <div>
    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
    <div v-else-if="!job" class="text-center py-12 text-gray-400">Job request not found.</div>
    <div v-else>
      <!-- Header -->
      <div class="flex flex-wrap items-start justify-between gap-4 mb-6">
        <div>
          <div class="flex items-center gap-2 mb-1">
            <a href="/job-requests" class="text-blue-600 hover:underline text-sm">← Job Requests</a>
          </div>
          <h1 class="text-2xl font-bold text-gray-900">{{ job.client?.company_name }}</h1>
          <p class="text-sm text-gray-500 mt-0.5">Job #{{ job.id }}</p>
        </div>
        <div class="flex items-center gap-3 flex-wrap">
          <SLACountdown :sla="job.sla" />
          <span :class="stageBadge" class="px-3 py-1 rounded-full text-sm font-semibold">
            {{ job.current_stage === 1 ? 'Sales Tasks' : 'CS Tasks' }}
          </span>
        </div>
      </div>

      <!-- Info bar -->
      <div class="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
        <div class="bg-white rounded-lg border border-gray-200 p-3">
          <div class="text-xs text-gray-500 mb-0.5">Sales PIC</div>
          <div class="text-sm font-medium">{{ job.assigned_sales?.name || '—' }}</div>
        </div>
        <div class="bg-white rounded-lg border border-gray-200 p-3">
          <div class="text-xs text-gray-500 mb-0.5">CS PIC</div>
          <div class="text-sm font-medium">{{ job.assigned_cs?.name || '—' }}</div>
        </div>
        <div class="bg-white rounded-lg border border-gray-200 p-3">
          <div class="text-xs text-gray-500 mb-0.5">Account Type</div>
          <div class="text-sm font-medium">{{ job.account_type || '—' }}</div>
        </div>
        <div class="bg-white rounded-lg border border-gray-200 p-3">
          <div class="text-xs text-gray-500 mb-0.5">Monthly Recurring</div>
          <div class="text-sm font-medium">{{ job.monthly_recurring ? 'RM ' + formatCurrency(job.monthly_recurring) : '—' }}</div>
        </div>
        <div class="bg-white rounded-lg border border-gray-200 p-3">
          <div class="text-xs text-gray-500 mb-0.5">Billing Due Day</div>
          <div class="text-sm font-medium">{{ billingDueDay || '—' }}</div>
        </div>
      </div>

      <!-- Stage tabs -->
      <div class="flex gap-1 mb-4 border-b border-gray-200">
        <button
          @click="activeTab = 'stage1'"
          :class="activeTab === 'stage1' ? 'border-b-2 border-blue-600 text-blue-600 font-semibold' : 'text-gray-500 hover:text-gray-700'"
          class="px-4 py-2 text-sm transition-colors"
        >
          Sales Tasks
        </button>
        <button
          @click="activeTab = 'stage2'"
          :disabled="job.current_stage < 2"
          :class="[
            activeTab === 'stage2' ? 'border-b-2 border-blue-600 text-blue-600 font-semibold' : 'text-gray-500 hover:text-gray-700',
            job.current_stage < 2 ? 'opacity-40 cursor-not-allowed' : '',
          ]"
          class="px-4 py-2 text-sm transition-colors flex items-center gap-1"
        >
          CS Tasks <span v-if="job.current_stage < 2">🔒</span>
        </button>
      </div>

      <!-- Stage 1 content -->
      <div v-if="activeTab === 'stage1'" class="space-y-4">
        <Stage1Form :job="job" @updated="refresh" />
        <AgreementPanel :job="job" :agreements="agreements" @refresh="refresh" />
        <SignedCopyUpload
          v-if="canShowSignedCopyUpload"
          :jobId="job.id"
          @uploaded="refresh"
        />
        <div v-if="job.signed_file_path" class="bg-green-50 border border-green-200 rounded-xl p-4 flex items-center justify-between">
          <div>
            <span class="text-green-800 font-medium text-sm">Signed copy uploaded</span>
            <span v-if="job.signed_uploaded_at" class="text-green-600 text-xs ml-2">{{ formatDate(job.signed_uploaded_at) }}</span>
          </div>
          <button
            @click="downloadSigned"
            class="text-green-700 hover:text-green-900 text-xs font-medium underline"
          >
            Download
          </button>
        </div>
      </div>

      <!-- Stage 2 content -->
      <div v-if="activeTab === 'stage2'" class="space-y-4">
        <Stage2Tasks :job="job" :tasks="tasks" @updated="refresh" />
        <AccountStatusActions v-if="job.client" :client="jobClient" @refresh="refresh" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useJobRequestsStore } from '../../stores/jobRequests';
import { useClientsStore } from '../../stores/clients';
import { useAuthStore } from '../../stores/auth';
import type { Agreement, Task } from '../../stores/jobRequests';
import type { Client } from '../../stores/clients';
import SLACountdown from './SLACountdown.vue';
import Stage1Form from './Stage1Form.vue';
import AgreementPanel from './AgreementPanel.vue';
import SignedCopyUpload from './SignedCopyUpload.vue';
import Stage2Tasks from './Stage2Tasks.vue';
import AccountStatusActions from './AccountStatusActions.vue';
import api from '../../utils/api';

const props = defineProps<{ jobId: number }>();

const jobStore = useJobRequestsStore();
const clientsStore = useClientsStore();
const auth = useAuthStore();

const loading = ref(true);
const activeTab = ref('stage1');
const agreements = ref<Agreement[]>([]);
const tasks = ref<Task[]>([]);
const jobClient = ref<Client>({} as Client);

const job = computed(() => jobStore.currentJob);

const canShowSignedCopyUpload = computed(() => {
  if (!job.value) return false;
  if (!auth.hasAnyRole(['sales', 'admin'])) return false;
  if (job.value.current_stage !== 1) return false;
  if (job.value.signed_file_path) return false;
  const hasSA = agreements.value.some((a) => a.type === 'service_agreement');
  const hasNDA = agreements.value.some((a) => a.type === 'nda');
  return hasSA && hasNDA;
});

const billingDueDay = computed(() => {
  const d = job.value?.recurring_start_date;
  if (!d) return null;
  const date = new Date(d);
  if (isNaN(date.getTime())) return null;
  const day = date.getUTCDate();
  const s = ['th', 'st', 'nd', 'rd'];
  const v = day % 100;
  const suffix = s[(v - 20) % 10] || s[v] || s[0];
  return `${day}${suffix} of each month`;
});

const stageBadge = computed(() => {
  const ind = job.value?.indicator;
  return {
    'bg-green-100 text-green-800': ind === 'green',
    'bg-red-100 text-red-800': ind === 'red',
    'bg-yellow-100 text-yellow-800': ind === 'yellow',
    'bg-gray-100 text-gray-700': ind === 'grey' || !ind,
  };
});

let refreshTimer: ReturnType<typeof setInterval> | null = null;

onMounted(async () => {
  await auth.fetchUser();
  await refresh();
//  refreshTimer = setInterval(refresh, 30000);
});

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer);
});

async function refresh() {
  loading.value = true;
  try {
    const [j, agRes, taskRes] = await Promise.all([
      jobStore.fetchJobRequest(props.jobId),
      jobStore.fetchAgreements(props.jobId),
      jobStore.fetchTasks(props.jobId),
    ]);
    agreements.value = agRes;
    tasks.value = taskRes;
    if (j?.client_id) {
      const clientData = await clientsStore.fetchClient(j.client_id);
      if (clientData) jobClient.value = clientData;
    }
  } finally {
    loading.value = false;
  }
}

async function downloadSigned() {
  const response = await api.get(`/api/job-requests/${props.jobId}/signed-copy/download`, { responseType: 'blob' });
  const url = window.URL.createObjectURL(response.data);
  const a = document.createElement('a');
  a.href = url;
  a.download = `signed-copy-job-${props.jobId}.pdf`;
  a.click();
  window.URL.revokeObjectURL(url);
}

function formatCurrency(val: number) {
  return Number(val).toLocaleString('en-MY', { minimumFractionDigits: 2 });
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('en-MY', { day: '2-digit', month: 'short', year: 'numeric' });
}
</script>
