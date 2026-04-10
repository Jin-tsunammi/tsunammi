<template>
  <div class="cex-api-mobile">
    <DashboardHeader>
      <template #header-left>
        <div class="cex-api-mobile__header">
          <UIButton color_type="primary" size="large" @cta="handleOpenModal('add-api')">
            Add new API
          </UIButton>
          <UIButton color_type="ghost" size="large" @cta="">
            How to connect API?
          </UIButton>
        </div>
      </template>
    </DashboardHeader>

    <div class="cex-api-mobile__table table">
      <UIMobileTable
        :rows="rows"
      >
        <template #parent-top="{item, index}">
          <div class="table__parent">
            <div class="table__parent_left">
              <div class="table__title paragraph-medium label">{{ item.title }}</div>
              <div class="paragraph-small black regular">
                {{ formatDate(new Date().toISOString()).date }}
                <span class="grey">{{ formatDate(new Date().toISOString()).time }}</span>
              </div>
              <div class="paragraph-small black regular">{{ item.lifetime }}</div>
              <div :class="['table__status medium', item.status]">
                <div class="indicator"></div>
                {{item.status}}
              </div>
            </div>

            <div class="table__parent_right">
              <div class="monospaced-medium">{{item.deposited}} <span class="grey">{{ item.deposited_usd }}</span></div>
              <div class="api-name paragraph-small black regular">API name <span class="bold">{{ item.api }}</span></div>
              <div class="table__actions">
                <button class="table__action_btn" @click.stop="emits('openCEXModal', {type: 'delete', item})">
                  <SVGDelete />
                </button>
                <button class="table__action_btn" @click.stop>
                  <SVGRefresh />
                </button>
              </div>
            </div>
          </div>
        </template>
      </UIMobileTable>
    </div>
  </div>
</template>
<script setup>
import DashboardHeader from "../../Base/DashboardHeader.vue";
import UIButton from "../../UI/UIButton.vue";
import UIMobileTable from "../../UI/UIMobileTable.vue";
import {formatDate} from "../../../helpers/index.js";
import SVGDelete from "../../SVG/SVGDelete.vue";
import SVGRefresh from "../../SVG/SVGRefresh.vue";
import {useRouter} from "vue-router";

defineProps({
  rows: {type: Array, default: []},
})
const emits = defineEmits(["openModal", "openCEXModal"]);

const router = useRouter();

const handleOpenModal = (type) => {
  emits('openModal', type);
}

</script>
<style scoped lang="scss">
.cex-api-mobile {
  display: none;
}

@media (max-width: 1200px) {
  .cex-api-mobile {
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
      height: 121px;
      display: flex;
      align-items: center;
      gap: 10px;
      justify-content: space-between;

      &_left {
        height: 100%;
        display: flex;
        flex-direction: column;

        & .label {
          margin-bottom: 6px;
        }
      }

      &_right {
        height: 100%;
        display: flex;
        flex-direction: column;
        align-items: flex-end;

        & .api-name {
          margin-top: auto;
        }
      }
    }

    &__status {
      margin-top: auto;
      text-transform: capitalize;
      display: flex;
      align-items: center;
      height: 36px;
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

    &__actions {
      display: flex;
      align-items: center;
      justify-content: flex-end;
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