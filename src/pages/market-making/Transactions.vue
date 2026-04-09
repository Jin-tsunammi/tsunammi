<template>
  <div class="transactions">
    <PageLoading v-if="isPageLoading"/>
    <template v-else>
      <div class="transactions__return">
        <router-link :to="{name: 'MarketHistory'}" class="paragraph-small regular">
          <SVGSmallArrowDown color="#4B5563"/>
          All campaigns
        </router-link>
      </div>
      <div v-if="!transactions.length" class="transactions__empty">
        <SVGFolderOpenDot/>
        <div class="paragraph-medium bold">No transactions yet</div>
        <p class="paragraph-small regular grey">Check this page later.</p>
      </div>

      <div v-else class="transactions__content">
        <div class="transactions__top">
          <UISectionTitleWithBorder>
            {{ 'Transactions' }}
          </UISectionTitleWithBorder>
        </div>
        <div class="transactions__table">
          <UITable
            :columns="columns"
            :rows="transactions"
          >
            <template #token_mint_from="{ item }">
              <div class="table__val token">
                <div class="name"><span class="paragraph-small regular">{{
                    `${token?.name || 'unknown'} ${token?.symbol ? `(${token?.symbol})` : ''}`
                  }}</span></div>
                <div class="address">
                      <span class="paragraph-small regular grey">
                        {{formatWalletAddress(campaignTokenMint, 7)}}
                      </span>
                  <UICopyText :copy-text="campaignTokenMint" />
                </div>
              </div>
            </template>
            <template #transaction_hash="{ item }">
              <div class="table__val hash monospaced-small">
                <a
                  :href="`https://solscan.io/tx/${item.transaction_hash}`"
                  target="_blank"
                  class="monospaced-small regular"
                >
                  {{ formatWalletAddress(item.transaction_hash) }}
                </a>
                <UICopyText v-if="item.transaction_hash" :copy-text="item.transaction_hash" />
              </div>
            </template>
            <template #status="{ item }">
              <div class="table__val status monospaced-small">
                {{ item.status }}
              </div>
            </template>
            <template #message="{ item }">
              <div class="table__val message monospaced-small">
                {{ item.message }}
              </div>
            </template>
          </UITable>
        </div>
        <div v-if="totalPages > 1" class="transactions__pagination">
          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>
        </div>
      </div>

      <MobileAdaptsNotification class="transactions__mobile"/>
    </template>
  </div>
</template>
<script setup>
import PageLoading from "../../components/UI/PageLoading.vue";
import {errorToast, formatWalletAddress} from "../../helpers/index.js";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import UITable from "../../components/UI/UITable.vue";
import SVGFolderOpenDot from "../../components/SVG/SVGFolderOpenDot.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import {useRoute, useRouter} from "vue-router";
import {useTokensStore} from "../../store/tokensStore.js";
import {computed, onBeforeUnmount, onMounted, ref} from "vue";
import {GetCampaignAllTransactions} from "../../api/api.js";
import {useToastStore} from "../../store/toastStore.js";
import SVGSmallArrowDown from "../../components/SVG/SVGSmallArrowDown.vue";
import {useCampaignsStore} from "../../store/campaignsStore.js";
import UICopyText from "../../components/UI/UICopyText.vue";
import {SOL_SCAN_BASE_URL} from "../../constants/const.js";
import Pagination from "../../components/UI/Pagination.vue";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const route = useRoute();
const router = useRouter();
const tokensStore = useTokensStore();
const campaignStore = useCampaignsStore();
const toastStore = useToastStore();
const campaign = ref(null);
const transactions = ref([]);
const isPageLoading = ref(true);
const currentPage = ref(1);
const itemsOnPage = 12;
const totalItems = ref(0);
const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsOnPage) || 0;
})

const columns = [
  {label: 'Token', field: 'token_mint_from'},
  {label: 'Hash', field: 'transaction_hash'},
  {label: 'Status', field: 'status'},
  {label: 'Message', field: 'message'},
];

