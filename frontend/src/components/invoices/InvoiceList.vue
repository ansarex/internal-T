<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Invoices</h1>
      <div class="flex items-center gap-3">
        <input
          v-model="selectedMonth"
          type="month"
          class="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
    </div>

    <!-- Admin overview -->
    <div v-if="auth.hasRole('admin')" class="space-y-4">
      <div v-if="overview" class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
        <div class="bg-white rounded-lg border p-3 text-center">
          <div class="text-2xl font-bold text-gray-900">{{ overview.total_clients }}</div>
          <div class="text-xs text-gray-500 mt-1">Active Clients</div>
        </div>
        <div class="bg-white rounded-lg border p-3 text-center">
          <div class="text-2xl font-bold text-green-600">{{ overview.invoiced }}</div>
          <div class="text-xs text-gray-500 mt-1">Invoiced</div>
        </div>
        <div class="bg-white rounded-lg border p-3 text-center">
          <div class="text-2xl font-bold text-red-600">{{ overview.missing }}</div>
          <div class="text-xs text-gray-500 mt-1">Missing</div>
        </div>
        <div class="bg-white rounded-lg border p-3 text-center">
          <div class="text-2xl font-bold text-blue-600">RM {{ formatCurrency(overview.total_amount) }}</div>
          <div class="text-xs text-gray-500 mt-1">Total Amount</div>
        </div>
      </div>

      <div class="bg-white rounded-xl shadow overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Client</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Invoice #</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Sales</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">CS</th>
              <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-if="!overviewClients.length">
              <td colspan="7" class="px-4 py-8 text-center text-gray-400">No active clients for this month.</td>
            </tr>
            <tr v-for="row in overviewClients" :key="row.client_id" class="hover:bg-gray-50">
              <td class="px-4 py-3 text-sm font-medium text-gray-900">
                {{ row.company_name }}
                <span v-if="row.overdue_missing" class="ml-1 text-xs text-red-600">⚠ Overdue</span>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600">{{ row.invoice?.invoice_number || '—' }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">{{ row.invoice ? 'RM ' + formatCurrency(row.invoice.amount) : '—' }}</td>
              <td class="px-4 py-3">
                <span v-if="row.invoice" :class="statusBadge(row.invoice.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ row.invoice.status }}
                </span>
                <span v-else class="text-xs text-gray-400">Not invoiced</span>
              </td>
              <td class="px-4 py-3 text-xs text-gray-500">{{ row.assigned_sales?.name || '—' }}</td>
              <td class="px-4 py-3 text-xs text-gray-500">{{ row.assigned_cs?.name || '—' }}</td>
              <td class="px-4 py-3 text-right">
                <div class="flex justify-end gap-2">
                  <button
                    v-if="row.invoice && row.invoice.status !== 'paid'"
                    @click="markPaid(row.invoice.id)"
                    class="text-xs text-green-600 hover:text-green-800 font-medium"
                  >
                    Mark Paid
                  </button>
                  <button
                    v-if="row.invoice?.file_path"
                    @click="downloadInvoice(row.invoice.id)"
                    class="text-xs text-blue-600 hover:underline"
                  >
                    Download
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Commission summary -->
      <div v-if="commissions" class="bg-white rounded-xl border border-gray-200 p-5">
        <h3 class="font-semibold text-gray-800 mb-3">Commission Summary — {{ selectedMonth }}</h3>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div v-for="entry in commissions" :key="entry.staff_id" class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
            <div>
              <div class="text-sm font-medium text-gray-800">{{ entry.name }}</div>
              <div class="text-xs text-gray-500">{{ entry.role }}</div>
            </div>
            <div class="text-right">
              <div class="text-sm font-semibold text-green-700">RM {{ formatCurrency(entry.commission) }}</div>
              <div class="text-xs text-gray-400">{{ entry.invoice_count }} invoice(s)</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sales view -->
    <div v-else-if="auth.hasAnyRole(['sales'])" class="space-y-4">
      <div class="flex justify-end mb-2">
        <button @click="showCreate = true" class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 text-sm font-medium">
          + Record Invoice
        </button>
      </div>

      <div class="bg-white rounded-xl shadow overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Invoice #</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Client</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Commission</th>
              <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-if="!invoicesStore.invoices.length">
              <td colspan="6" class="px-4 py-8 text-center text-gray-400">No invoices found.</td>
            </tr>
            <tr v-for="inv in invoicesStore.invoices" :key="inv.id" class="hover:bg-gray-50">
              <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ inv.invoice_number }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">{{ inv.client?.company_name }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">RM {{ formatCurrency(inv.amount) }}</td>
              <td class="px-4 py-3">
                <span :class="statusBadge(inv.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ inv.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-sm text-green-700">RM {{ formatCurrency(inv.sales_commission) }}</td>
              <td class="px-4 py-3 text-right">
                <div class="flex justify-end gap-2">
                  <button
                    v-if="!inv.file_path"
                    @click="openUploadFile(inv.id)"
                    class="text-xs text-blue-600 hover:underline"
                  >
                    Upload PDF
                  </button>
                  <button
                    v-else
                    @click="downloadInvoice(inv.id)"
                    class="text-xs text-blue-600 hover:underline"
                  >
                    Download
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Invoice Modal -->
    <div v-if="showCreate" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md mx-4">
        <h2 class="text-lg font-semibold mb-4">Record Invoice</h2>
        <InvoiceForm @created="onInvoiceCreated" @cancel="showCreate = false" />
      </div>
    </div>

    <!-- Upload PDF Modal -->
    <div v-if="uploadModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm mx-4">
        <h3 class="text-base font-semibold mb-3">Upload Invoice PDF</h3>
        <input type="file" ref="fileInputRef" accept=".pdf" class="w-full text-sm mb-3" />
        <div class="flex gap-3">
          <button @click="doUploadFile" :disabled="uploadingFile" class="flex-1 bg-blue-600 text-white py-2 rounded text-sm disabled:opacity-50">
            {{ uploadingFile ? 'Uploading...' : 'Upload' }}
          </button>
          <button @click="uploadModal = false" class="flex-1 bg-gray-100 text-gray-700 py-2 rounded text-sm">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { useInvoicesStore } from '../../stores/invoices';
