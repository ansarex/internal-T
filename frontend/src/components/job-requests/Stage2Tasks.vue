<template>
  <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
    <!-- Header -->
    <div class="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
      <h3 class="font-semibold text-gray-800">CS Tasks — Onboarding</h3>
      <span v-if="allDone" class="inline-flex items-center gap-1.5 text-xs font-medium text-green-700 bg-green-100 px-2.5 py-1 rounded-full">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/>
        </svg>
        All tasks completed — locked
      </span>
    </div>

    <!-- Task List -->
    <div class="divide-y divide-gray-100">
      <div
        v-for="task in tasks"
        :key="task.id"
        class="px-5 py-4"
        :class="taskBg(task.status)"
      >
        <div class="flex items-start gap-3">
          <!-- Status Icon -->
          <div class="mt-0.5 shrink-0 w-5 h-5 rounded-full flex items-center justify-center"
            :class="{
              'bg-green-500': task.status === 'completed',
              'bg-yellow-400': task.status === 'in_progress',
              'bg-blue-400': task.status === 'pending_on_client',
              'bg-gray-200': task.status === 'pending',
            }"
          >
            <svg v-if="task.status === 'completed'" class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/>
            </svg>
            <svg v-else-if="task.status === 'in_progress'" class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <svg v-else-if="task.status === 'pending_on_client'" class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <!-- Task name + badge -->
            <div class="flex items-center justify-between gap-2 mb-1">
              <span class="text-sm font-medium text-gray-800">{{ formatTaskType(task.task_type) }}</span>
              <span :class="taskBadge(task.status)" class="shrink-0 px-2 py-0.5 rounded-full text-xs font-medium capitalize">
                {{ formatStatus(task.status) }}
              </span>
            </div>

            <!-- Read-only remarks + audit (always visible) -->
            <div v-if="task.remarks && !canEdit" class="text-xs text-gray-500 mb-1">
              {{ task.remarks }}
            </div>
            <div v-if="task.updated_by_user" class="text-xs text-gray-400">
              Updated by {{ task.updated_by_user.name }}
              <span v-if="task.completed_at && task.status === 'completed'">
                · {{ formatDate(task.completed_at) }}
              </span>
            </div>

            <!-- Edit controls (hidden when all done) -->
            <div v-if="canEdit" class="mt-2.5 space-y-2">
              <div class="flex items-center gap-2">
                <select
                  v-model="taskEdits[task.id].status"
                  class="border border-gray-200 rounded-md px-2.5 py-1.5 text-xs text-gray-700 bg-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                >
                  <option value="in_progress">In Progress</option>
                  <option value="pending_on_client">Pending on Client</option>
                  <option value="completed">Completed</option>
                </select>
                <button
                  @click="saveTask(task.id)"
                  :disabled="savingId === task.id"
                  class="text-xs bg-indigo-600 hover:bg-indigo-700 text-white px-3 py-1.5 rounded-md font-medium disabled:opacity-50 transition-colors"
                >
                  {{ savingId === task.id ? 'Saving…' : 'Save' }}
                </button>
              </div>
              <textarea
                v-model="taskEdits[task.id].remarks"
                rows="2"
                class="w-full border border-gray-200 rounded-md px-2.5 py-1.5 text-xs text-gray-700 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none"
                placeholder="Add remarks…"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Locked footer -->
    <div v-if="allDone" class="px-5 py-3 bg-green-50 border-t border-green-100">
      <p class="text-xs text-green-700">
        All 6 CS tasks have been completed. This section is now read-only.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useJobRequestsStore } from '../../stores/jobRequests';
import { useAuthStore } from '../../stores/auth';
import type { Task, JobRequest } from '../../stores/jobRequests';

const props = defineProps<{ job: JobRequest; tasks: Task[] }>();
const emit = defineEmits<{ updated: [] }>();

const jobStore = useJobRequestsStore();
const auth = useAuthStore();

const allDone = computed(() => props.job.status === 'completed');
const canEdit = computed(() =>
  auth.hasAnyRole(['cs', 'cs_manager', 'admin']) && props.job.current_stage === 2 && !allDone.value
);

const savingId = ref<number | null>(null);

type TaskEdit = { status: string; remarks: string };
const taskEdits = ref<Record<number, TaskEdit>>({});

watch(
  () => props.tasks,
  (tasks) => {
    tasks.forEach((t) => {
      taskEdits.value[t.id] = {
        status: t.status === 'pending' ? 'in_progress' : t.status,
        remarks: t.remarks ?? '',
      };
    });
  },
  { immediate: true }
);

async function saveTask(taskId: number) {
  const edit = taskEdits.value[taskId];
  if (!edit) return;
  savingId.value = taskId;
  try {
    await jobStore.updateTask(taskId, {
      status: edit.status,
      remarks: edit.remarks || undefined,
    });
    emit('updated');
  } finally {
    savingId.value = null;
  }
}

function formatTaskType(type: string) {
  return (
    {
      verify_details: 'Verify Details',
      business_flow: 'Business Flow',
      crm: 'CRM Setup',
      business_accelerator: 'Business Accelerator',
      database_reactive: 'Database Reactive',
      onboarding: 'Onboarding',
    }[type] ?? type
  );
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-MY', { day: 'numeric', month: 'short', year: 'numeric' });
}

function formatStatus(status: string) {
  return (
    {
      pending: 'Pending',
      in_progress: 'In Progress',
      pending_on_client: 'Pending on Client',
      completed: 'Completed',
    }[status] ?? status
  );
}

function taskBg(status: string) {
  return {
    'bg-green-50/60': status === 'completed',
    'bg-yellow-50/40': status === 'in_progress',
    'bg-blue-50/40': status === 'pending_on_client',
    '': status === 'pending',
  };
}

function taskBadge(status: string) {
  return {
    'bg-green-100 text-green-700': status === 'completed',
    'bg-yellow-100 text-yellow-700': status === 'in_progress',
    'bg-blue-100 text-blue-700': status === 'pending_on_client',
    'bg-gray-100 text-gray-500': status === 'pending',
  };
}
</script>
