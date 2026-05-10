<template>
  <div class="bg-white rounded-xl shadow-md p-8">
    <h2 class="text-2xl font-bold text-gray-800 mb-2">Forgot Password</h2>
    <p class="text-gray-500 text-sm mb-6">Enter your email and we'll send you a reset link.</p>

    <div v-if="success" class="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
      <p class="text-green-700 text-sm">{{ success }}</p>
    </div>

    <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700 text-sm">{{ errorMessage }}</p>
    </div>

    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <input
          v-model="email"
          type="email"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="admin@example.com"
        />
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 font-medium"
      >
        {{ loading ? 'Sending...' : 'Send Reset Link' }}
      </button>
    </form>

    <div class="mt-4 text-center">
      <a href="/login" class="text-sm text-blue-600 hover:underline">Back to login</a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import api from '../../utils/api';

const email = ref('');
const loading = ref(false);
const success = ref('');
const errorMessage = ref('');

async function handleSubmit() {
  loading.value = true;
  success.value = '';
  errorMessage.value = '';
  try {
    const response = await api.post('/api/forgot-password', { email: email.value });
    success.value = response.data.message;
  } catch (e: any) {
    errorMessage.value = e.response?.data?.message || 'An error occurred.';
  } finally {
    loading.value = false;
  }
}
</script>
