<template>
  <div class="page-component">
    <div class="page-component__inner">
      <div class="page-component__header">
        <DashboardHeader @handle-page-data-refresh="emits('handlePageDataRefresh')">
          <template #header-left>
            <slot name="header-slot" />
          </template>
        </DashboardHeader>
      </div>

      <div :class="['page-component__content', {fullWidth: isPageFullWidth, padding: !isMain}]">
        <slot name="content-slot"/>
      </div>
    </div>
  </div>
</template>

<script setup>
import DashboardHeader from "../Base/DashboardHeader.vue";
import {computed} from "vue";
import {useRoute} from "vue-router";

defineProps({
  isMain: {type: Boolean, default: false},
})
const emits = defineEmits(["handlePageDataRefresh"]);
const route = useRoute();

const isPageFullWidth = computed(() => {
  const pages = ['Dashboard', 'MarketSmartBuyback', 'TokenCreate', 'TokenVolumeMaker', 'TokenHistory', 'DashboardNotFound'];

  return pages.includes(route.name);
})
</script>

<style scoped lang="scss">
.page-component {
  height: 100%;

  &__inner {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  &__header {
    display: flex;
  }

  &__content {
    position: relative;
    z-index: 10;
    flex-grow: 1;
    display: flex;
    overflow-y: auto;
    max-width: 1211px;
    flex-direction: column;

    &.fullWidth {
      max-width: none;
    }

    &.padding {
      padding: 24px;
    }
  }
}

@media (max-width: 1200px) {
  .page-component {
    &__content {
      padding: 0;
    }

    &__header {
      display: none;
    }
  }
}
</style>
