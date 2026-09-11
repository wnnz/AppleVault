<template>
  <input
    ref="inputRef"
    :value="modelValue"
    :type="type"
    :disabled="disabled"
    :readonly="readonly"
    class="app-input clean-input"
    @input="onInput"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'

withDefaults(defineProps<{
  modelValue?: string | number | null
  type?: string
  disabled?: boolean
  readonly?: boolean
}>(), {
  modelValue: '',
  type: 'text',
  disabled: false,
  readonly: false
})

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const inputRef = ref<HTMLInputElement | null>(null)

function onInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}

defineExpose({
  focus: () => inputRef.value?.focus(),
  select: () => inputRef.value?.select(),
  element: inputRef
})
</script>
