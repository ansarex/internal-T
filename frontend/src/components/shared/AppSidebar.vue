<template>
  <aside class="w-60 min-h-screen bg-[#191b1f] text-white flex flex-col select-none shrink-0">

    <!-- Workspace Header -->
    <div class="px-3 py-3 flex items-center gap-2.5 border-b border-white/5">
      <div class="w-7 h-7 bg-indigo-600 rounded-md flex items-center justify-center font-bold text-xs shrink-0">T</div>
      <div class="min-w-0">
        <p class="text-sm font-semibold text-white truncate leading-tight">TWD Group</p>
        <p class="text-[10px] text-gray-500 leading-tight">Internal System</p>
      </div>
    </div>

    <!-- User Profile -->
    <div class="px-3 py-2.5 border-b border-white/5">
      <div class="flex items-center gap-2">
        <div class="w-6 h-6 rounded-full bg-indigo-700 flex items-center justify-center text-[11px] font-semibold shrink-0">
          {{ initials }}
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-xs font-medium text-gray-200 truncate leading-tight">{{ auth.user?.name }}</p>
          <div class="flex flex-wrap gap-0.5 mt-0.5">
            <span
              v-for="role in auth.user?.role"
              :key="role"
              :class="roleColor(role)"
              class="text-[9px] px-1.5 py-px rounded-full font-medium"
            >{{ role }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Spaces Nav -->
    <nav class="flex-1 overflow-y-auto py-2">

      <!-- ── Workspace section (admin only) ── -->
      <div v-if="auth.hasRole('admin')" class="mb-3">
        <div class="px-3 mb-1">
          <span class="text-[10px] font-semibold text-gray-600 uppercase tracking-widest">Workspace</span>
        </div>
        <NavItem href="/workspace/staff" :active="path.startsWith('/workspace/staff')">
          <template #icon>
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
          </template>
          Staff
        </NavItem>
      </div>

      <!-- Section Label -->
      <div class="px-3 mb-1 flex items-center justify-between">
        <span class="text-[10px] font-semibold text-gray-600 uppercase tracking-widest">Projects</span>
      </div>

      <!-- ── Client Tracking Space ── -->
      <div v-if="auth.isInProject('crm')">
        <!-- Space Row -->
        <button
          @click="trackingOpen = !trackingOpen"
          class="w-full flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-white/5 transition-colors group"
        >
          <svg
            class="w-3 h-3 text-gray-500 transition-transform shrink-0"
            :class="trackingOpen ? 'rotate-90' : ''"
            fill="none" stroke="currentColor" viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
          </svg>
          <span class="w-4 h-4 rounded bg-indigo-600 flex items-center justify-center text-[9px] font-bold shrink-0">C</span>
          <span class="font-medium text-gray-200 text-xs">Trustwired CRM</span>
        </button>

        <!-- Space Items -->
        <div v-show="trackingOpen" class="ml-5 border-l border-white/[0.07] pl-2 mt-0.5 space-y-px pb-1">
          <NavItem href="/dashboard" :active="path === '/dashboard'">
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M4 5a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v3a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zm10-1a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
              </svg>
            </template>
            Dashboard
          </NavItem>

          <NavItem href="/clients" :active="path === '/clients'">
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
            </template>
            Clients
          </NavItem>

          <NavItem href="/job-requests" :active="path.startsWith('/job-requests')">
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
              </svg>
            </template>
            Job Requests
          </NavItem>

          <NavItem
            v-if="auth.hasAnyRole(['sales', 'admin'])"
            href="/invoices"
            :active="path === '/invoices'"
          >
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
              </svg>
            </template>
            Invoices
          </NavItem>

          <NavItem
            v-if="auth.hasRole('admin')"
            href="/approvals"
            :active="path === '/approvals'"
          >
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </template>
            Approvals
            <span
              v-if="pendingCount > 0"
              class="ml-auto bg-red-500 text-white text-[9px] font-bold rounded-full px-1.5 py-px min-w-[16px] text-center"
            >{{ pendingCount }}</span>
          </NavItem>

          <NavItem
            href="/staff"
            :active="path === '/staff'"
          >
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
            </template>
            Team
          </NavItem>

          <NavItem
            v-if="auth.hasRole('admin')"
            href="/audit-logs"
            :active="path === '/audit-logs'"
          >
            <template #icon>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M4 6h16M4 10h16M4 14h16M4 18h16"/>
              </svg>
            </template>
            Audit Logs
          </NavItem>
        </div>
      </div>

      <!-- ── Placeholder Spaces ── -->
      <div class="mt-1 space-y-px">
        <div
          v-for="space in placeholderSpaces"
          :key="space.label"
          class="flex items-center gap-2 px-3 py-1.5 text-xs text-gray-600 cursor-not-allowed"
        >
          <svg class="w-3 h-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
          </svg>
          <span class="w-4 h-4 rounded flex items-center justify-center text-[9px] font-bold shrink-0" :style="{ background: space.color }">
            {{ space.initial }}
          </span>
          <span class="font-medium truncate flex-1">{{ space.label }}</span>
          <span class="text-[9px] bg-gray-800 text-gray-600 px-1.5 py-px rounded-full shrink-0">soon</span>
        </div>
      </div>

    </nav>

    <!-- Sign Out -->
    <div class="p-3 border-t border-white/5">
      <button
        @click="auth.logout()"
        class="w-full flex items-center gap-2 px-2 py-1.5 text-xs text-gray-500 hover:text-red-400 hover:bg-red-500/10 rounded-md transition-colors"
      >
        <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
        </svg>
        Sign Out
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, defineComponent, h, onMounted } from 'vue';
import { useAuthStore } from '../../stores/auth';
import api from '../../utils/api';

const auth = useAuthStore();
const pendingCount = ref(0);
const trackingOpen = ref(true);

const path = typeof window !== 'undefined' ? window.location.pathname : '';

const initials = computed(() => {
  const name = auth.user?.name || '';
  return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
});

const placeholderSpaces = [
  { label: 'Investiland', initial: 'I', color: '#374151' },
  { label: 'Addhoc TWD', initial: 'A', color: '#374151' },
];

function roleColor(role: string): string {
  const map: Record<string, string> = {
    admin: 'bg-indigo-600 text-white',
    support: 'bg-sky-600 text-white',
    sales: 'bg-emerald-600 text-white',
    cs: 'bg-orange-500 text-white',
  };
  return map[role] || 'bg-gray-600 text-white';
}

// NavItem inline component
const NavItem = defineComponent({
  props: { href: String, active: Boolean },
  setup(props, { slots }) {
    return () => h('a', {
      href: props.href,
      class: [
        'flex items-center gap-2 px-2 py-1 rounded text-xs font-medium transition-colors',
        props.active
          ? 'bg-indigo-600/20 text-indigo-300'
          : 'text-gray-400 hover:bg-white/5 hover:text-gray-200',
      ].join(' '),
    }, [
      slots.icon?.(),
      slots.default?.(),
    ]);
  },
});

onMounted(async () => {
  if (!auth.user) {
    await auth.fetchUser();
  }
  if (auth.hasRole('admin')) {
    try {
      const res = await api.get('/api/dashboard');
      pendingCount.value = res.data.summary?.pending_approvals || 0;
    } catch {}
  }
});
</script>
