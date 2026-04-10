<template>
  <div class="loader-wrapper heading-5 grey">
    <UISpinner color="blue"/>
    <span>Page is loading...</span>
  </div>
</template>
<script setup>
import UISpinner from "./UISpinner.vue";
</script>
<style scoped lang="scss">
.loader-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  position: absolute;
  top: calc(50% - 63px);
  left: 50%;
  transform: translate(-50%, -50%);
}
</style>