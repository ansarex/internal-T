<template>
  <div class="bg-white rounded-xl shadow p-5 flex items-start gap-4">
    <div :class="['w-12 h-12 rounded-lg flex items-center justify-center flex-shrink-0', iconBg]">
      <slot name="icon">
        <div class="w-6 h-6 bg-white bg-opacity-50 rounded"></div>
      </slot>
    </div>
    <div>
      <p class="text-sm text-gray-500">{{ label }}</p>
      <p :class="['text-3xl font-bold mt-0.5', valueColor]">{{ value }}</p>
      <p v-if="subtitle" class="text-xs text-gray-400 mt-0.5">{{ subtitle }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string;
  value: number | string;
  subtitle?: string;
  iconBg?: string;
  valueColor?: string;
}>();
</script>
