<template>
<div :class="['ui-badge paragraph-mini', type, color_type]">
  <slot/>
</div>
</template>
<script setup>
defineProps({
  type: {type: String, default: 'default'}, //default (rectangle) | round
  color_type: {type: String, default: 'outline'}, //primary | secondary | outline | ghost | destructive | success
})
</script>
<style scoped lang="scss">
.ui-badge {
  font-weight: 500 !important;
  padding: 0 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 24px;

  &.round {
    border-radius: 9999px;
  }

  &.default {
    border-radius: 8px;
  }

  &.primary {
    background: #111827;
  }

  &.secondary {
    background: #F3F4F6;
  }

  &.outline {
    background: rgba(255, 255, 255, 0.10);
    border: 1px solid #E5E7EB;
  }

  &.ghost {
    background: transparent;
  }

  &.destructive {
    background: #DC2626;
  }

  &.success {
    background: #16A34A;
    color: #FFF;
  }
}
</style>