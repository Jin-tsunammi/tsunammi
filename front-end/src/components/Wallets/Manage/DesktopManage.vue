<template>
  <div class="project-desktop">
    <div class="project-desktop__return">
      <router-link :to="{name: 'WalletsProjects'}" class="paragraph-small regular">
        <SVGSmallArrowDown color="#4B5563"/>
        All pools
      </router-link>
    </div>
    <div class="project-desktop__project">
      <div class="project-desktop__project_top">
        <div class="avatar">
          <DefaultAvatar />
        </div>
        <span class="heading-3">{{projectStore.selectedProject?.name || ''}}</span>
        <div class="controls">
          <div class="tooltip-wrapper" @mouseenter="toggleUIKit('export')" @mouseleave="toggleUIKit('export')">
            <UIButton
              color_type="ghost"
              size="small"
              :is_disabled="!projectStore.selectedProject?.wallets?.length"
              @cta="handleExportExcel"
            >
              Export XLSX
              <template #left-icon>
                <SVGUpload/>
              </template>
            </UIButton>

            <div class="tooltip paragraph-mini regular">
              <UIToolTip
                position="hidden"
                :is-shown="UIKitVisible === 'export'"
                text="No wallets available to export."
              />
            </div>
          </div>
          <UIDotsMenu
            :menu="projectDotsMenu"
            @handle-option-select="handleMenuOptionClick"
          />
        </div>
      </div>

      <div class="project-desktop__project_content">
        <div class="balances">
          <div class="balance">
            <span class="paragraph-mini regular grey">Total balance</span>
            <span class="monospaced-small">{{formatAmount(projectStore.selectedProject?.total_balance_sol) || 0}} Sol</span>
          </div>
          <div class="balance">
            <div class="balance__tooltip">
              <span class="paragraph-mini regular grey">Frozen Balance</span>
              <div class="tooltip-wrapper" @mouseenter="toggleUIKit('project-balance')" @mouseleave="toggleUIKit('project-balance')">
                <div class="tooltip paragraph-mini">
                  <UIToolTip
                    position="bottom"
                    :is-shown="UIKitVisible === 'project-balance'"
                    text="Funds reserved by the network for wallet-related operations, such as associated token account creation."
                  />
                </div>
                <SVGAlertInfo color="#4B5563"/>
              </div>
            </div>
            <span class="monospaced-small">{{formatAmount(projectStore.selectedProject?.rent_total) || 0}} Sol</span>
          </div>
        </div>
        <div class="information">
          <div class="paragraph-small regular grey">
            Last sync
            <div class="paragraph-small regular black">{{formatDate(projectStore.selectedProject?.last_sync).date}} <span class="paragraph-small regular grey">{{formatDate(projectStore.selectedProject?.last_sync).time}}</span></div>
          </div>
          <div class="paragraph-small regular grey">
            Project age
            <span class="black">{{daysSince(projectStore.selectedProject?.created_at)}}</span>
          </div>
          <div class="paragraph-small regular grey">
            Wallets quantity
            <span class="black">{{projectStore.selectedProject?.wallet_count || 0}}</span>
          </div>
        </div>
      </div>
    </div>
    <div v-if="!projectStore.selectedProject || !projectStore.selectedProject?.wallets?.length" class="project-desktop__empty">
      <SVGGlobus />
      <div class="project-desktop__empty_title paragraph-medium">No address pool</div>
      <p class="grey paragraph-small regular">This project doesn’t have any address pool yet. Create new address pool or import existing ones to start.</p>
      <div class="project-desktop__empty_btns">
        <UIButton color_type="outline" size="large" @cta="emits('openWalletModal', {type: 'create'})">
          Create wallets
        </UIButton>
        <UIButton color_type="outline" size="large" @cta="emits('openWalletModal', {type: 'import'})">
          <template #left-icon>
            <SVGImport/>
          </template>
          Import wallets
        </UIButton>
      </div>
    </div>

    <div v-else class="project-desktop__table">
      <div class="project-desktop__table_top">
        <div class="heading-5">All wallets</div>
        <div class="imports">
          <UIButton
            color_type="ghost"
            size="large"
            @cta="emits('openWalletModal', {type: 'import'})"
          >
            <template #left-icon>
              <SVGUpload/>
            </template>
            Import wallets
          </UIButton>
          <UIButton
            color_type="primary"
            size="large"
            @cta="emits('openWalletModal', {type: 'create'})"
          >
            Create wallet
          </UIButton>
        </div>
      </div>
      <UITable :columns="columns" :rows="rows" :is-table-nested="true">
        <template #col_frozen_money="{ item }">
          <div class="table__head frozen_money">
            <div class="tooltip-wrapper" @mouseenter="toggleUIKit('table-balance')" @mouseleave="toggleUIKit('table-balance')">
              <div class="tooltip paragraph-mini">
                <UIToolTip
                  position="bottom"
                  :is-shown="UIKitVisible === 'table-balance'"
                  text="Funds reserved by the network for wallet-related operations, such as associated token account creation."
                />
              </div>
              <SVGAlertInfo color="#4B5563"/>
            </div>
            <div class="paragraph-small medium">{{ item.label }}</div>
          </div>
        </template>
        <template #wallets="{ item }">
          <div class="table__wallet">
            <a
              :href="`${SOL_SCAN_BASE_URL}${item.public_key}`"
              target="_blank"
              class="monospaced-small regular"
            >
              {{ formatWalletAddress(item.public_key) }}
            </a>
            <UICopyText :copy-text="item.public_key" />
          </div>
        </template>
        <template #lifetime="{ item }">
          <span class="paragraph-small regular">{{ daysSince(item.created_at) }}</span>
        </template>
        <template #balance="{ item }">
          <div :class="['table__balance']">
            <div class="monospaced-small" @mouseenter="openWalletAsset(item, $event)" @mouseleave="closeWalletAsset">{{ calculateTotalWalletAmount(item) }} Sol</div>
            <div v-show="walletAssets === item.id" :class="['table__balance_tokens', { 'table__balance_tokens--above': balanceTokensAbove }]">
              <div class="table__balance_tokens-top">
                All assets
              </div>
              <ul class="table__balance_tokens-list">
                <li
                  v-if="tokensStore.solTokensData[SOLANA_MINT] && tokensStore.solTokensData[SOLANA_MINT]?.symbol"
                  class="table__balance_tokens-item"
                >
                  <div class="image">
                    <img v-if="isTokenPictureValid(SOLANA_MINT)" :src="tokensStore.solTokensData[SOLANA_MINT]?.image" alt="image">
                    <DefaultAvatar v-else />
                  </div>
                  <span class="symbol paragraph-small regular">{{tokensStore.solTokensData[SOLANA_MINT]?.symbol}}</span>
                  <span class=" amount monospaced-small">{{formatTokenAmount(item.balance_sol)}} Sol</span>
                </li>
                <template v-for="token in item?.tokens || []" :key="token?.token_symbol">
                  <li
                    v-if="tokensStore.solTokensData[token?.token_symbol] && tokensStore.solTokensData[token?.token_symbol]?.symbol"
                    class="table__balance_tokens-item"
                  >
                    <div class="image">
                      <img v-if="isTokenPictureValid(token?.token_symbol)" :src="tokensStore.solTokensData[token.token_symbol]?.image" alt="image">
                      <DefaultAvatar v-else />
                    </div>
                    <span class="symbol paragraph-small regular">{{tokensStore.solTokensData[token.token_symbol]?.symbol}}</span>
                    <span class=" amount monospaced-small">{{formatTokenAmount(token.balance_sol)}} Sol</span>
                  </li>
                </template>
              </ul>
            </div>
          </div>
        </template>
        <template #frozen_money="{ item }">
          <div class="monospaced-small">{{ formatAmount(item.rent) }} Sol</div>
        </template>
        <template #actions="{item}">
          <div class="table__actions">
            <UIDotsMenu
              :menu="walletDotsMenu"
              @handle-option-select="handleMenuTableClick($event, item)"
            />
          </div>
        </template>
      </UITable>
    </div>
  </div>
