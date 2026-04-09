<template>
  <div class="manage__content">
    <EmptyPageNotification />
    <Modals />
  </div>
</template>
<script setup>
import EmptyPageNotification from "../../components/UI/Modals/EmptyPageNotification.vue";
import Modals from "../../components/UI/Modals.vue";
</script>
<style scoped lang="scss">
.manage {
  &__content {
    padding-top: 64px;
    width: 100%;
    display: flex;
    justify-content: center;
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