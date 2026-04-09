<template>
  <div :class="['ui-tabs', size]">
    <slot />
  </div>
</template>
<script setup>
defineProps({
  size: {type: String, default: 'regular'}, //regular | large | small
})
</script>
<style scoped lang="scss">
.ui-tabs {
  display: flex;
  align-items: center;
  background: #E5E7EB;
  gap: 3px;

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