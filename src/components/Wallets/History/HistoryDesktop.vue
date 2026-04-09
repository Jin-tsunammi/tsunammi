<template>
  <div class="history-desktop">
    <UISectionTitleWithBorder>
      <template #default>
        History
      </template>
      <template #actions>
        <div class="history-desktop__range">
          <UIMultiCalendars @handle-dates-change="emits('handleDatesChange', $event)"/>
        </div>
      </template>
    </UISectionTitleWithBorder>

    <div class="history-desktop__empty" v-if="!rows.length">
      <SVGClockFading />
      <div class="selected-cex-desktop__empty_title paragraph-medium bold">No history yet</div>
      <span class="grey regular paragraph-small">Actions and system events will appear here once they occur.</span>
    </div>

    <div v-else class="history-desktop__table">
      <UITable :columns="columns" :rows="rows">
        <template #created_at="{ item }">
          <div class="paragraph-small regular">{{formatDate(item.created_at).date}} <span class="grey">{{ formatDate(item.created_at).time }}</span></div>
        </template>у
        <template #value="{ item }">
          <span :class="{ 'paragraph-small regular': item }">{{ formatText(item.value) }}</span>
        </template>
        <template #action="{item}">
          <div class="table__actions">
            <span class="paragraph-mini medium">{{formatText(item.action)}}</span>
          </div>
        </template>
      </UITable>
    </div>
  </div>
</template>
<script setup>
import UITable from "../../UI/UITable.vue";
import {formatDate, formatText} from "../../../helpers/index.js";
import UISectionTitleWithBorder from "../../UI/UISectionTitleWithBorder.vue";
import UIMultiCalendars from "../../UI/UIMultiCalendars.vue";
import SVGClockFading from "../../SVG/SVGClockFading.vue";

defineProps({
  columns: {type: Array, default: []},
  rows: {type: Array, default: []},
})
const emits = defineEmits(['handleDatesChange'])
</script>
<style scoped lang="scss">
.history-desktop {
  &__range {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__empty {
    margin-top: 20px;
    border-radius: 12px;
    background: #FFF;
    height: 309px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    & div {
      margin-top: 12px;
      margin-bottom: 4px;
    }

    & span {
      max-width: 317px;
      text-align: center;
    }
  }

  &__table {
    margin-top: 20px;
    width: 100%;
    max-width: 1400px;

    ::v-deep(.table__header_col) {
      &.created_at, &.action {
        width: calc((210 / 1163) * 100%);
      }

      &.value {
        width: calc((743 / 1163) * 100%);
      }
    }

    ::v-deep(.table__row_cell) {
      &.created_at, &.action {
        width: calc((210 / 1163) * 100%);
      }

      &.value {
        width: calc((743 / 1163) * 100%);
      }
    }
  }
}

.table {
  &__actions {
    border-radius: 4px;
    background: #F3F4F6;
    height: 24px;
    display: flex;
    align-items: center;
    padding: 0 8px;
    width: fit-content;
  }
}

@media (max-width: 1200px) {
  .history-desktop {
    display: none;
  }
}
</style>