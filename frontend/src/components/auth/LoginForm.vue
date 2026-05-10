<template>
  <div class="bg-white rounded-xl shadow-md p-8">
    <h2 class="text-2xl font-bold text-gray-800 mb-2">Sign In</h2>
    <p class="text-sm text-gray-500 mb-6">Trustwired Internal System</p>

    <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700 text-sm">{{ errorMessage }}</p>
    </div>

    <form @submit.prevent="handleLogin" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input
          v-model="form.email"
          type="email"
          required
          autofocus
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="you@example.com"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <input
          v-model="form.password"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="••••••••"
        />
      </div>
      <div class="flex justify-end">
        <a href="/forgot-password" class="text-sm text-indigo-600 hover:underline">Forgot password?</a>
      </div>
      <button
        type="submit"
        :disabled="loading"
        class="w-full bg-indigo-600 text-white py-2.5 rounded-lg hover:bg-indigo-700 disabled:opacity-50 font-medium transition-colors"
      >
        {{ loading ? 'Signing in…' : 'Sign In' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useAuthStore } from '../../stores/auth';

const auth = useAuthStore();
const form = reactive({ email: '', password: '' });
const loading = ref(false);
const errorMessage = ref('');

async function handleLogin() {
  loading.value = true;
  errorMessage.value = '';

  const result = await auth.login(form.email, form.password);

  if (result.must_change_password) {
    window.location.href = '/change-password';
    return;
  }

  if (!result.email_verified) {
    errorMessage.value = result.message || 'Email not verified.';
    loading.value = false;
    return;
  }

  if (auth.error) {
    errorMessage.value = auth.error;
    loading.value = false;
    return;
  }

  window.location.href = '/dashboard';
}
</script>
