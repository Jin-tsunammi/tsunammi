<template>
  <div class="history-mobile">
    <DashboardHeader>
      <template #header-left>
        <div class="history-mobile__header">
          <UIButton color_type="primary" size="large" @cta="openWalletModal('create')">
            Add new API
          </UIButton>
          <UISelect
            size="large"
            placeholder="Select time"
            v-model="isSelectTimeOptionOpen"
            :selected="selectedTime"
          >
            <template #left-icon>
              <SVGClock/>
            </template>
          </UISelect>
        </div>
      </template>
    </DashboardHeader>

    <div class="history-mobile__content">
      <div class="history-mobile__content_label heading-3">History</div>

      <div class="history-mobile__table table">
        <UIMobileTable
          :rows="rows"
        >
          <template #parent-top="{item}">
            <div class="table__parent">
              <div class="table__parent_left">
                <div class="paragraph-small regular">{{formatDate(new Date().toISOString()).date}} <span class="grey">{{ formatDate(new Date().toISOString()).time }}</span></div>
                <span class="paragraph-small bold">{{ item.value }}</span>
              </div>
              <div class="table__parent_actions">
                <UIButton color_type="secondary" size="small">Add API</UIButton>
              </div>
            </div>
          </template>
        </UIMobileTable>
      </div>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../../UI/UIButton.vue";
import {formatDate} from "../../../helpers/index.js";
import DashboardHeader from "../../Base/DashboardHeader.vue";
import UIMobileTable from "../../UI/UIMobileTable.vue";
import SVGClock from "../../SVG/SVGClock.vue";
import UISelect from "../../UI/UISelect.vue";

defineProps({
  columns: {type: Array, default: []},
  rows: {type: Array, default: []},
  nestedColumns: {type: Array, default: []},
})
</script>
<style scoped lang="scss">
.history-mobile {
  display: none;
}

@media (max-width: 1200px) {
  .history-mobile {
    display: flex;
    flex-direction: column;

    &__header {
      width: 100%;
      display: flex;
      align-items: center;
      gap: 12px;

      ::v-deep(.ui-select__input) {
        width: 100%;
        min-width: 177px;
      }
    }

    &__content {
      padding: 16px;

      &_label {
        position: relative;
        color: #030712;

        &::after {
          content: '';
          position: absolute;
          bottom: -12px;
          left: 0;
          width: 100%;
          height: 1px;
          background: #030712;
        }
      }
    }

    &__table {
      margin-top: 32px;
      width: 100%;
      max-width: 1400px;
    }
  }

  .table {
    &__parent {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 10px;

      &_left {
        display: flex;
        flex-direction: column;
      }
    }
    & .bold {
      font-weight: 600;
    }

    & .medium {
      font-weight: 500;
    }

    & .regular {
      font-weight: 400;
    }

    & .grey {
      color: #6B7280;
    }
  }
}
</style>