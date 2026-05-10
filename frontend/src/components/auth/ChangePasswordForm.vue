<template>
  <div class="bg-white rounded-xl shadow-md p-8">
    <div class="mb-6">
      <h2 class="text-2xl font-bold text-gray-800 mb-1">Set your password</h2>
      <p class="text-sm text-gray-500">Your account requires a new password before you can continue.</p>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700 text-sm">{{ error }}</p>
    </div>
    <div v-if="success" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg">
      <p class="text-green-700 text-sm">Password changed. Redirecting…</p>
    </div>

    <form @submit.prevent="submit" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Current (temporary) password</label>
        <input
          v-model="form.current"
          type="password"
          required
          autofocus
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="••••••••"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">New password</label>
        <input
          v-model="form.newPwd"
          type="password"
          required
          minlength="8"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="Min. 8 characters"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Confirm new password</label>
        <input
          v-model="form.confirm"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="••••••••"
        />
      </div>
      <button
        type="submit"
        :disabled="loading"
        class="w-full bg-indigo-600 text-white py-2.5 rounded-lg hover:bg-indigo-700 disabled:opacity-50 font-medium transition-colors"
      >
        {{ loading ? 'Saving…' : 'Set password & continue' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import api from '../../utils/api';

const form    = reactive({ current: '', newPwd: '', confirm: '' });
const loading = ref(false);
const error   = ref('');
const success = ref(false);

async function submit() {
  error.value = '';
  if (form.newPwd !== form.confirm) {
    error.value = 'New passwords do not match.';
    return;
  }
  loading.value = true;
  try {
    await api.post('/api/auth/change-password', {
      current_password: form.current,
      new_password:     form.newPwd,
    });
    success.value = true;
    setTimeout(() => { window.location.href = '/dashboard'; }, 1000);
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Failed to change password.';
  } finally {
    loading.value = false;
  }
}
</script>
