<template>
  <div class="table" :class="{ 'table--nested-inner': nestedLevel > 0, 'table--nested': isTableNested }">

    <!-- HEADER -->
    <div v-if="nestedLevel === 0" class="table__header">
      <div
        v-for="(col, index) in columns"
        :class="['table__header_col', col.field]"
        :key="index"
      >
        <slot
          v-if="$slots[`col_${col.field}`]"
          :name="`col_${col.field}`"
          :item="col"
        />

        <div v-else :class="`paragraph-small table__header_label --${index+1}`" @click="emits('handleSort', col.field)">
          {{ col.label || '' }}
        </div>
      </div>
    </div>

    <!-- BODY -->
    <div class="table__body">
      <template v-for="(element, index) in rows" :key="index">

        <!-- ROW -->
        <div
          :class="['table__row', { 'table__row--parent': hasChildren(element) }]"
          @click="handleRowClick(element, index)"
        >
          <div
            v-for="(item, i) in columns"
            :class="['paragraph-medium table__row_cell', item.field]"
            :key="i"
          >
            <slot
              v-if="resolveSlot(item.field)"
              :name="resolveSlot(item.field)"
              :item="element"
              :row-index="index"
              :is-expanded="expandedRows[index]"
              :nested-level="nestedLevel"
              :toggle-expand="() => toggleExpand(index)"
            />

            <span v-else class="color-primary">
              {{ getColValue(element[item.field]) }}
            </span>
          </div>
        </div>

        <!-- NESTED -->
        <Transition name="nested-expand">
          <div
            v-if="expandedRows[index] && hasChildren(element)"
            class="table__nested"
          >
            <UITable
              :rows="element[nestedColumnsName]"
              :columns="nestedColumns || columns"
              :nested-level="nestedLevel + 1"
            >
              <!-- Forward all slots -->
              <template
                v-for="(_, slotName) in $slots"
                #[slotName]="slotProps"
              >
                <slot :name="slotName" v-bind="slotProps" />
              </template>
            </UITable>
          </div>

          <div v-else-if="expandedRows[index] && isEmptyState">
            <slot name="table-nested-empty" :item="element"/>
          </div>
        </Transition>

      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, useSlots } from 'vue'

const props = defineProps({
  rows: { type: Array, default: () => [] },
  columns: { type: Array, default: () => [] },
  nestedColumns: { type: Array, default: null },
  nestedLevel: { type: Number, default: 0 },
  nestedColumnsName: { type: String, default: '' },
  isEmptyState: { type: Boolean, default: false },
  isTableNested: { type: Boolean, default: false },
})
const emits = defineEmits(['handleRowClick', 'handleSort'])

const slots = useSlots()
const expandedRows = ref({})

const hasChildren = (el) =>
  Array.isArray(el[props.nestedColumnsName]) && el[props.nestedColumnsName].length > 0

const toggleExpand = (index) => {
  expandedRows.value[index] = !expandedRows.value[index]
}

const getColValue = (value) =>
  value === null || value === undefined ? '' : value

const resolveSlot = (field) => {
  const nestedSlot = `${field}_nested`

  if (props.nestedLevel > 0 && slots[nestedSlot]) {
    return nestedSlot
  }

  if (slots[field]) {
    return field
  }

  return null
}

const handleRowClick = (element, index) => {
  (hasChildren(element) || props.isEmptyState) && props.nestedLevel === 0 && toggleExpand(index);

  emits('handleRowClick', element);
}
</script>

<style scoped lang="scss">
.table {
  display: flex;
  flex-direction: column;

  &__header {
    border-bottom: 1px solid #E6E7EB;
    background: #FFF;
    display: flex;
    align-items: center;
    height: 36px;

    &_col {
      width: 100%;
      padding-inline: 16px;
    }

    &_label {
      font-weight: 500;
    }
  }

  &__body {
    background: transparent;
  }

  &__row {
    display: flex;
    align-items: center;
    height: 48px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.10);
    transition: background-color 0.2s;
    background: #E5E7EB;

    &:last-child {
      border: none;
    }

    &--parent {
      cursor: pointer;
    }

    &_cell {
      color: #302F2F;
      text-overflow: ellipsis;
      font-weight: 400;
      width: 100%;
      padding-inline: 16px;
    }
  }

  &--nested {
    &-inner {
      .table__header {
        display: none;
      }

      .table__body {
        background: transparent;
      }

      .table__row {
        background: transparent;
        border-bottom: none;
      }
    }


    .table__body {
      background: transparent;
    }

    .table__row {
      background: transparent;
      border-bottom: none;
    }
  }

  &__nested {
    background-color: transparent;

    max-height: calc(48px * 6);
    overflow-y: auto;

    .table {
      border-radius: 0;
      margin: 0;
    }
  }
}

.nested-expand-enter-active,
.nested-expand-leave-active {
  transition:
    max-height 0.3s ease,
    opacity 0.25s ease,
    padding 0.25s ease;
  overflow: hidden;
}

.nested-expand-enter-from,
.nested-expand-leave-to {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.nested-expand-enter-to,
.nested-expand-leave-from {
  max-height: calc(48px * 6);
  opacity: 1;
}
</style>