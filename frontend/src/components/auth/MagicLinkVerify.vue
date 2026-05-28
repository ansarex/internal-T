<template>
  <div class="bg-white rounded-xl shadow-md p-8 text-center">

    <!-- Verifying -->
    <div v-if="state === 'verifying'">
      <div class="w-14 h-14 bg-indigo-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-7 h-7 text-indigo-600 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
        </svg>
      </div>
      <h2 class="text-xl font-bold text-gray-800 mb-2">Signing you in…</h2>
      <p class="text-gray-500 text-sm">Verifying your sign-in link.</p>
    </div>

    <!-- Success -->
    <div v-else-if="state === 'success'">
      <div class="w-14 h-14 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-7 h-7 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"/>
        </svg>
      </div>
      <h2 class="text-xl font-bold text-gray-800 mb-2">You're signed in!</h2>
      <p class="text-gray-500 text-sm">Redirecting you now…</p>
    </div>

    <!-- Error -->
    <div v-else>
      <div class="w-14 h-14 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-7 h-7 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/>
        </svg>
      </div>
      <h2 class="text-xl font-bold text-gray-800 mb-2">Sign-in failed</h2>
      <p class="text-red-600 text-sm font-medium mb-6">{{ errorMessage }}</p>
      <a
        href="/magic-link"
        class="inline-block bg-indigo-600 text-white px-6 py-2.5 rounded-lg hover:bg-indigo-700 text-sm font-medium"
      >Request a new link</a>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '../../stores/auth';
import api from '../../utils/api';

const auth = useAuthStore();
const state = ref<'verifying' | 'success' | 'error'>('verifying');
const errorMessage = ref('');

onMounted(async () => {
  try {
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');
    const email = params.get('email');

    if (!token || !email) {
      throw new Error('Invalid sign-in link.');
    }

    const res = await api.post('/api/magic-link/verify', { token, email });
    const { token: sessionToken, user } = res.data;

    localStorage.setItem('auth_token', sessionToken);
    auth.user = user;
    auth.token = sessionToken;

    state.value = 'success';
    setTimeout(() => { window.location.href = '/dashboard'; }, 800);
  } catch (e: any) {
    state.value = 'error';
    errorMessage.value = e.response?.data?.message ?? e.message ?? 'This sign-in link is invalid or has expired.';
  }
});
</script>
