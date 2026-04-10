<template>
  <div class="cex-api-desktop">
    <UISectionTitleWithBorder v-if="rows.length">
      CEX API
      <template #actions>
        <UIButton color_type="primary" size="large" @cta="emits('openCEXModal', {type: 'add-api'})">
          Add new API
        </UIButton>
      </template>
    </UISectionTitleWithBorder>

    <div v-if="!rows.length" class="cex-api-desktop__empty">
      <SVGPlug/>
      <div class="title paragraph-medium">No CEX API connected</div>
      <span class="paragraph-small">Add your exchange API to enable wallet top-ups and automated operations.</span>
      <div class="cex-api-desktop__empty_btns">
        <UIButton color_type="outline" size="large" @cta="emits('openCEXModal', {type: 'add-api'})">
          Add new API
        </UIButton>
      </div>
    </div>
    <div v-else class="cex-api-desktop__table">
      <UITable :columns="columns" :rows="rows">
        <template #title="{ item }">
          <div class="table__title">
            <span class="paragraph-small regular">{{ item.name }}</span>
          </div>
        </template>
        <template #status="{ item }">
          <div :class="['table_status paragraph-small medium', item.status]">
            <div class="indicator"></div>
            {{item.status}}
          </div>
        </template>
        <template #date="{ item }">
          <div class="paragraph-small regular">{{formatDate(item.created_at).date}} <span class="grey">{{ formatDate(item.created_at).time }}</span></div>
        </template>
        <template #lifetime="{ item }">
          <span class="paragraph-small regular">{{ daysSince(item.created_at) }}</span>
        </template>
        <template #api="{ item }">
          <span class="paragraph-small regular">{{ item.api_name }}</span>
        </template>
        <template #deposited="{ item }">
          <div class="monospaced-small">{{toDynamicFix(item.total_deposits_sol)}} Sol</div>
        </template>
        <template #actions="{item}">
          <div class="table__actions">
            <button class="table__action_btn" @click.stop="emits('openCEXModal', {type: 'delete', item})">
              <SVGDelete />
            </button>
            <button class="table__action_btn" @click.stop="cexApiStore.updateCEX(item.id)">
              <SVGRefresh />
            </button>
          </div>
        </template>
      </UITable>
    </div>
  </div>
</template>
<script setup>
import UITable from "../../UI/UITable.vue";
import SVGDelete from "../../SVG/SVGDelete.vue";
import SVGRefresh from "../../SVG/SVGRefresh.vue";
import {daysSince, formatDate, toDynamicFix} from "../../../helpers/index.js";
import {useRouter} from "vue-router";
import UIButton from "../../UI/UIButton.vue";
import SVGPlug from "../../SVG/SVGPlug.vue";
import {useCEXApiStore} from "../../../store/cexStore.js";
import UISectionTitleWithBorder from "../../UI/UISectionTitleWithBorder.vue";

defineProps({
  columns: {type: Array, default: []},
  rows: {type: Array, default: []},
})
const emits = defineEmits(['openCEXModal']);
const cexApiStore = useCEXApiStore();
const router = useRouter();
</script>
<style scoped lang="scss">
.cex-api-desktop {
  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: 100%;

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

  &__empty {
    border-radius: 12px;
    background: #FFF;
    height: 309px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;

    & .title {
      margin-top: 12px;
      margin-bottom: 4px;
      font-weight: 600;
    }

    & span {
      color: #6B7280;
      text-align: center;
      max-width: 317px;
      font-weight: 400;
    }

    &_btns {
      margin-top: 20px;
      display: flex;
      align-items: center;
      gap: 10px;
    }
  }

  &__table {
    margin-top: 20px;
    width: 100%;
    max-width: 1400px;

    ::v-deep(.table__header_col) {
      &.title, &.status, &.date, &.lifetime, &.api, &.deposited {
        width: calc((170 / 1163) * 100%);
      }


      &.actions {
        width: calc((143 / 1163) * 100%);
      }

      &.deposited {
        display: flex;
        justify-content: flex-end;
      }
    }

    ::v-deep(.table__row_cell) {
      &.title, &.status, &.date, &.lifetime, &.api, &.deposited {
        width: calc((170 / 1163) * 100%);
      }


      &.actions {
        width: calc((143 / 1163) * 100%);
      }

      &.deposited {
        display: flex;
        justify-content: flex-end;
      }
    }
  }
}

.table {
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

    & .indicator {
      border-radius: 50%;
      aspect-ratio: 1/1;
      min-width: 8px;
      max-width: 8px;
    }

    &.active {
      & .indicator {
        background: #16A34A;
      }

      & .disconnect {
        background: #DC2626;
      }
    }

    &.disconnect {
      & .indicator {
        background: #DC2626;
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

@media (max-width: 1200px) {
  .cex-api-desktop {
    display: none;
  }
}
</style>