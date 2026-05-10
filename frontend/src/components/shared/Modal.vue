<template>
  <Teleport to="body">
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-center justify-center"
    >
      <!-- Backdrop -->
      <div
        class="absolute inset-0 bg-black bg-opacity-50"
        @click="$emit('update:modelValue', false)"
      ></div>

      <!-- Modal -->
      <div
        :class="['relative bg-white rounded-xl shadow-xl w-full', maxWidthClass]"
        @click.stop
      >
        <!-- Header -->
        <div class="flex items-center justify-between p-5 border-b border-gray-100">
          <h3 class="text-lg font-semibold text-gray-800">{{ title }}</h3>
          <button
            @click="$emit('update:modelValue', false)"
            class="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Body -->
        <div class="p-5">
          <slot />
        </div>

        <!-- Footer (optional) -->
        <div v-if="$slots.footer" class="px-5 pb-5">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  modelValue: boolean;
  title?: string;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl';
}>();

defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
}>();

const maxWidthClass = computed(() => {
  const map = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
    '2xl': 'max-w-2xl',
  };
  return map[props.maxWidth || 'md'];
});
</script>
