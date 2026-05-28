<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Client Tracking — Team</h1>
        <p class="text-sm text-gray-500 mt-0.5">Staff assigned to this project and their CRM roles.</p>
      </div>
      <button
        v-if="auth.hasRole('admin')"
        @click="showAssignModal = true; loadUnassigned()"
        class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 text-sm font-medium"
      >
        + Assign Staff
      </button>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
    </div>

    <div v-else class="bg-white rounded-xl shadow overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CRM Roles</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th v-if="auth.hasRole('admin')" class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="members.length === 0">
            <td colspan="5" class="px-6 py-10 text-center text-gray-400">
              No staff assigned to Client Tracking yet.
              <span v-if="auth.hasRole('admin')" class="block mt-1 text-sm">
                Use <strong>+ Assign Staff</strong> to add from workspace.
              </span>
            </td>
          </tr>
          <tr v-for="m in members" :key="m.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap font-medium text-gray-900">{{ m.user.name }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{{ m.user.email }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div v-if="editingId === m.user.id" class="flex items-center gap-2 flex-wrap">
                <RolePicker v-model="editingRoles" :disableAdmin="!auth.hasRole('admin')" />
                <button @click="saveRole(m.user.id)" class="text-green-600 text-xs font-medium">Save</button>
                <button @click="editingId = null" class="text-gray-400 text-xs">Cancel</button>
              </div>
              <div v-else class="flex flex-wrap gap-1">
                <span
                  v-for="role in m.user.role"
                  :key="role"
                  :class="roleBadgeClass(role)"
                  class="px-2 py-0.5 rounded-full text-xs font-semibold"
                >{{ role }}</span>
                <span v-if="m.user.role.length === 0" class="text-xs text-gray-400 italic">no roles</span>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span
                :class="m.user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-700'"
                class="px-2 py-1 rounded-full text-xs font-medium"
              >{{ m.user.is_active ? 'Active' : 'Inactive' }}</span>
            </td>
            <td v-if="auth.hasRole('admin')" class="px-6 py-4 whitespace-nowrap text-right">
              <div class="flex justify-end gap-2">
                <button
                  v-if="editingId !== m.user.id"
                  @click="startEdit(m.user)"
                  class="text-indigo-600 hover:text-indigo-800 text-xs font-medium"
                >Edit Roles</button>
                <button
                  v-if="m.user.id !== auth.user?.id"
                  @click="remove(m)"
                  class="text-red-500 hover:text-red-700 text-xs font-medium"
                >Remove</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Assign Staff Modal -->
    <div v-if="showAssignModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md mx-4">
        <h2 class="text-base font-semibold mb-1">Assign to Client Tracking</h2>
        <p class="text-sm text-gray-500 mb-4">Pick a workspace user to add to this project.</p>
        <div class="space-y-2 max-h-64 overflow-y-auto mb-4">
          <label
            v-for="user in unassigned"
            :key="user.id"
            class="flex items-center gap-3 p-2.5 rounded-lg border cursor-pointer hover:bg-indigo-50"
            :class="assignId === user.id ? 'border-indigo-400 bg-indigo-50' : 'border-gray-200'"
          >
            <input type="radio" :value="user.id" v-model="assignId" class="text-indigo-600" />
            <div>
              <p class="text-sm font-medium text-gray-800">{{ user.name }}</p>
              <p class="text-xs text-gray-500">{{ user.email }}</p>
            </div>
          </label>
          <p v-if="unassigned.length === 0" class="text-sm text-gray-400 text-center py-4">
            All workspace staff are already in this project.
          </p>
        </div>
        <div v-if="assignError" class="text-red-600 text-sm bg-red-50 px-3 py-2 rounded mb-3">{{ assignError }}</div>
        <div class="flex gap-3">
          <button
            @click="doAssign"
            :disabled="!assignId || assigning"
            class="flex-1 bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
          >{{ assigning ? 'Assigning…' : 'Assign' }}</button>
          <button @click="showAssignModal = false; assignId = null" class="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg text-sm">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useStaffStore } from '../../stores/staff';
import { useAuthStore } from '../../stores/auth';
import type { ProjectMember, StaffUser } from '../../stores/staff';
import RolePicker from './RolePicker.vue';

const staffStore = useStaffStore();
const auth = useAuthStore();

const loading = ref(true);
const members = ref<ProjectMember[]>([]);
const crmProjectId = ref<number | null>(null);

const editingId = ref<number | null>(null);
const editingRoles = ref<string[]>([]);

const showAssignModal = ref(false);
const unassigned = ref<StaffUser[]>([]);
const assignId = ref<number | null>(null);
const assigning = ref(false);
const assignError = ref('');

onMounted(async () => {
  await auth.fetchUser();
  await load();
});

async function load() {
  loading.value = true;
  try {
    const projects = await staffStore.fetchProjects();
    const crm = projects.find((p) => p.slug === 'crm');
    if (!crm) return;
    crmProjectId.value = crm.id;
    members.value = await staffStore.fetchProjectStaff(crm.id);
  } finally {
    loading.value = false;
  }
}

async function loadUnassigned() {
  await staffStore.fetchStaff();
  const assignedIds = new Set(members.value.map((m) => m.user_id));
  unassigned.value = staffStore.staff.filter((s) => !assignedIds.has(s.id));
}

function startEdit(user: any) {
  editingId.value = user.id;
  editingRoles.value = [...user.role];
}

async function saveRole(userId: number) {
  await staffStore.updateRole(userId, editingRoles.value);
  const m = members.value.find((x) => x.user_id === userId);
  if (m) m.user.role = [...editingRoles.value];
  editingId.value = null;
}

async function remove(m: ProjectMember) {
  if (!confirm(`Remove ${m.user.name} from Client Tracking?`)) return;
  if (!crmProjectId.value) return;
  await staffStore.removeFromProject(crmProjectId.value, m.user_id);
  members.value = members.value.filter((x) => x.id !== m.id);
}

async function doAssign() {
  if (!assignId.value || !crmProjectId.value) return;
  assigning.value = true;
  assignError.value = '';
  try {
    const member = await staffStore.addToProject(crmProjectId.value, assignId.value);
    members.value.push(member);
    showAssignModal.value = false;
    assignId.value = null;
  } catch (e: any) {
    assignError.value = e.response?.data?.message || 'Failed to assign.';
  } finally {
    assigning.value = false;
  }
}

function roleBadgeClass(role: string) {
  return {
    'bg-blue-100 text-blue-800': role === 'admin',
    'bg-sky-100 text-sky-800': role === 'support',
    'bg-green-100 text-green-800': role === 'sales',
    'bg-orange-100 text-orange-800': role === 'cs',
    'bg-teal-100 text-teal-800': role === 'cs_manager',
  };
}
</script>
