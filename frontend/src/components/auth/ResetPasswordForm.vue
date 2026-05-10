<template>
  <div class="bg-white rounded-xl shadow-md p-8">
    <h2 class="text-2xl font-bold text-gray-800 mb-2">Reset Password</h2>
    <p class="text-gray-500 text-sm mb-6">Enter your new password.</p>

    <div v-if="success" class="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
      <p class="text-green-700 text-sm">{{ success }}</p>
      <a href="/login" class="mt-2 inline-block text-blue-600 hover:underline text-sm">Go to Login</a>
    </div>

    <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700 text-sm">{{ errorMessage }}</p>
    </div>

    <form v-if="!success" @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
        <input
          v-model="form.password"
          type="password"
          required
          minlength="8"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Confirm Password</label>
        <input
          v-model="form.password_confirmation"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 font-medium"
      >
        {{ loading ? 'Resetting...' : 'Reset Password' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import api from '../../utils/api';

const form = reactive({ password: '', password_confirmation: '' });
const loading = ref(false);
const success = ref('');
const errorMessage = ref('');
const token = ref('');
const email = ref('');

onMounted(() => {
  const params = new URLSearchParams(window.location.search);
  token.value = params.get('token') || '';
  email.value = params.get('email') || '';
});

async function handleSubmit() {
  loading.value = true;
  success.value = '';
  errorMessage.value = '';
  try {
    const response = await api.post('/api/reset-password', {
      token: token.value,
      email: email.value,
      password: form.password,
      password_confirmation: form.password_confirmation,
    });
    success.value = response.data.message;
  } catch (e: any) {
    errorMessage.value = e.response?.data?.message || 'Failed to reset password.';
  } finally {
    loading.value = false;
  }
}
</script>