</template>
<script setup>
import UITable from "../../UI/UITable.vue";
import SVGKey from "../../SVG/SVGKey.vue";
import UIButton from "../../UI/UIButton.vue";
import {
  daysSince, errorToast,
  formatAmount,
  formatDate,
  formatWalletAddress,
  isTokenPictureValid,
  toDynamicFix
} from "../../../helpers/index.js";
import {SOL_SCAN_BASE_URL, SOLANA_MINT} from "../../../constants/const.js";
import {useProjectsStore} from "../../../store/projectsStore.js";
import SVGImport from "../../SVG/SVGImport.vue";
import SVGGlobus from "../../SVG/SVGGlobus.vue";
import SVGSmallArrowDown from "../../SVG/SVGSmallArrowDown.vue";
import SVGUpload from "../../SVG/SVGUpload.vue";
import UIDotsMenu from "../../UI/UIDotsMenu.vue";
import SVGEdit from "../../SVG/SVGEdit.vue";
import SVGDelete from "../../SVG/SVGDelete.vue";
import UIToolTip from "../../UI/UIToolTip.vue";
import {ref} from "vue";
import SVGAlertInfo from "../../SVG/SVGAlertInfo.vue";
import SVGGitPullRequest from "../../SVG/SVGGitPullRequest.vue";
import {useExcelExport} from "../../../composable/useExcelExport.js";
import {usePrivateKeysExport} from "../../../composable/usePrivateKeysExport.js";
import DefaultAvatar from "../../UI/DefaultAvatar.vue";
import {useTokensStore} from "../../../store/tokensStore.js";
import {GetWalletsPrivateKeys} from "../../../api/api.js";
import {useRoute} from "vue-router";
import UICopyText from "../../UI/UICopyText.vue";

