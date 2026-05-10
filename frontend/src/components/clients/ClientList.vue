<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Clients</h1>
      <div class="flex items-center gap-3">
        <span v-if="clientsStore.totalRecurring > 0" class="text-sm text-gray-500">
          Total MRR: <strong class="text-green-600">RM {{ formatCurrency(clientsStore.totalRecurring) }}</strong>
        </span>
        <button
          v-if="auth.hasAnyRole(['support', 'admin'])"
          @click="showForm = true"
          class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium"
        >
          + New Client
        </button>
      </div>
    </div>

    <div v-if="clientsStore.loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else-if="clientsStore.error" class="bg-red-50 text-red-600 p-4 rounded-lg">
      {{ clientsStore.error }}
    </div>

    <div v-else class="bg-white rounded-xl shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Company</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Sales PIC</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CS PIC</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">MRR</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Stage</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="clientsStore.clients.length === 0">
            <td colspan="7" class="px-6 py-10 text-center text-gray-400">No clients found.</td>
          </tr>
          <tr
            v-for="client in clientsStore.clients"
            :key="client.id"
            class="hover:bg-gray-50 transition-colors"
          >
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="font-medium text-gray-900">{{ client.company_name }}</div>
              <div v-if="client.pending_account_status" class="text-xs text-orange-500 mt-0.5">
                Deactivation pending approval
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="statusBadgeClass(client.account_status)" class="px-2 py-1 rounded-full text-xs font-medium">
                {{ client.account_status }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ getJobRequest(client)?.assigned_sales?.name || '—' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ getJobRequest(client)?.assigned_cs?.name || '—' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              {{ getJobRequest(client)?.monthly_recurring ? 'RM ' + formatCurrency(getJobRequest(client)!.monthly_recurring!) : '—' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
              Stage {{ getJobRequest(client)?.current_stage || 1 }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right">
              <a
                :href="`/job-requests/${getJobRequest(client)?.id}`"
                class="text-blue-600 hover:text-blue-800 text-sm font-medium"
              >
                View →
              </a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Client Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md mx-4">
        <h2 class="text-lg font-semibold mb-4">Create New Client</h2>
        <ClientForm @created="onClientCreated" @cancel="showForm = false" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useClientsStore } from '../../stores/clients';
import { useAuthStore } from '../../stores/auth';
import ClientForm from './ClientForm.vue';

const clientsStore = useClientsStore();
const auth = useAuthStore();
const showForm = ref(false);

onMounted(async () => {
  await auth.fetchUser();
  await clientsStore.fetchClients();
});

function getJobRequest(client: any) {
  return client.job_requests?.[0] ?? null;
}

function statusBadgeClass(status: string) {
  return {
    'bg-green-100 text-green-800': status === 'active',
    'bg-yellow-100 text-yellow-800': status === 'paused',
    'bg-gray-100 text-gray-700': status === 'inactive',
  };
}

function formatCurrency(val: number) {
  return Number(val).toLocaleString('en-MY', { minimumFractionDigits: 2 });
}

async function onClientCreated() {
  showForm.value = false;
  await clientsStore.fetchClients();
}
</script>
