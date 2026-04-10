<template>
  <div class="section-title">
    <div class="section-title__flex">
      <div class="heading-3"><slot /></div>
      <slot name="actions" />
    </div>
    <div class="divider"></div>
  </div>
</template>
<script setup>

</script>
<style scoped lang="scss">
.section-title {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;

  &__flex {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  & .divider {
    height: 1px;
    width: 100%;
    background: #D1D5DB;
  }
}
</style>