defineProps({
  columns: {type: Array, default: []},
  rows: {type: Array, default: []},
})

const emits = defineEmits(['getWalletPrivateKey', 'openWalletModal']);
const route = useRoute();
const projectStore = useProjectsStore();
const tokensStore = useTokensStore();
const {exportProjectExcel} = useExcelExport();
const {exportPrivateKeysExcel} = usePrivateKeysExport();
const walletAssets = ref('');
const balanceTokensAbove = ref(false);
const BALANCE_TOKENS_APPROX_HEIGHT = 200;
const projectDotsMenu = [
  [{label: "Edit", icon: SVGEdit, action: "edit-project"}, {label: "Export Private keys", icon: SVGUpload, action: "export-keys"}],
  [{label: "Delete", icon: SVGDelete, action: "delete"}],
];
const walletDotsMenu = [
  [{label: "Security", icon: SVGKey, action: "secret-key"}, {label: "Map", icon: SVGGitPullRequest, action: "map"}],
];
const UIKitVisible = ref('');

const handleMenuOptionClick = (action) => {
  if (action === 'export-keys') {
    getAllPrivateKeys();
    return;
  }
  emits('openWalletModal', {type: action});
}

const openWalletAsset = (wallet, event) => {
  if (!wallet) return;

  walletAssets.value = wallet.id;
  if (event?.currentTarget) {
    const rect = event.currentTarget.getBoundingClientRect();
    balanceTokensAbove.value = rect.bottom + BALANCE_TOKENS_APPROX_HEIGHT > window.innerHeight;
  } else {
    balanceTokensAbove.value = false;
  }
}

const closeWalletAsset = () => {
  walletAssets.value = '';
}

const handleExportExcel = () => {
  exportProjectExcel({project: projectStore.selectedProject, wallets: projectStore.selectedProject?.wallets || []});
}

const handleMenuTableClick = (action, wallet=null) => {
  if (action === 'map') {
    if (!wallet) return;
    window.open(`https://v2.bubblemaps.io/map?address=${wallet?.public_key}&chain=solana&limit=80`)
  } else {
    if (!wallet) return;
    emits('getWalletPrivateKey', wallet.id);
  }
}

const calculateTotalWalletAmount = (wallet) => {
  if (!wallet) return 0;

  return toDynamicFix(wallet.balance_sol + wallet.tokens_balance_sol);
}

const formatTokenAmount = (amount) => {
  if (!amount) return '--';

  return toDynamicFix(amount);
}

const getAllPrivateKeys = async() => {
  try {
    const keysResp = await GetWalletsPrivateKeys(route.params.project_id);
    const keys = keysResp?.data?.data ?? keysResp?.data ?? [];
    if (!keys.length) {
      errorToast('No private keys to export.');
      return;
    }
    const projectName = projectStore.selectedProject?.name ?? 'project';
    exportPrivateKeysExcel(keys, projectName);
  } catch (e) {
    errorToast(e?.response?.data ?? e);
  }
}

