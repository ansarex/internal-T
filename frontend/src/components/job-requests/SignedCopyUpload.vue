<template>
  <div class="bg-blue-50 border border-blue-200 rounded-xl p-5">
    <h3 class="font-semibold text-blue-900 mb-2">Upload Signed Copy</h3>
    <p class="text-sm text-blue-700 mb-4">
      Upload the final copy signed by the customer and your side to unlock CS Tasks.
    </p>
    <div class="flex items-center gap-3">
      <input
        ref="fileInput"
        type="file"
        accept=".pdf"
        class="text-sm text-gray-600 flex-1"
      />
      <button
        @click="upload"
        :disabled="uploading"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium whitespace-nowrap"
      >
        {{ uploading ? 'Uploading...' : 'Upload & Unlock CS Tasks' }}
      </button>
    </div>
    <div v-if="error" class="text-red-600 text-sm mt-2">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useJobRequestsStore } from '../../stores/jobRequests';

const props = defineProps<{ jobId: number }>();
const emit = defineEmits<{ uploaded: [] }>();

const jobStore = useJobRequestsStore();
const fileInput = ref<HTMLInputElement | null>(null);
const uploading = ref(false);
const error = ref('');

async function upload() {
  const file = fileInput.value?.files?.[0];
  if (!file) { error.value = 'Please select a PDF file.'; return; }
  uploading.value = true;
  error.value = '';
  try {
    await jobStore.uploadSignedCopy(props.jobId, file);
    emit('uploaded');
  } catch (e: any) {
    error.value = e.response?.data?.message || 'Upload failed.';
  } finally {
    uploading.value = false;
  }
}
</script>