import { useAuthStore } from '../../stores/auth';
import InvoiceForm from './InvoiceForm.vue';
import api from '../../utils/api';

const invoicesStore = useInvoicesStore();
const auth = useAuthStore();

const selectedMonth = ref(new Date().toISOString().slice(0, 7));
const showCreate = ref(false);
const overview = ref<any>(null);
const overviewClients = ref<any[]>([]);
const commissions = ref<any[]>([]);
const uploadModal = ref(false);
const uploadTargetId = ref<number | null>(null);
const uploadingFile = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);

onMounted(async () => {
  await auth.fetchUser();
  await loadData();
});

watch(selectedMonth, loadData);

async function loadData() {
  const month = selectedMonth.value + '-01';
  if (auth.hasRole('admin')) {
    const [ov, comm] = await Promise.all([
      invoicesStore.fetchAdminOverview(month),
      invoicesStore.fetchCommissions(month),
    ]);
    overview.value = ov.summary;
    overviewClients.value = ov.clients || [];
    commissions.value = comm;
  } else {
    await invoicesStore.fetchInvoices({ month });
    const comm = await invoicesStore.fetchCommissions(month);
    commissions.value = comm;
  }
}

async function markPaid(id: number) {
  await invoicesStore.markPaid(id);
  await loadData();
}

async function downloadInvoice(id: number) {
  const response = await api.get(`/api/invoices/${id}/download`, { responseType: 'blob' });
  const url = window.URL.createObjectURL(response.data);
  const a = document.createElement('a');
  a.href = url;
  a.download = `invoice-${id}.pdf`;
  a.click();
  window.URL.revokeObjectURL(url);
}

function openUploadFile(id: number) {
  uploadTargetId.value = id;
  uploadModal.value = true;
}

async function doUploadFile() {
  const file = fileInputRef.value?.files?.[0];
  if (!file || !uploadTargetId.value) return;
  uploadingFile.value = true;
  try {
    await invoicesStore.uploadFile(uploadTargetId.value, file);
    uploadModal.value = false;
    await loadData();
  } finally {
    uploadingFile.value = false;
  }
}

async function onInvoiceCreated() {
  showCreate.value = false;
  await loadData();
}

function formatCurrency(val: number) {
  return Number(val).toLocaleString('en-MY', { minimumFractionDigits: 2 });
}

function statusBadge(status: string) {
  return {
    'bg-green-100 text-green-800': status === 'paid',
    'bg-yellow-100 text-yellow-800': status === 'pending',
    'bg-red-100 text-red-700': status === 'overdue',
  };
}
</script>