const toggleUIKit = (type) => {
  if (type === 'export' && projectStore.selectedProject?.wallets?.length) return;

  if (UIKitVisible.value !== type) {
    UIKitVisible.value = type;
  } else {
    UIKitVisible.value = '';
  }
}
</script>
<style scoped lang="scss">
.project-desktop {
  & .tooltip {
    position: absolute;
    bottom: calc(100% + 10px);
    left: 50%;
    transform: translateX(-50%);
    width: 205px;
    z-index: 5;
    font-weight: 400;

    &-wrapper {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  &__return {
    display: flex;

    & a {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 24px;
      color: #111827;

      & svg {
        transform: rotate(90deg);
      }
    }
  }

  &__project {
    border-radius: 12px;
    background: #FFF;
    padding: 24px;
    margin-bottom: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;

    &_top {
      display: flex;
      align-items: center;

      & .tooltip {
        top: calc(100% + 5px);
      }

      & .controls {
        margin-left: auto;
        display: flex;
        align-items: center;
        ::v-deep(.ui-dots-menu__dropdown) {
          width: max-content;
        }
      }

      & .avatar {
        margin-right: 8px;
        aspect-ratio: 1/1;
        min-width: 32px;
        max-width: 32px;
      }
    }

    &_content {
      display: flex;
      align-items: flex-end;

      & .balances {
        display: flex;
        align-items: center;
        gap: 8px;
      }

      & .balance {
        width: 179px;
        height: 70px;
        display: flex;
        flex-direction: column;
        justify-content: center;
        padding: 12px;
        gap: 7px;
        border-radius: 8px;
        border: 1px solid #E5E7EB;

        &__tooltip {
          display: flex;
          align-items: center;
          gap: 8px;
        }
      }

      & .information {
        margin-bottom: 12px;
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: 32px;
        width: max-content;

        &>div {
          display: flex;
          align-items: center;
          gap: 4px;
        }
      }
    }
  }

  &__empty {
    display: flex;
    height: 309px;
    padding: 32px 0;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    flex-shrink: 0;
    align-self: stretch;
    border-radius: 12px;
    background: #FFF;

    & div {
      margin-top: 12px;
    }

    & p {
      max-width: 400px;
      text-align: center;
      margin: 4px 0 24px;
    }

    &_btns {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  &__table {
    width: 100%;
    max-width: 1400px;
    margin-bottom: 20px;

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 16px;

      & .imports {
        display: flex;
        align-items: center;
        gap: 12px;
      }
    }

    ::v-deep(.table__header_col) {
      &.wallets {
        width: calc((320 / 1163) * 100%);
      }

      &.lifetime {
        width: calc((160 / 1163) * 100%);
      }

      &.balance {
        width: calc((180 / 1163) * 100%);
      }

      &.frozen_money {
        width: calc((190 / 1163) * 100%);
      }

      &.actions {
        width: calc((281 / 1163) * 100%);
      }

      &.balance, &.frozen_money {
        display: flex;
        align-items: center;
        justify-content: flex-end;
      }
    }

    ::v-deep(.table__row_cell) {
      &.wallets {
        width: calc((320 / 1163) * 100%);
      }

      &.lifetime {
        width: calc((160 / 1163) * 100%);
      }

      &.balance {
        width: calc((180 / 1163) * 100%);
      }

      &.frozen_money {
        width: calc((190 / 1163) * 100%);
      }

      &.actions {
        width: calc((281 / 1163) * 100%);
      }

      &.balance, &.frozen_money {
        display: flex;
        justify-content: flex-end;
      }
    }
  }
}

.table {
  &__wallet {
    width: 100%;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;

    & a {
      display: block;
      overflow: hidden;
      width: 100%;
      text-overflow: ellipsis;
      color: #2563EB;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  &__balance {
    position: relative;

    &_tokens {
      position: absolute;
      display: flex;
      flex-direction: column;
      border-radius: 8px;
      overflow: hidden;
      background: #F3F4F6;
      box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.10), 0 4px 6px -4px rgba(0, 0, 0, 0.10);
      width: 230px;
      right: -30%;
      z-index: 2;
      top: calc(100% + 8px);

      &--above {
        top: auto;
        bottom: calc(100% + 8px);
      }

      &-top {
        padding: 12px;
        display: flex;
        align-items: center;
        background: #FFF;
      }

      &-list {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding: 16px 12px;
      }

      &-item {
        display: flex;
        align-items: center;

        & .image {
          aspect-ratio: 1/1;
          max-width: 22px;
          min-width: 22px;
          border-radius: 50%;
          overflow: hidden;
          margin-right: 8px;
        }

        & .amount {
          margin-left: auto;
        }
      }
    }
  }

  &__head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 8px;
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
  }

  &__map_btn {
    min-width: 112px;
  }
}

@media (max-width: $tablet) {
  .project-desktop {
    display: none;
  }
}
</style>