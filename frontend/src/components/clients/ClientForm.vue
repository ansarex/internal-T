<template>
  <form @submit.prevent="submit" class="space-y-4">
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Company Name *</label>
      <input
        v-model="form.company_name"
        type="text"
        required
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder="Acme Corp"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Todo List / Notes</label>
      <textarea
        v-model="form.todo_list"
        rows="3"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        placeholder="Optional notes..."
      ></textarea>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Assign Sales PIC</label>
      <select
        v-model="form.assigned_sales_id"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option value="">— Select Sales —</option>
        <option v-for="s in salesStaff" :key="s.id" :value="s.id">{{ s.name }}</option>
      </select>
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Assign CS PIC</label>
      <select
        v-model="form.assigned_cs_id"
        class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option value="">— Select CS —</option>
        <option v-for="s in csStaff" :key="s.id" :value="s.id">{{ s.name }}</option>
      </select>
    </div>
    <div v-if="error" class="text-red-600 text-sm bg-red-50 px-3 py-2 rounded">{{ error }}</div>
    <div class="flex gap-3 pt-2">
      <button
        type="submit"
        :disabled="loading"
        class="flex-1 bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
      >
        {{ loading ? 'Creating...' : 'Create Client' }}
      </button>
      <button
        type="button"
        @click="$emit('cancel')"
        class="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg hover:bg-gray-200 text-sm font-medium"
      >
        Cancel
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useClientsStore } from '../../stores/clients';
import api from '../../utils/api';

const emit = defineEmits<{ created: []; cancel: [] }>();
const clientsStore = useClientsStore();

const salesStaff = ref<any[]>([]);
const csStaff = ref<any[]>([]);
const loading = ref(false);
const error = ref('');

const form = ref({
  company_name: '',
  todo_list: '',
  assigned_sales_id: '' as number | '',
  assigned_cs_id: '' as number | '',
});

onMounted(async () => {
  const [salesRes, csRes] = await Promise.all([
    api.get('/api/staff', { params: { role: 'sales' } }),
    api.get('/api/staff', { params: { role: 'cs' } }),
  ]);
  salesStaff.value = salesRes.data;
  csStaff.value = csRes.data;
});

async function submit() {
  if (!form.value.company_name.trim()) return;
  loading.value = true;
  error.value = '';
  try {
    await clientsStore.createClient({
      company_name: form.value.company_name,
      todo_list: form.value.todo_list || undefined,
      assigned_sales_id: form.value.assigned_sales_id || undefined,
      assigned_cs_id: form.value.assigned_cs_id || undefined,
    });
    emit('created');
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to create client.';
  } finally {
    loading.value = false;
  }
}
</script>
