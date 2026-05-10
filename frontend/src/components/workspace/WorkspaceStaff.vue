<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Workspace Staff</h1>
        <p class="text-sm text-gray-500 mt-0.5">All user accounts across the workspace. Assign to projects to grant access.</p>
      </div>
      <button
        @click="showCreateForm = true"
        class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 text-sm font-medium"
      >
        + New Staff Account
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
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Projects</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-if="allStaff.length === 0">
            <td colspan="5" class="px-6 py-10 text-center text-gray-400">No staff accounts yet.</td>
          </tr>
          <tr v-for="member in allStaff" :key="member.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="font-medium text-gray-900">{{ member.name }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-600">{{ member.email }}</td>
            <td class="px-6 py-4">
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="proj in memberProjects(member.id)"
                  :key="proj.id"
                  class="inline-flex items-center gap-1 bg-indigo-50 text-indigo-700 text-xs px-2 py-0.5 rounded-full font-medium"
                >
                  {{ proj.name }}
                  <button
                    @click="removeProject(proj.projectStaffId, member.name, proj.name)"
                    class="text-indigo-400 hover:text-red-500 ml-0.5 leading-none"
                    title="Remove from project"
                  >×</button>
                </span>
                <button
                  @click="openAssign(member)"
                  class="text-xs text-indigo-500 hover:text-indigo-700 border border-dashed border-indigo-300 px-2 py-0.5 rounded-full hover:border-indigo-500 transition-colors"
                >
                  + Assign
                </button>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span
                :class="member.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-700'"
                class="px-2 py-1 rounded-full text-xs font-medium"
              >{{ member.is_active ? 'Active' : 'Inactive' }}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right">
              <div class="flex justify-end gap-3">
                <button
                  v-if="member.is_active && member.id !== currentUserId"
                  @click="deactivate(member.id)"
                  class="text-orange-600 hover:text-orange-800 text-xs font-medium"
                >Deactivate</button>
                <button
                  v-else-if="!member.is_active"
                  @click="activate(member.id)"
                  class="text-green-600 hover:text-green-800 text-xs font-medium"
                >Activate</button>
                <button
                  v-if="member.id !== currentUserId"
                  @click="deleteUser(member.id, member.name)"
                  class="text-red-600 hover:text-red-800 text-xs font-medium"
                >Delete</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Assign to Project Modal -->
    <div v-if="assignTarget" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-sm mx-4">
        <h2 class="text-base font-semibold mb-1">Assign to Projects</h2>
        <p class="text-sm text-gray-500 mb-4">
          <span class="font-medium text-gray-700">{{ assignTarget.name }}</span>
          — select one or more projects
        </p>
        <div class="space-y-2 mb-4">
          <label
            v-for="proj in availableProjectsFor(assignTarget.id)"
            :key="proj.id"
            class="flex items-center gap-3 p-2.5 rounded-lg border cursor-pointer hover:bg-indigo-50 transition-colors"
            :class="assignProjectIds.includes(proj.id) ? 'border-indigo-400 bg-indigo-50' : 'border-gray-200'"
          >
            <input
              type="checkbox"
              :value="proj.id"
              v-model="assignProjectIds"
              class="rounded text-indigo-600 focus:ring-indigo-500"
            />
            <div>
              <p class="text-sm font-medium text-gray-800">{{ proj.name }}</p>
              <p class="text-xs text-gray-500">{{ proj.description }}</p>
            </div>
          </label>
          <p v-if="availableProjectsFor(assignTarget.id).length === 0" class="text-sm text-gray-400 text-center py-3">
            Already assigned to all projects.
          </p>
        </div>
        <div v-if="assignError" class="text-red-600 text-sm bg-red-50 px-3 py-2 rounded mb-3">{{ assignError }}</div>
        <div class="flex gap-3">
          <button
            @click="doAssign"
            :disabled="assignProjectIds.length === 0 || assigning"
            class="flex-1 bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
          >{{ assigning ? 'Assigning…' : `Assign to ${assignProjectIds.length || ''} project${assignProjectIds.length === 1 ? '' : 's'}` }}</button>
          <button @click="closeAssign" class="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg text-sm">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Create Staff Modal -->
    <div v-if="showCreateForm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md mx-4">
        <h2 class="text-lg font-semibold mb-1">New Staff Account</h2>
        <p class="text-sm text-gray-500 mb-1">Staff will receive a welcome email with these credentials.</p>
        <p class="text-xs text-indigo-600 bg-indigo-50 px-3 py-2 rounded-lg mb-4">They will be required to change their password on first login.</p>
        <form @submit.prevent="createStaff" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Full Name *</label>
            <input
              v-model="newStaff.name"
              type="text"
              required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email *</label>
            <input
              v-model="newStaff.email"
              type="email"
              required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Password *</label>
            <input
              v-model="newStaff.password"
              type="password"
              required
              minlength="8"
              class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
              placeholder="Min. 8 characters"
            />
          </div>
          <div v-if="createError" class="text-red-600 text-sm bg-red-50 px-3 py-2 rounded">{{ createError }}</div>
          <div class="flex gap-3 pt-1">
            <button
              type="submit"
              :disabled="createLoading"
              class="flex-1 bg-indigo-600 text-white py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
            >{{ createLoading ? 'Creating…' : 'Create Account' }}</button>
            <button
              type="button"
              @click="showCreateForm = false"
              class="flex-1 bg-gray-100 text-gray-700 py-2 rounded-lg text-sm"
            >Cancel</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useStaffStore } from '../../stores/staff';
