<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Job Requests</h1>
    </div>

    <div v-if="store.loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else class="bg-white rounded-xl shadow overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Client</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Stage</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Sales</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CS</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">SLA</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="store.jobRequests.length === 0">
            <td colspan="7" class="px-6 py-10 text-center text-gray-400">No job requests found.</td>
          </tr>
          <tr
            v-for="job in store.jobRequests"
            :key="job.id"
            class="hover:bg-gray-50 transition-colors"
          >
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="font-medium text-gray-900">{{ job.client?.company_name || '—' }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="statusBadge(job.status)" class="px-2 py-1 rounded-full text-xs font-medium">
                {{ formatStatus(job.status) }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ job.current_stage === 1 ? 'Sales Tasks' : 'CS Tasks' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ job.assigned_sales?.name || '—' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ job.assigned_cs?.name || '—' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <SLACountdown :sla="job.sla" />
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right">
              <a
                :href="`/job-requests/${job.id}`"
                class="text-blue-600 hover:text-blue-800 text-sm font-medium"
              >
                View →
              </a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useJobRequestsStore } from '../../stores/jobRequests';
import { useAuthStore } from '../../stores/auth';
import SLACountdown from './SLACountdown.vue';

const store = useJobRequestsStore();
const auth = useAuthStore();

onMounted(async () => {
  await auth.fetchUser();
  await store.fetchJobRequests();
});

function statusBadge(status: string) {
  return {
    'bg-gray-100 text-gray-700': status === 'pending',
    'bg-yellow-100 text-yellow-800': status === 'client_pending' || status === 'pending_to_owner',
    'bg-green-100 text-green-800': status === 'completed',
  };
}

function formatStatus(status: string) {
  return {
    pending: 'Pending',
    client_pending: 'Client Pending',
    pending_to_owner: 'Pending Approval',
    completed: 'Completed',
  }[status] ?? status;
}
</script>
