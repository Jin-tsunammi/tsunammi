<template>
  <div class="manage__content">
    <EmptyPageNotification />
  </div>
</template>
<script setup>
import EmptyPageNotification from "../../components/UI/Modals/EmptyPageNotification.vue";
</script>
<style scoped lang="scss">
.manage {
  &__content {
    height: 100%;
    padding-top: 64px;
    width: 100%;
    display: flex;
    justify-content: center;
    position: relative;
  }
}

@media (max-width: 1200px) {
  .manage {
    &__content {
      padding: 16px;
    }
  }
}
</style>