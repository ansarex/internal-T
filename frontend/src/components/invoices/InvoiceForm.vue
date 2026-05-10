<template>
  <form @submit.prevent="submit" class="space-y-4">
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Client *</label>
      <select
        v-model="form.client_id"
        required
        @change="onClientChange"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option value="">— Select Client —</option>
        <option v-for="c in activeClients" :key="c.client_id" :value="c.client_id">
          {{ c.company_name }}
        </option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Amount (RM) *</label>
      <input
        v-model.number="form.amount"
        type="number"
        step="0.01"
        min="0"
        required
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder="5000.00"
      />
      <div v-if="form.amount" class="text-xs text-gray-500 mt-1">
        Commission: RM {{ formatCurrency(form.amount * 0.1) }} each (Sales & CS)
      </div>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Billing Month *</label>
      <input
        v-model="billingMonthInput"
        type="month"
        required
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
      <textarea
        v-model="form.notes"
        rows="2"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      ></textarea>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Invoice PDF (optional)</label>
      <input ref="fileInput" type="file" accept=".pdf" class="w-full text-sm" />
    </div>
    <div v-if="error" class="text-red-600 text-sm bg-red-50 px-3 py-2 rounded">{{ error }}</div>
    <div class="flex gap-3">
      <button type="submit" :disabled="loading" class="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium">
        {{ loading ? 'Creating...' : 'Create Invoice' }}
      </button>
      <button type="button" @click="$emit('cancel')" class="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg hover:bg-gray-200 text-sm font-medium">
        Cancel
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useInvoicesStore } from '../../stores/invoices';

const emit = defineEmits<{ created: []; cancel: [] }>();
const invoicesStore = useInvoicesStore();

const activeClients = ref<any[]>([]);
const loading = ref(false);
const error = ref('');
const fileInput = ref<HTMLInputElement | null>(null);

const today = new Date();
const billingMonthInput = ref(`${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}`);

const form = ref({
  client_id: '' as number | '',
  job_request_id: 0,
  amount: null as number | null,
  notes: '',
});

onMounted(async () => {
  const month = billingMonthInput.value + '-01';
  activeClients.value = await invoicesStore.fetchActiveClients(month);
});

function onClientChange() {
  const client = activeClients.value.find((c) => c.client_id === form.value.client_id);
  if (client) {
    form.value.job_request_id = client.job_request_id;
    if (client.monthly_recurring) {
      form.value.amount = Number(client.monthly_recurring);
    }
  }
}

async function submit() {
  if (!form.value.client_id || !form.value.amount || !form.value.job_request_id) return;
  loading.value = true;
  error.value = '';
  try {
    await invoicesStore.createInvoice({
      client_id: form.value.client_id as number,
      job_request_id: form.value.job_request_id,
      amount: form.value.amount,
      billing_month: billingMonthInput.value + '-01',
      notes: form.value.notes || undefined,
      file: fileInput.value?.files?.[0],
    });
    emit('created');
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to create invoice.';
  } finally {
    loading.value = false;
  }
}

function formatCurrency(val: number) {
  return Number(val).toLocaleString('en-MY', { minimumFractionDigits: 2 });
}
</script>