const campaignTokenMint = computed(() => {
  if (!campaign.value) return '';
  const type = campaign.value.type?.name?.toLowerCase() || '';
  if (type === 'pull up') {
    return campaign.value.token_mint_to;
  } else {
    return campaign.value.token_mint_from;
  }
})

const token = computed(() => {
  if (!campaignTokenMint.value) return null;

  return tokensStore.solTokensData[campaignTokenMint.value] || null;
})

const getTransactions = async (isRefreshing=false) => {
  if (!route.params.campaign_id) {
    toastStore.error({text: "Failed to open campaign"})
    router.back();
  } else {
    const params = {
      page: currentPage.value,
      pageSize: itemsOnPage,
    }

    try {
      isPageLoading.value = true;
      campaign.value = await campaignStore.getCampaign(route.params.campaign_id);
      const sourceToken = [{ source_token_mint: campaignTokenMint.value }];
      await tokensStore.updateSolTokensData(sourceToken, 'source_token_mint');

      const resp = await GetCampaignAllTransactions(route.params.campaign_id, params);

      if (Array.isArray(resp.data)) {
        transactions.value = resp.data;
      } else {
        transactions.value = resp.data.transactions;
      }
      totalItems.value = resp.data.total || 1;

      if (isRefreshing) {
        toastStore.success({text: 'Page has been refreshed.'})
      }
    } catch (e) {
      errorToast(e.response.data);
    } finally {
      isPageLoading.value = false;
    }
  }
}

const handlePageChange = async (page) => {
  if (!page) return;
  currentPage.value = page;
  await getTransactions();
}
useHeaderRefresh(() => getTransactions(true));
onMounted(async () => {
  await getTransactions();
})

onBeforeUnmount(() => {
  campaignStore.clearStore();
})
</script>
<style scoped lang="scss">
.transactions {
  height: fit-content;
  display: flex;
  flex-direction: column;

  &__pagination {
    width: 100%;
    margin-top: auto;
    display: flex;
    align-items: center;
    justify-content: center;
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

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
  }

  &__mobile {
    display: none;
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
      max-width: 339px;
      text-align: center;
      margin: 4px 0 24px;
    }
  }

  &__table {
    width: 100%;
    max-width: 1400px;
    margin-bottom: 20px;

    ::v-deep(.table__header_col) {
      &.token_mint_from {
        width: calc((200 / 1163) * 100%);
      }

      &.transaction_hash {
        width: calc((300 / 1163) * 100%);
      }

      &.status {
        width: calc((200 / 1163) * 100%);
      }

      &.message {
        width: calc((463 / 1163) * 100%);
      }
    }

    ::v-deep(.table__row) {
      height: 55px;
      background: transparent;
    }

    ::v-deep(.table__row_cell) {
      &.token_mint_from {
        width: calc((200 / 1163) * 100%);
      }

      &.transaction_hash {
        width: calc((300 / 1163) * 100%);
      }

      &.status {
        width: calc((200 / 1163) * 100%);
      }

      &.message {
        width: calc((463 / 1163) * 100%);
      }
    }
  }
}

.table {
  &__val {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &.type, &.message {
      text-transform: capitalize;
    }

    &.hash {
      color: #2563EB;
    }

    & .address, &.hash, & .name {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 6px;
      overflow: hidden;
      text-overflow: ellipsis;
      text-wrap: nowrap;

      & .ui-copy-text {
        width: 16px;
        height: 16px;
        margin-left: 0;
      }

      & span {
        text-overflow: ellipsis;
        text-wrap: nowrap;
        overflow: hidden;
      }
    }
  }

  &__token {
    display: flex;
    flex-direction: column;
    overflow: hidden;

    & .name {
      overflow: hidden;

      & span {
        text-overflow: ellipsis;
        text-wrap: nowrap;
        overflow: hidden;
      }
    }
  }

  &__status {
    display: flex;
    align-items: center;
    gap: 7px;
  }
}

@media (max-width: 1200px) {
  .wallet-project {
    &__content {
      display: none;
    }

    &__mobile {
      margin-top: 16px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>