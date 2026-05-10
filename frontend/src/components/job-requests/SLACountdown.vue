<template>
  <div class="flex items-center gap-2">
    <span :class="dotClass" class="inline-block w-2.5 h-2.5 rounded-full flex-shrink-0"></span>
    <span :class="textClass" class="text-sm font-medium">
      <template v-if="sla?.sla_overdue">
        Overdue by {{ Math.abs(sla.days_remaining) }}d
      </template>
      <template v-else-if="sla?.indicator === 'green'">
        Completed
      </template>
      <template v-else-if="sla">
        {{ countdown }}
      </template>
      <template v-else>—</template>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import type { SLAStatus } from '../../stores/jobRequests';

const props = defineProps<{ sla?: SLAStatus | null }>();

const now = ref(new Date());
let timer: ReturnType<typeof setInterval> | null = null;

onMounted(() => {
  timer = setInterval(() => { now.value = new Date(); }, 1000);
});

onUnmounted(() => {
  if (timer) clearInterval(timer);
});

const dotClass = computed(() => {
  const ind = props.sla?.indicator;
  if (ind === 'green') return 'bg-green-500';
  if (ind === 'red') return 'bg-red-500';
  if (ind === 'yellow') return 'bg-yellow-500';
  return 'bg-gray-400';
});

const textClass = computed(() => {
  const ind = props.sla?.indicator;
  if (ind === 'green') return 'text-green-600';
  if (ind === 'red') return 'text-red-600';
  if (ind === 'yellow' && (props.sla?.days_remaining ?? 99) <= 3) return 'text-orange-600';
  return 'text-gray-600';
});

const countdown = computed(() => {
  if (!props.sla?.sla_deadline) return '—';
  const deadline = new Date(props.sla.sla_deadline);
  const diff = deadline.getTime() - now.value.getTime();
  if (diff <= 0) return 'Overdue';
  const days = Math.floor(diff / 86400000);
  const hours = Math.floor((diff % 86400000) / 3600000);
  const mins = Math.floor((diff % 3600000) / 60000);
  const secs = Math.floor((diff % 60000) / 1000);
  if (days > 0) return `${days}d ${hours}h ${mins}m`;
  return `${hours}h ${mins}m ${secs}s`;
});
</script>
