<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5">
    <h3 class="font-semibold text-gray-800 mb-3">Account Status</h3>
    <div class="flex items-center gap-3 flex-wrap">
      <span :class="statusBadge(client.account_status)" class="px-3 py-1 rounded-full text-sm font-semibold">
        {{ client.account_status }}
      </span>
      <span v-if="client.pending_account_status" class="text-sm text-orange-600">
        → Deactivation pending admin approval
      </span>
    </div>

    <div class="flex flex-wrap gap-2 mt-4">
      <!-- Sales: activate (inactive → active, needs stage 2 done) -->
      <button
        v-if="auth.hasAnyRole(['sales', 'admin']) && client.account_status === 'inactive' && !client.pending_account_status"
        @click="activate"
        :disabled="loading"
        class="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 disabled:opacity-50 text-sm"
      >
        Activate Account
      </button>

      <!-- Sales: unpause (paused → active) -->
      <button
        v-if="auth.hasAnyRole(['sales', 'admin']) && client.account_status === 'paused'"
        @click="activate"
        :disabled="loading"
        class="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 disabled:opacity-50 text-sm"
      >
        Unpause
      </button>

      <!-- Sales: pause (active → paused) -->
      <button
        v-if="auth.hasAnyRole(['sales', 'admin']) && client.account_status === 'active'"
        @click="pause"
        :disabled="loading"
        class="bg-yellow-500 text-white px-4 py-2 rounded-lg hover:bg-yellow-600 disabled:opacity-50 text-sm"
      >
        Pause
      </button>

      <!-- Sales: request deactivation -->
      <button
        v-if="auth.hasAnyRole(['sales', 'admin']) && (client.account_status === 'active' || client.account_status === 'paused') && !client.pending_account_status"
        @click="requestDeactivate"
        :disabled="loading"
        class="bg-red-100 text-red-700 px-4 py-2 rounded-lg hover:bg-red-200 disabled:opacity-50 text-sm"
      >
        Request Deactivation
      </button>

      <!-- Admin: approve/reject deactivation -->
      <template v-if="auth.hasRole('admin') && client.pending_account_status === 'inactive'">
        <button
          @click="approveDeactivate"
          :disabled="loading"
          class="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 disabled:opacity-50 text-sm"
        >
          Approve Deactivation
        </button>
        <button
          @click="rejectDeactivate"
          :disabled="loading"
          class="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300 disabled:opacity-50 text-sm"
        >
          Reject Deactivation
        </button>
      </template>
    </div>

    <div v-if="error" class="text-red-600 text-sm mt-2">{{ error }}</div>
    <div v-if="message" class="text-green-600 text-sm mt-2">{{ message }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useClientsStore } from '../../stores/clients';
import { useAuthStore } from '../../stores/auth';
import type { Client } from '../../stores/clients';

const props = defineProps<{ client: Client }>();
const emit = defineEmits<{ refresh: [] }>();

const clientsStore = useClientsStore();
const auth = useAuthStore();
const loading = ref(false);
const error = ref('');
const message = ref('');

async function doAction(fn: () => Promise<any>) {
  loading.value = true;
  error.value = '';
  message.value = '';
  try {
    const res = await fn();
    message.value = res?.message || 'Done.';
    emit('refresh');
    setTimeout(() => { message.value = ''; }, 3000);
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Action failed.';
  } finally {
    loading.value = false;
  }
}

const activate = () => doAction(() => clientsStore.activateClient(props.client.id));
const pause = () => doAction(() => clientsStore.pauseClient(props.client.id));
const requestDeactivate = () => doAction(() => clientsStore.requestDeactivate(props.client.id));
const approveDeactivate = () => doAction(() => clientsStore.approveDeactivate(props.client.id));
const rejectDeactivate = () => doAction(() => clientsStore.rejectDeactivate(props.client.id));

function statusBadge(status: string) {
  return {
    'bg-green-100 text-green-800': status === 'active',
    'bg-yellow-100 text-yellow-800': status === 'paused',
    'bg-gray-100 text-gray-700': status === 'inactive',
  };
}
</script>