import { useAuthStore } from '../../stores/auth';
import type { StaffUser, Project, ProjectMember } from '../../stores/staff';

const staffStore = useStaffStore();
const auth = useAuthStore();

const loading = ref(true);
const allStaff = ref<StaffUser[]>([]);
const allProjects = ref<Project[]>([]);

// memberships[projectId] = ProjectMember[]
const memberships = ref<Record<number, ProjectMember[]>>({});

const currentUserId = computed(() => auth.user?.id);

// Assign modal
const assignTarget = ref<StaffUser | null>(null);
const assignProjectIds = ref<number[]>([]);
const assigning = ref(false);
const assignError = ref('');

// Create modal
const showCreateForm = ref(false);
const createLoading = ref(false);
const createError = ref('');
const newStaff = ref({ name: '', email: '', password: '' });

onMounted(async () => {
  await auth.fetchUser();
  await load();
});

async function load() {
  loading.value = true;
  try {
    const [staff, projects] = await Promise.all([
      staffStore.fetchStaff(),
      staffStore.fetchProjects(),
    ]);
    allStaff.value = staffStore.staff;
    allProjects.value = projects;

    // Load memberships for all projects in parallel
    const results = await Promise.all(projects.map((p) => staffStore.fetchProjectStaff(p.id)));
    const map: Record<number, ProjectMember[]> = {};
    projects.forEach((p, i) => { map[p.id] = results[i]; });
    memberships.value = map;
  } finally {
    loading.value = false;
  }
}

// Returns { id, projectStaffId, name } for each project this user is in
function memberProjects(userId: number) {
  const result: { id: number; projectStaffId: number; name: string }[] = [];
  for (const proj of allProjects.value) {
    const ms = memberships.value[proj.id] || [];
    const m = ms.find((x) => x.user_id === userId);
    if (m) result.push({ id: proj.id, projectStaffId: m.id, name: proj.name });
  }
  return result;
}

// Projects this user is NOT yet in
function availableProjectsFor(userId: number) {
  const assigned = new Set(memberProjects(userId).map((p) => p.id));
  return allProjects.value.filter((p) => !assigned.has(p.id));
}

function openAssign(user: StaffUser) {
  assignTarget.value = user;
  assignProjectIds.value = [];
  assignError.value = '';
}

function closeAssign() {
  assignTarget.value = null;
  assignProjectIds.value = [];
  assignError.value = '';
}

async function doAssign() {
  if (!assignTarget.value || assignProjectIds.value.length === 0) return;
  assigning.value = true;
  assignError.value = '';
  try {
    await Promise.all(
      assignProjectIds.value.map(async (projectId) => {
        const member = await staffStore.addToProject(projectId, assignTarget.value!.id);
        if (!memberships.value[projectId]) memberships.value[projectId] = [];
        memberships.value[projectId].push(member);
      })
    );
    closeAssign();
  } catch (e: any) {
    assignError.value = e.response?.data?.message || 'Failed to assign.';
  } finally {
    assigning.value = false;
  }
}

async function removeProject(projectStaffId: number, userName: string, projectName: string) {
  if (!confirm(`Remove ${userName} from ${projectName}?`)) return;
  // Find which project this belongs to
  for (const [projId, ms] of Object.entries(memberships.value)) {
    const m = ms.find((x) => x.id === projectStaffId);
    if (m) {
      await staffStore.removeFromProject(Number(projId), m.user_id);
      memberships.value[Number(projId)] = ms.filter((x) => x.id !== projectStaffId);
      break;
    }
  }
}

async function deactivate(id: number) {
  if (!confirm('Deactivate this staff member?')) return;
  await staffStore.deactivateStaff(id);
  const idx = allStaff.value.findIndex((s) => s.id === id);
  if (idx !== -1) allStaff.value[idx].is_active = false;
}

async function activate(id: number) {
  await staffStore.activateStaff(id);
  const idx = allStaff.value.findIndex((s) => s.id === id);
  if (idx !== -1) allStaff.value[idx].is_active = true;
}

async function deleteUser(id: number, name: string) {
  if (!confirm(`Delete ${name}? This cannot be undone.`)) return;
  await staffStore.deleteStaff(id);
  allStaff.value = allStaff.value.filter((s) => s.id !== id);
}

async function createStaff() {
  createLoading.value = true;
  createError.value = '';
  try {
    const created = await staffStore.createStaff(newStaff.value);
    allStaff.value.push(created);
    showCreateForm.value = false;
    newStaff.value = { name: '', email: '', password: '' };
  } catch (e: any) {
    createError.value = e.response?.data?.message || 'Failed to create account.';
  } finally {
    createLoading.value = false;
  }
}
</script>
