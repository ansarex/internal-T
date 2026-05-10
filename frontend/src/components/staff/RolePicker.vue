<template>
  <div class="flex flex-wrap gap-2">
    <button
      v-for="role in availableRoles"
      :key="role.value"
      type="button"
      @click="toggleRole(role.value)"
      :class="[
        'px-3 py-1 rounded-full text-xs font-semibold transition-all border',
        modelValue.includes(role.value)
          ? role.activeClass
          : 'bg-white text-gray-400 border-gray-200 hover:border-gray-400',
      ]"
    >
      {{ role.label }}
    </button>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ modelValue: string[]; disableAdmin?: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>();

const availableRoles = [
  { value: 'admin', label: 'Admin', activeClass: 'bg-blue-600 text-white border-blue-600' },
  { value: 'support', label: 'Support', activeClass: 'bg-sky-500 text-white border-sky-500' },
  { value: 'sales', label: 'Sales', activeClass: 'bg-green-600 text-white border-green-600' },
  { value: 'cs', label: 'CS', activeClass: 'bg-orange-500 text-white border-orange-500' },
];

function toggleRole(role: string) {
  if (props.disableAdmin && role === 'admin') return;
  const current = [...props.modelValue];
  const idx = current.indexOf(role);
  if (idx === -1) {
    current.push(role);
  } else {
    if (current.length === 1) return; // must have at least one role
    current.splice(idx, 1);
  }
  emit('update:modelValue', current);
}
</script>
