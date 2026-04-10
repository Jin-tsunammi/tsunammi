<template>
  <div class="ui-empty-state" :style="{background}">
    <div class="ui-empty-state__icon">
      <component v-if="icon" :is="icon" />
    </div>
    <div class="ui-empty-state__main-text paragraph-medium">{{main_text}}</div>
    <span class="ui-empty-state__add-text paragraph-small gray regular">{{add_text}}</span>
    <div class="ui-empty-state__actions">
      <slot name="actions" />
    </div>
  </div>
</template>
<script setup>
defineProps({
  icon: {type: Object, default: null},
  main_text: {type: String, default: ''},
  add_text: {type: String, default: ''},
  background: {type: String, default: 'transparent'},
})
</script>
<style scoped lang="scss">
.ui-empty-state {
  height: 217px;
  border-radius: 8px;
  border: 1px solid #D1D5DB;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;

  &__main-text {
    margin: 12px 0 4px;
  }

  &__add-text {
    display: block;
    margin-bottom: 24px;
  }
}
</style>