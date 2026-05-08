<template>
<div class="selected-cex-desktop">
  <UISectionTitleWithBorder class="history">History</UISectionTitleWithBorder>
  <div v-if="!rows.length" class="selected-cex-desktop__empty">
    <SVGClockFading />
    <div class="selected-cex-desktop__empty_title paragraph-medium bold">No top-up history yet</div>
    <span class="grey regular paragraph-small">Once you top up wallets from CEX, transactions will appear here.</span>
  </div>

  <div v-else class="selected-cex-desktop__table">
    <UITable :columns="columns" :rows="rows" :nested-columns-name="'transactions'">
      <template #project_name="{ item }">
        <span v-if="item.project_id" class="paragraph-small medium">{{ item.project_name }}</span>
        <a
          v-else
          :href="`${SOL_SCAN_BASE_URL}${item.public_key}`"
          target="_blank"
          class="monospaced-medium regular"
        >
          {{ formatWalletAddress(item.public_key, 6) }}
        </a>
      </template>
      <template #status="{ item }">
        <div :class="['table_status paragraph-small medium black-2', handleStatusCheck(item, true), {nested: !item.project_id}]">
          <div v-if="item.project_id" class="indicator"></div>

          <div v-else class="indicator">
            <SVGCloseSquare v-if="item.status.toLowerCase() === 'failed'" />
            <SVGCheckedSquare v-else />
          </div>
          {{handleStatusCheck(item)}}
        </div>
      </template>
      <template #created_at="{ item }">
        <div v-if="item.project_id" class="paragraph-small regular">{{formatDate(item.created_at).date}} <span class="grey">{{ formatDate(item?.created_at).time }}</span></div>
      </template>
      <template #total_sum_sol="{ item }">
        <span v-if="item.project_id" class="monospaced-small black">{{ formatAmount(item.total_sum_sol) }}</span>
        <span v-else class="monospaced-small regular black">{{ item.sum_sol }}</span>
      </template>
      <template #result="{ item }">
        <div v-if="!item.project_id" class="paragraph-medium regular black">Balance after: <span class="monospaced-medium grey">{{formatAmount(item.balance_sol)}}</span></div>
      </template>
      <template #actions="{item, isExpanded}">
        <div class="table__actions">
          <button v-if="item.project_id" :class="['table__action_btn arrow', {open: isExpanded}]">
            <SVGSmallArrowDown color="#030712" class="arrow-down"/>
          </button>
        </div>
      </template>
    </UITable>
  </div>

<!--  <div v-if="true" class="selected-cex-desktop__pagination">-->
<!--    <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>-->
<!--  </div>-->
</div>
</template>
<script setup>
import SVGCloseSquare from "../../SVG/SVGCloseSquare.vue";
import UITable from "../../UI/UITable.vue";
import SVGSmallArrowDown from "../../SVG/SVGSmallArrowDown.vue";
import SVGCheckedSquare from "../../SVG/SVGCheckedSquare.vue";
import {formatAmount, formatDate, formatWalletAddress} from "../../../helpers/index.js";
import {computed, ref} from "vue";
import SVGClockFading from "../../SVG/SVGClockFading.vue";
import {SOL_SCAN_BASE_URL} from "../../../constants/const.js";
import UISectionTitleWithBorder from "../../UI/UISectionTitleWithBorder.vue";

defineProps({
  columns: {type: Array, default: []},
  rows: {type: Array, default: []},
  nestedColumns: {type: Array, default: []},
})
const currentPage = ref(1);
const itemsOnPage = 20;
const totalOrders = ref(300);

const totalPages = computed(() => {
  return Math.ceil(totalOrders.value / itemsOnPage);
})

const handlePageChange = async (page) => {
  currentPage.value = page;

  // await getUsers();
}

const handleStatusCheck = (project, isClass=false) => {
  if (!project) return '';

  if (!project.project_id && !isClass) return project.status;

  if (project.transactions?.some(transaction => transaction.status === 'PENDING' || transaction.status === 'AWAITING_APPROVAL')) {
    return 'pending'
  }

  if (project.transactions?.some(transaction => transaction.status === 'COMPLETED' || transaction.status === 'Success')) {
    return 'completed'
  }

  if (project.transactions?.every(transaction => transaction.status === 'FAILED')) {
    return 'failed'
  }

  return '';
}
</script>
<style scoped lang="scss">
.selected-cex-desktop {
  display: flex;
  flex-direction: column;
  flex-grow: 1;

  & .history {
    margin-top: 32px;
  }

  & .black-2 {
    color: #374151;
  }

  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: 100%;

    &_label {
      position: relative;
      color: #030712;

      &.top {
        max-width: 668px;
      }

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

  &__empty {
    border-radius: 12px;
    background: #FFF;
    margin-top: 32px;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    flex-grow: 1;
    max-height: 309px;
    padding: 32px 0;

    &_title {
      margin: 12px 0 4px;
    }

    & span {
      display: block;
      text-align: center;
      max-width: 317px;
    }
  }

  &__pagination {
    margin-inline: auto;
    margin-top: auto;
    display: flex;
    justify-content: center;
  }

  &__table {
    margin-top: 32px;
    width: 100%;
    max-width: 1400px;
    margin-bottom: 20px;

    ::v-deep(.table__header_col) {
      &.project_name {
        width: calc((200 / 1163) * 100%);
      }

      &.status {
        width: calc((160 / 1163) * 100%);
      }

      &.total_sum_sol {
        width: calc((190 / 1163) * 100%);
      }

      &.result {
        width: calc((383 / 1163) * 100%);
      }

      &.created_at {
        width: calc((160 / 1163) * 100%);
      }

      &.actions {
        width: calc((70 / 1163) * 100%);
      }
    }

    ::v-deep(.table__row_cell) {
      &.project_name {
        width: calc((200 / 1163) * 100%);
      }

      &.status {
        width: calc((160 / 1163) * 100%);
      }

      &.total_sum_sol {
        width: calc((190 / 1163) * 100%);
      }

      &.result {
        width: calc((383 / 1163) * 100%);
      }

      &.created_at {
        width: calc((160 / 1163) * 100%);
      }

      &.actions {
        width: calc((70 / 1163) * 100%);
      }
    }

    ::v-deep(.project_name) {
      overflow: hidden;
    }
  }
}

.table {
  & .nested-api {
    text-decoration: underline;
    color: #2563EB;
    font-weight: 400;
  }

  &__actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
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

      &.completed, &.pending, &.failed {
        & .indicator {
          background: transparent;
        }
      }
    }

    & .indicator {
      border-radius: 50%;
      aspect-ratio: 1/1;
      min-width: 8px;
      max-width: 8px;
    }

    &.completed {
      & .indicator {
        background: #16A34A;
      }
    }

    &.failed {
      & .indicator {
        background: #DC2626;
      }
    }

    &.pending {
      & .indicator {
        background: #D97706;
      }
    }
  }

  &__action_btn {
    background: transparent;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 36px;
    max-width: 36px;
    aspect-ratio: 1/1;


    & svg {
      transition: .3s ease;
    }

    &.open {
      & svg {
        transform: rotate(180deg);
      }
    }

    &.arrow {
      & svg {
        width: 11px;
        height: 11px;
      }
    }
  }
}
</style>