<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Audit Logs</h1>

    <div class="bg-white rounded-xl shadow overflow-hidden">
      <div v-if="loading" class="p-8 text-center text-gray-500">Loading...</div>

      <div v-else>
        <table class="w-full text-sm">
          <thead class="bg-gray-50 border-b">
            <tr>
              <th class="px-4 py-3 text-left font-medium text-gray-600">Time</th>
              <th class="px-4 py-3 text-left font-medium text-gray-600">User</th>
              <th class="px-4 py-3 text-left font-medium text-gray-600">Action</th>
              <th class="px-4 py-3 text-left font-medium text-gray-600">Type</th>
              <th class="px-4 py-3 text-left font-medium text-gray-600">ID</th>
              <th class="px-4 py-3 text-left font-medium text-gray-600">IP</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-gray-50">
              <td class="px-4 py-3 text-gray-500">{{ formatDate(log.created_at) }}</td>
              <td class="px-4 py-3">
                <span v-if="log.user" class="font-medium text-gray-800">{{ log.user.name }}</span>
                <span v-else class="text-gray-400">—</span>
              </td>
              <td class="px-4 py-3">
                <span :class="actionClass(log.action)" class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ log.action }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-600">{{ log.auditable_type }}</td>
              <td class="px-4 py-3 text-gray-500">#{{ log.auditable_id }}</td>
              <td class="px-4 py-3 text-gray-400 font-mono text-xs">{{ log.ip_address || '—' }}</td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div class="flex items-center justify-between px-4 py-3 border-t">
          <p class="text-sm text-gray-500">
            Page {{ currentPage }} of {{ lastPage }} ({{ total }} total)
          </p>
          <div class="flex gap-2">
            <button
              @click="loadPage(currentPage - 1)"
              :disabled="currentPage === 1"
              class="px-3 py-1 text-sm border rounded hover:bg-gray-50 disabled:opacity-40"
            >Previous</button>
            <button
              @click="loadPage(currentPage + 1)"
              :disabled="currentPage === lastPage"
              class="px-3 py-1 text-sm border rounded hover:bg-gray-50 disabled:opacity-40"
            >Next</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import api from '../../utils/api';

const logs = ref<any[]>([]);
const loading = ref(true);
const currentPage = ref(1);
const lastPage = ref(1);
const total = ref(0);

async function loadPage(page: number) {
  if (page < 1 || page > lastPage.value) return;
  loading.value = true;
  try {
    const response = await api.get('/api/audit-logs', { params: { page } });
    logs.value = response.data.data;
    currentPage.value = response.data.current_page;
    lastPage.value = response.data.last_page;
    total.value = response.data.total;
  } finally {
    loading.value = false;
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString();
}

function actionClass(action: string) {
  const map: Record<string, string> = {
    created: 'bg-green-100 text-green-700',
    updated: 'bg-blue-100 text-blue-700',
    deleted: 'bg-red-100 text-red-700',
    approved: 'bg-emerald-100 text-emerald-700',
    rejected: 'bg-rose-100 text-rose-700',
    login: 'bg-indigo-100 text-indigo-700',
    logout: 'bg-gray-100 text-gray-700',
  };
  return map[action] || 'bg-gray-100 text-gray-700';
}

onMounted(() => loadPage(1));
</script>
