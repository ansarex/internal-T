<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5">
    <h3 class="font-semibold text-gray-800 mb-4">Stage 1 — Client Details</h3>
    <form @submit.prevent="save" class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Customer PIC</label>
          <input
            v-model="form.customer_pic"
            type="text"
            :disabled="!canEdit"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500"
            placeholder="Contact person name"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Monthly Recurring (RM)</label>
          <input
            v-model.number="form.monthly_recurring"
            type="number"
            step="0.01"
            min="0"
            :disabled="!canEdit"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500"
            placeholder="5000.00"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Account Type</label>
          <select
            v-model="form.account_type"
            :disabled="!canEdit"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500"
          >
            <option value="">— Select —</option>
            <option value="Standard">Standard</option>
            <option value="Premium">Premium</option>
            <option value="Enterprise">Enterprise</option>
          </select>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Recurring Start Date</label>
          <input
            v-model="form.recurring_start_date"
            type="date"
            :disabled="!canEdit"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500"
          />
          <p class="text-xs text-gray-400 mt-1">
            Customer must pay before the
            <strong v-if="billingDay" class="text-gray-600">{{ billingDay }}{{ ordinal(billingDay) }}</strong>
            <span v-else>this day</span>
            of each month when account is active.
          </p>
        </div>
      </div>

      <div v-if="error" class="text-red-600 text-sm bg-red-50 px-3 py-2 rounded">{{ error }}</div>
      <div v-if="success" class="text-green-600 text-sm bg-green-50 px-3 py-2 rounded">Saved successfully.</div>
      <div v-if="canEdit" class="flex justify-end">
        <button
          type="submit"
          :disabled="saving"
          class="bg-blue-600 text-white px-5 py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
        >
          {{ saving ? 'Saving...' : 'Save Details' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useJobRequestsStore } from '../../stores/jobRequests';
import { useAuthStore } from '../../stores/auth';
import type { JobRequest } from '../../stores/jobRequests';

const props = defineProps<{ job: JobRequest }>();
const emit = defineEmits<{ updated: [] }>();

const jobStore = useJobRequestsStore();
const auth = useAuthStore();

function toDateInput(iso?: string): string {
  if (!iso) return '';
  return iso.split('T')[0];
}

const form = ref({
  customer_pic: props.job.customer_pic || '',
  monthly_recurring: props.job.monthly_recurring || null as number | null,
  account_type: props.job.account_type || '',
  recurring_start_date: toDateInput(props.job.recurring_start_date),
});

watch(() => props.job, (j) => {
  form.value.customer_pic = j.customer_pic || '';
  form.value.monthly_recurring = j.monthly_recurring || null;
  form.value.account_type = j.account_type || '';
  form.value.recurring_start_date = toDateInput(j.recurring_start_date);
});

const billingDay = computed(() => {
  if (!form.value.recurring_start_date) return null;
  const d = new Date(form.value.recurring_start_date + 'T00:00:00');
  return isNaN(d.getTime()) ? null : d.getDate();
});

function ordinal(n: number): string {
  const s = ['th', 'st', 'nd', 'rd'];
  const v = n % 100;
  return s[(v - 20) % 10] || s[v] || s[0];
}

const canEdit = auth.hasAnyRole(['sales', 'admin']) && props.job.current_stage === 1;
const saving = ref(false);
const error = ref('');
const success = ref(false);

async function save() {
  saving.value = true;
  error.value = '';
  success.value = false;
  try {
    await jobStore.updateStage1(props.job.id, {
      customer_pic: form.value.customer_pic || undefined,
      monthly_recurring: form.value.monthly_recurring || undefined,
      account_type: form.value.account_type || undefined,
      recurring_start_date: form.value.recurring_start_date || undefined,
    });
    success.value = true;
    emit('updated');
    setTimeout(() => { success.value = false; }, 3000);
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to save.';
  } finally {
    saving.value = false;
  }
}
</script>
