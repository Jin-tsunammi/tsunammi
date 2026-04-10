<template>
  <div class="mobile-table">
    <div
      class="mobile-table__item"
      v-for="(parent, index) in rows"
      :key="index"
    >
      <div class="mobile-table__item_top">
        <slot name="parent-top" :item="parent" :index="index" />
      </div>

      <div 
        v-if="hasChildren(parent)"
        class="mobile-table__nested"
      >
        <div 
          class="mobile-table__nested_top"
          @click="toggleExpand(index)"
        >
          <slot name="nested-label" :item="parent" :index="index" />
          <div
            class="mobile-table__nested_toggle"
            :class="{ 'mobile-table__nested_toggle--expanded': expandedRows[index] }"
          >
            <SVGSmallArrowDown color="#030712"/>
          </div>
        </div>

        <Transition name="nested-expand">
          <div
            v-if="expandedRows[index]"
            class="mobile-table__nested_list"
          >
            <div
              v-for="(child, childIndex) in parent.children"
              :key="childIndex"
              class="mobile-table__nested_item"
            >
              <slot 
                name="nested-item" 
                :item="child" 
                :parent="parent"
                :index="childIndex"
                :parent-index="index"
              />
            </div>
          </div>
        </Transition>
      </div>

      <div 
        v-if="hasActionsSlot"
        class="mobile-table__item_actions"
      >
        <slot name="parent-actions" :item="parent" :index="index" />
      </div>
    </div>
  </div>
</template>
<script setup>
import {ref, useSlots, computed} from "vue";
import SVGSmallArrowDown from "../SVG/SVGSmallArrowDown.vue";

defineProps({
  rows: {type: Array, default: () => []},
})

const slots = useSlots()
const expandedRows = ref({})

const hasChildren = (item) => {
  return Array.isArray(item?.children) && item.children.length > 0
}

const toggleExpand = (index) => {
  expandedRows.value[index] = !expandedRows.value[index]
}

const hasActionsSlot = computed(() => {
  return !!slots['parent-actions']
})
</script>
<style scoped lang="scss">
.mobile-table {
  display: flex;
  flex-direction: column;
  gap: 12px;

  &__item {
    display: flex;
    flex-direction: column;
    border-radius: 8px;
    border: 1px solid #D1D5DB;

    &_actions {
      padding: 16px;
      width: 100%;
      background: #E5E7EB;
    }

    &_top {
      padding: 16px;
      width: 100%;
      background: #E5E7EB;
    }
  }

  &__nested {
    width: 100%;
    padding: 12px 16px;

    &_toggle {
      display: flex;
      align-items: center;
      justify-content: center;
      background: transparent;
      border: none;
      cursor: pointer;
      padding: 4px;
      transition: transform 0.3s ease;
      width: 36px;
      height: 36px;

      &--expanded {
        transform: rotate(180deg);
      }
    }

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      cursor: pointer;
      user-select: none;
    }

    &_list {
      width: 100%;
      margin-top: 12px;
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    &_item {
      width: 100%;
      padding: 16px;
      border-radius: 8px;
      border: 1px solid #E5E7EB;
    }
  }
}

.nested-expand-enter-active,
.nested-expand-leave-active {
  transition:
    max-height 0.3s ease,
    opacity 0.25s ease,
    margin-top 0.25s ease;
  overflow: hidden;
}

.nested-expand-enter-from,
.nested-expand-leave-to {
  max-height: 0;
  opacity: 0;
  margin-top: 0;
}

.nested-expand-enter-to,
.nested-expand-leave-from {
  max-height: 1000px;
  opacity: 1;
  margin-top: 12px;
}
</style>