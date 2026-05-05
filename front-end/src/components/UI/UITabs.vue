<template>
  <div :class="['ui-tabs', size, {disabled: is_disabled}]">
    <slot />
  </div>
</template>
<script setup>
defineProps({
  size: {type: String, default: 'regular'}, //regular | large | small
  is_disabled: {type: Boolean, default: false},
})
</script>
<style scoped lang="scss">
.ui-tabs {
  display: flex;
  align-items: center;
  background: #E5E7EB;
  gap: 3px;

  &.disabled {
    opacity: .5;

    &:hover {
      cursor: not-allowed;
    }
  }

  &.regular {
    height: 35px;
    border-radius: 10px;
    padding: 3px;
  }

  &.large {
    height: 40px;
    border-radius: 12px;
    padding: 4px;
  }

  &.small {
    height: 24px;
    border-radius: 10px;
    padding: 2px;
  }
}
</style>