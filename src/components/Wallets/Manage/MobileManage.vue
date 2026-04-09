<template>
  <div class="manage-mobile">
    <DashboardHeader>
      <template #header-left>
        <div class="manage-mobile__header">
          <UIButton color_type="primary" size="large" @cta="handleOpenWalletModal('create')">
            Create wallets
          </UIButton>
          <UIButton color_type="ghost" size="large" @cta="handleOpenWalletModal('import')">
            <template #left-icon>
              <SVGImport/>
            </template>
            Import wallets
          </UIButton>
        </div>
      </template>
    </DashboardHeader>

    <div class="manage-mobile__table table">
      <UIMobileTable
        :rows="rows"
        :nested-rows="nestedRows"
      >
        <template #parent-top="{item, index}">
          <div class="table__parent">
            <div class="table__parent_left">
              <div class="paragraph-medium black label">{{ item.project }}</div>
              <div class="paragraph-small black regular">
                {{ formatDate(new Date().toISOString()).date }}
                <span class="grey">{{ formatDate(new Date().toISOString()).time }}</span>
              </div>
              <div class="paragraph-small black regular">{{ item.lifetime }}</div>
            </div>

            <div class="table__parent_right">
              <div class="monospaced-medium">{{ item.balance }} <span class="grey">{{ item.balance_usd }}</span></div>
              <OnClickOutside @trigger="closeUIKit(index)" class="frozen">
                <div class="tooltip-wrapper" @click="toggleUIKit(index)">
                  <div class="tooltip paragraph-mini">
                    <UIToolTip
                      position="bottom"
                      :is-shown="uiKitStates[index]"
                      text="Funds reserved by the network for wallet-related operations, such as associated token account creation."
                    />
                  </div>
                  <SVGAlertInfo color="#4B5563"/>
                </div>
                <div class="monospaced-medium">
                  {{ item.frozen_money }}
                  <span class="grey">{{ item.frozen_money_usd }}</span>
                </div>
              </OnClickOutside>
            </div>
          </div>
        </template>

        <template #nested-label="parent">
          <div class="paragraph-small regular">
            Wallets Q-ty
            <span class="bold">{{ parent.item.wallets_qty }}</span>
          </div>
        </template>

        <template #nested-item="nested">
          <div class="table__nested">
            <div class="table__nested_top">
              <div class="left">
                <div class="paragraph-small regular label">{{ nested.item.project }}</div>
                <div class="paragraph-small black regular">
                  {{ formatDate(new Date().toISOString()).date }}
                  <span class="grey">{{ formatDate(new Date().toISOString()).time }}</span>
                </div>
                <div class="paragraph-small black regular">{{ nested.item.lifetime }}</div>
              </div>
              <div class="right">
                <div class="monospaced-medium">
                  {{ nested.item.balance }}
                  <span class="grey">{{ nested.item.balance_usd }}</span>
                </div>
              </div>
            </div>
            <div class="table__nested_bottom">
              <UIButton color_type="ghost" size="small">
                <template #left-icon><SVGKey /></template>
              </UIButton>
              <UIButton class="map" size="small" color_type="primary">Map</UIButton>
            </div>
          </div>
        </template>

        <template #parent-actions="{item}">
          <div class="table__parent_actions">
            <UIButton
              color_type="secondary"
              size="large"
            >
              <template #left-icon>
                <SVGDownload/>
              </template>
              Export
            </UIButton>
            <UIButton
              color_type="secondary"
              size="large"
            >
              <template #left-icon>
                <SVGReload/>
              </template>
              Sync
            </UIButton>
          </div>
        </template>
      </UIMobileTable>
    </div>
  </div>
</template>
<script setup>
import DashboardHeader from "../../Base/DashboardHeader.vue";
import SVGImport from "../../SVG/SVGImport.vue";
import UIButton from "../../UI/UIButton.vue";
import UIMobileTable from "../../UI/UIMobileTable.vue";
import {formatDate} from "../../../helpers/index.js";
import SVGAlertInfo from "../../SVG/SVGAlertInfo.vue";
import UIToolTip from "../../UI/UIToolTip.vue";
import {ref, watch} from "vue";
import {OnClickOutside} from "@vueuse/components";
import SVGKey from "../../SVG/SVGKey.vue";
import SVGReload from "../../SVG/SVGReload.vue";
import SVGDownload from "../../SVG/SVGDownload.vue";

const props = defineProps({
  rows: {type: Array, default: []},
  nestedRows: {type: Array, default: []},
})
const emits = defineEmits(["openWalletModal", "getWalletPrivateKey"]);

const isFrozenMoneyToolTipVisible = ref(false);
const uiKitStates = ref({});

const toggleUIKit = (index) => {
  uiKitStates.value[index] = !uiKitStates.value[index];
}

const closeUIKit = (index) => {
  uiKitStates.value[index] = false;
}

const handleOpenWalletModal = (type) => {
  emits('openWalletModal', type);
}

watch(() => props.rows, (newVal) => {
  if (newVal.length) {
    newVal.forEach((_, i) => {
      uiKitStates.value[i] = false;
    })
  } else {
    uiKitStates.value = {};
  }
}, {deep: true})
</script>
<style scoped lang="scss">
.manage-mobile {
  display: none;
}

@media (max-width: 1200px) {
  .manage-mobile {
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
      align-items: center;
      gap: 10px;
      justify-content: space-between;

      &_left {
        display: flex;
        flex-direction: column;

        & .label {
          margin-bottom: 6px;
        }
      }

      &_right {
        display: flex;
        flex-direction: column;
        align-items: flex-end;

        & .frozen {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        & .tooltip {
          position: absolute;
          width: 250px;
          bottom: calc(100% + 10px);
          left: 50%;
          transform: translateX(-50%);
          &-wrapper {
            position: relative;
          }
        }
      }

      &_actions {
        display: flex;
        align-items: center;
        gap: 12px;
        width: 100%;

        & .ui-button {
          width: 100%;
        }
      }
    }

    &__nested {
      display: flex;
      flex-direction: column;
      gap: 20px;

      &_top {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 10px;

        & .label {
          margin-bottom: 8px;
          text-decoration: underline;
          color: #2563EB;
        }
      }

      &_bottom {
        display: flex;
        gap: 12px;

        & .map {
          width: 100%;
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