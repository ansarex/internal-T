<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Dashboard</h1>

    <div v-if="loading" class="text-gray-500">Loading...</div>

    <div v-else>
      <!-- Stat Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Jobs" :value="data.summary.total_jobs" icon-bg="bg-blue-100" value-color="text-blue-700">
          <template #icon><svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg></template>
        </StatCard>
        <StatCard label="Active Clients" :value="data.summary.active_clients" icon-bg="bg-green-100" value-color="text-green-700">
          <template #icon><svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/></svg></template>
        </StatCard>
        <StatCard label="Overdue Jobs" :value="data.summary.overdue_jobs" icon-bg="bg-red-100" value-color="text-red-600">
          <template #icon><svg class="w-6 h-6 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg></template>
        </StatCard>
        <StatCard label="Pending Approvals" :value="data.summary.pending_approvals" icon-bg="bg-yellow-100" value-color="text-yellow-700">
          <template #icon><svg class="w-6 h-6 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg></template>
        </StatCard>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-8">
        <StatCard label="Stale Jobs" :value="data.summary.stale_jobs" icon-bg="bg-orange-100" value-color="text-orange-600" subtitle=">3 days inactive" />
        <StatCard label="Stuck in CS Tasks" :value="data.summary.stuck_stage2_jobs" icon-bg="bg-purple-100" value-color="text-purple-700" />
        <StatCard label="Missing Fields" :value="data.summary.missing_fields_jobs" icon-bg="bg-gray-100" value-color="text-gray-700" subtitle="Sales Tasks incomplete" />
      </div>

      <!-- Flag Sections -->
      <div class="space-y-6">
        <!-- Overdue Jobs -->
        <FlagSection title="Overdue Jobs" color="red" :count="data.overdue_jobs.length">
          <JobTable :jobs="data.overdue_jobs" />
        </FlagSection>

        <!-- Pending Approvals -->
        <FlagSection title="Pending Agreement Approvals" color="yellow" :count="data.pending_approvals.length">
          <div class="space-y-2">
            <div v-for="ag in data.pending_approvals" :key="ag.id" class="flex items-center justify-between p-3 bg-yellow-50 rounded-lg border border-yellow-200">
              <div>
                <p class="font-medium text-sm">{{ ag.job_request?.client?.company_name || 'Unknown Client' }}</p>
                <p class="text-xs text-gray-500">{{ ag.type === 'service_agreement' ? 'Service Agreement' : 'NDA' }} v{{ ag.version }}</p>
              </div>
              <a :href="`/job-requests/${ag.job_request_id}`" class="text-xs text-blue-600 hover:underline">View</a>
            </div>
          </div>
        </FlagSection>

        <!-- Stale Jobs -->
        <FlagSection title="Stale Jobs (No activity > 3 days)" color="orange" :count="data.stale_jobs.length">
          <JobTable :jobs="data.stale_jobs" />
        </FlagSection>

        <!-- Stuck CS Tasks -->
        <FlagSection title="Stuck in CS Tasks" color="purple" :count="data.stuck_tasks.length">
          <JobTable :jobs="data.stuck_tasks" />
        </FlagSection>

        <!-- Missing Fields -->
        <FlagSection title="Missing Sales Task Fields" color="gray" :count="data.missing_fields.length">
          <JobTable :jobs="data.missing_fields" />
        </FlagSection>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineComponent, h } from 'vue';
import api from '../../utils/api';
import StatCard from './StatCard.vue';

const loading = ref(true);
const data = ref({
  summary: {
    total_jobs: 0,
    active_clients: 0,
    overdue_jobs: 0,
    pending_approvals: 0,
    stale_jobs: 0,
    stuck_stage2_jobs: 0,
    missing_fields_jobs: 0,
  },
  overdue_jobs: [] as any[],
  pending_approvals: [] as any[],
  stale_jobs: [] as any[],
  stuck_tasks: [] as any[],
  missing_fields: [] as any[],
});

const FlagSection = defineComponent({
  props: { title: String, color: String, count: Number },
  setup(props, { slots }) {
    const colorMap: Record<string, string> = {
      red: 'text-red-700 bg-red-50 border-red-200',
      yellow: 'text-yellow-700 bg-yellow-50 border-yellow-200',
      orange: 'text-orange-700 bg-orange-50 border-orange-200',
      purple: 'text-purple-700 bg-purple-50 border-purple-200',
      gray: 'text-gray-700 bg-gray-50 border-gray-200',
    };
    const colors = colorMap[props.color || 'gray'];
    return () =>
      props.count === 0
        ? null
        : h('div', { class: `border rounded-xl overflow-hidden ${colors}` }, [
            h('div', { class: `px-4 py-3 border-b flex items-center gap-2 ${colors}` }, [
              h('h3', { class: 'font-semibold text-sm' }, props.title),
              h('span', { class: 'ml-auto text-xs font-bold' }, `(${props.count})`),
            ]),
            h('div', { class: 'p-4 bg-white' }, slots.default?.()),
          ]);
  },
});

const JobTable = defineComponent({
  props: { jobs: Array as () => any[] },
  setup(props) {
    return () =>
      h('div', { class: 'overflow-x-auto' }, [
        h('table', { class: 'w-full text-sm' }, [
          h('thead', {}, [
            h('tr', { class: 'border-b text-gray-500' }, [
              h('th', { class: 'text-left py-2 pr-4 font-medium' }, 'Client'),
              h('th', { class: 'text-left py-2 pr-4 font-medium' }, 'Tasks'),
              h('th', { class: 'text-left py-2 pr-4 font-medium' }, 'Status'),
              h('th', { class: 'text-left py-2 font-medium' }, 'Action'),
            ]),
          ]),
          h('tbody', {}, [
            ...(props.jobs || []).map((job: any) =>
              h('tr', { key: job.id, class: 'border-b hover:bg-gray-50' }, [
                h('td', { class: 'py-2 pr-4' }, job.client?.company_name || `Client #${job.client_id}`),
                h('td', { class: 'py-2 pr-4 text-gray-500' }, job.current_stage === 1 ? 'Sales Tasks' : 'CS Tasks'),
                h('td', { class: 'py-2 pr-4 text-gray-500' }, job.status),
                h('td', { class: 'py-2' }, [
                  h('a', { href: `/job-requests/${job.id}`, class: 'text-blue-600 hover:underline text-xs' }, 'View'),
                ]),
              ])
            ),
          ]),
        ]),
      ]);
  },
});

onMounted(async () => {
  try {
    const response = await api.get('/api/dashboard');
    const d = response.data;
    data.value = {
      summary: d.summary ?? data.value.summary,
      overdue_jobs: d.overdue_jobs ?? [],
      pending_approvals: d.pending_approvals ?? [],
      stale_jobs: d.stale_jobs ?? [],
      stuck_tasks: d.stuck_tasks ?? [],
      missing_fields: d.missing_fields ?? [],
    };
  } catch (e) {
    console.error('Dashboard load error:', e);
  } finally {
    loading.value = false;
  }
});
</script>
