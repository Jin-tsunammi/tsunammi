<template>
  <div class="selected-cex-history table">
    <UIMobileTable
      :rows="rows"
    >
      <template #parent-top="{item, index}">
        <div class="table__parent">
          <div class="table__parent_left">
            <div class="paragraph-medium black label">{{ item.result }}</div>
            <div class="date paragraph-small black regular">
              {{ formatDate(new Date().toISOString()).date }}
              <span class="grey">{{ formatDate(new Date().toISOString()).time }}</span>
            </div>
            <div :class="['table_status medium black-2', item.status]">
              <div class="indicator"></div>

              {{item.status}}
            </div>
          </div>

          <div class="table__parent_right">
            <div class="monospaced-medium">{{ item.sum_for_deposit }}</div>
          </div>
        </div>
      </template>

      <template #nested-label="parent">
        <div class="paragraph-small regular">
          {{`View wallets (${parent.item.children?.length || 0})`}}
          <span class="bold">{{ parent.item.wallets_qty }}</span>
        </div>
      </template>

      <template #nested-item="nested">
        <div class="table__nested">
          <div class="table__nested_top">
            <div class="paragraph-small regular label">{{ nested.item.api }}</div>
            <div class="right">
              <div class="paragraph-small regular">
                Total balance after:
                <span class="monospaced-small regular grey">{{ nested.item.balance }}</span>
              </div>
            </div>
            <div class="paragraph-small black regular">{{ nested.item.lifetime }}</div>
          </div>
          <div class="table__nested_bottom">
            <div :class="['table_status medium black-2', nested.item.status, {nested: !nested.item.isProject}]">
              <div class="indicator">
                <SVGCloseSquare v-if="nested.item.status === 'unsuccess'" />
                <SVGCheckedSquare v-else />
              </div>
              {{nested.item.status}}
            </div>
            <UIButton color_type="ghost" size="small">
              <template #left-icon><SVGKey /></template>
            </UIButton>
          </div>
        </div>
      </template>
    </UIMobileTable>
  </div>
</template>
<script setup>
import {formatDate} from "../../../helpers/index.js";
import SVGKey from "../../SVG/SVGKey.vue";
import UIButton from "../../UI/UIButton.vue";
import UIMobileTable from "../../UI/UIMobileTable.vue";
import SVGCloseSquare from "../../SVG/SVGCloseSquare.vue";
import SVGCheckedSquare from "../../SVG/SVGCheckedSquare.vue";

defineProps({
  rows: {type: Array, default: []},
})
</script>
<style scoped lang="scss">
.selected-cex-history {
  display: none;
}
@media (max-width: 1200px) {
  .selected-cex-history {
    margin-top: 20px;
    display: flex;
    flex-direction: column;

    &__header {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    &__table {
      margin-top: 18px;
      padding: 0 16px;
    }
  }

  .table {
    &__parent {
      display: flex;
      align-items: flex-start;
      gap: 10px;
      justify-content: space-between;

      &_left {
        display: flex;
        flex-direction: column;

        & .date {
          margin-bottom: 6px;
        }
      }

      &_right {
        height: 100%;
        display: flex;
        flex-direction: column;
        align-items: flex-start;
      }
    }

    &__nested {
      display: flex;
      flex-direction: column;

      & .label {
        margin-bottom: 8px;
        text-decoration: underline;
        color: #2563EB;
      }

      &_bottom {
        display: flex;
        align-items: center;
        justify-content: space-between;
      }
    }

    &_status {
      text-transform: capitalize;
      display: flex;
      align-items: center;
      gap: 8px;

      &.nested {
        & .indicator {
          display: flex;
          min-width: auto;
          max-width: none;
        }
      }

      & .indicator {
        border-radius: 50%;
        aspect-ratio: 1/1;
        min-width: 8px;
        max-width: 8px;
      }

      &.done {
        & .indicator {
          background: #16A34A;
        }
      }

      &.in_progress {
        & .indicator {
          background: #D97706;
        }
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

    & .black {
      color: #030712;
    }
  }
}
</style>