<template>
  <button :class="['ui-tab paragraph-small', {active: is_active, disabled: is_disabled}, size]">
    <slot/>
    <slot name="add-info"/>
  </button>
</template>
<script setup>
defineProps({
  size: {type: String, default: 'regular'}, //regular | large | small
  is_disabled: {type: Boolean, default: false},
  is_active: {type: Boolean, default: false},
})
</script>
<style scoped lang="scss">
.ui-tab {
  height: 100%;
  width: 100%;
  background: transparent;
  color: #030712;
  font-weight: 500;
  border-radius: 10px;
  transition: .3s ease;
  display: flex;
  flex-direction: column;
  justify-content: center;

  &:hover {
    background: #F3F4F6;
  }

  &:focus {
    box-shadow: 0 0 0 3px #D1D5DB;
    background: #F3F4F6;
  }

  &.active {
    background: #FFF;

    &:focus {
      box-shadow: 0 0 0 3px #D1D5DB;
    }
  }
}
</style>