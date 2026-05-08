<template>
  <div class="transactions">
    <div class="transactions__return">
      <router-link :to="{name: 'MarketHistory', query: {type: route.query.type}}" class="paragraph-small regular">
        <SVGSmallArrowDown color="#4B5563"/>
        All campaigns
      </router-link>
    </div>

    <CampaignDetailsHeader
      :campaign="campaign"
      :is-page-loading="isPageLoading"
      :transactions="transactions"
      :wallet-pools="walletPools"
    />
    <div class="transactions__bottom">
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
            <template #address_from="{ item }">
              <div class="table__val hash monospaced-small">
                <a
                  :href="`https://solscan.io/tx/${item.address_from}`"
                  target="_blank"
                  class="monospaced-small regular"
                >
                  {{ formatWalletAddress(item.address_from) }}
                </a>
                <UICopyText v-if="item.address_from" :copy-text="item.address_from" />
              </div>
            </template>
            <template #address_to="{ item }">
              <div class="table__val hash monospaced-small">
                <a
                  :href="`https://solscan.io/tx/${item.address_to}`"
                  target="_blank"
                  class="monospaced-small regular"
                >
                  {{ formatWalletAddress(item.address_to) }}
                </a>
                <UICopyText v-if="item.address_to" :copy-text="item.address_to" />
              </div>
            </template>
            <template #status="{ item }">
              <UIStatus class="table__status" :status="{status: item?.status, tooltip: item?.message}" />
            </template>
            <template #created_at="{ item }">
              <div class="paragraph-small regular black">{{formatDate(item?.created_at || '').date}} <span class="paragraph-small regular grey">{{formatDate(item?.created_at || '').time}}</span></div>
            </template>
            <template #amount_token_from="{ item }">
              <div class="table__val amount message monospaced-small">
                {{ `${toDynamicFix(item.amount_token_from)} ${getAmountTokenSymbol(item?.token_mint_from)}` }}
              </div>
            </template>
          </UITable>
        </div>
        <div v-if="totalPages > 1" class="transactions__pagination">
          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>
        </div>
      </div>

      <MobileAdaptsNotification class="transactions__mobile"/>
    </div>
  </div>
</template>
<script setup>
import {errorToast, formatDate, formatWalletAddress, toDynamicFix} from "../../helpers/index.js";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import UITable from "../../components/UI/UITable.vue";
import SVGFolderOpenDot from "../../components/SVG/SVGFolderOpenDot.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import {useRoute, useRouter} from "vue-router";
import {useTokensStore} from "../../store/tokensStore.js";
import {computed, onBeforeUnmount, onMounted, ref} from "vue";
import {
  GetAllProjectsNameOnly,
  GetCampaignAllTransactions,
  GetCampaignByID,
  GetSmartBuyBack,
  GetSmartBuyBackTransactions
} from "../../api/api.js";
import {useToastStore} from "../../store/toastStore.js";
import SVGSmallArrowDown from "../../components/SVG/SVGSmallArrowDown.vue";
import {useCampaignsStore} from "../../store/campaignsStore.js";
import UICopyText from "../../components/UI/UICopyText.vue";
import Pagination from "../../components/UI/Pagination.vue";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";
import CampaignDetailsHeader from "../../components/MarketMakingPages/CampaignDetailsHeader.vue";
import UIStatus from "../../components/UI/UIStatus.vue";

const route = useRoute();
const router = useRouter();
const tokensStore = useTokensStore();
const campaignStore = useCampaignsStore();
const toastStore = useToastStore();
const campaign = ref(null);
const transactions = ref([]);
const walletPools = ref([]);
const isPageLoading = ref(true);
const currentPage = ref(1);
const itemsOnPage = 12;
const totalItems = ref(0);
const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsOnPage) || 0;
})

const columns = [
  {label: 'Time', field: 'created_at'},
  {label: 'Status', field: 'status'},
  {label: 'Hash', field: 'transaction_hash'},
  {label: 'From', field: 'address_from'},
  {label: 'To', field: 'address_to'},
  {label: 'Amount', field: 'amount_token_from'},
];
const isSmartCampaign = computed(() => {
  return route.query.type === 'smart';
})
const campaignTokenMint = computed(() => {
  if (!campaign.value) return '';
  const type = campaign.value.type?.name?.toLowerCase() || '';
  if (type === 'pull up') {
    return campaign.value.token_mint_to;
  } else if (type === 'pull down') {
    return campaign.value.token_mint_from;
  } {
    return campaign.value.token_mint;
  }
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
      let transResp;
      let campaignResp;

      if (isSmartCampaign.value) {
        campaignResp = await GetSmartBuyBack(route.params.campaign_id);
        transResp = await GetSmartBuyBackTransactions(route.params.campaign_id, params);
      } else {
        campaignResp = await GetCampaignByID(route.params.campaign_id);
        transResp = await GetCampaignAllTransactions(route.params.campaign_id, params);
      }

      if (campaignResp?.data) {
        campaign.value = campaignResp.data;
      }

      if (Array.isArray(transResp.data)) {
        transactions.value = transResp.data;
      } else {
        transactions.value = transResp.data.transactions;
      }
      totalItems.value = transResp.data.total || 1;

      const sourceToken = [{ source_token_mint: campaignTokenMint.value }];
      await tokensStore.updateSolTokensData(sourceToken, 'source_token_mint');

      const walletPoolsResp = await GetAllProjectsNameOnly();
      walletPools.value = walletPoolsResp.data;

      if (isRefreshing) {
        toastStore.success({text: 'Page has been refreshed.'})
      }
    } catch (e) {
      errorToast(e.response.data);
      setTimeout(() => {
        router.push({ name: 'MarketHistory', query: {type: route.query.type} });
      }, 1000)
    } finally {
      isPageLoading.value = false;
    }
  }
}

const getAmountTokenSymbol = (mint='') => {
  const token = tokensStore.solTokensData[mint];

  return token?.symbol ? token.symbol : '$TOKEN';
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
      &.created_at {
        width: calc((125 / 1163) * 100%);
      }

      &.transaction_hash, &.address_from, &.address_to {
        width: calc((248 / 1163) * 100%);
      }

      &.status {
        width: calc((120 / 1163) * 100%);
      }

      &.amount_token_from {
        width: calc((167 / 1163) * 100%);
      }
    }

    ::v-deep(.table__row) {
      height: 55px;
    }

    ::v-deep(.table__row_cell) {
      &.created_at {
        width: calc((125 / 1163) * 100%);
      }

      &.transaction_hash, &.address_from, &.address_to {
        width: calc((248 / 1163) * 100%);
      }

      &.status {
        width: calc((120 / 1163) * 100%);
      }

      &.amount_token_from {
        width: calc((167 / 1163) * 100%);
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

    &.hash {
      color: #2563EB;
    }

    &.amount {
      display: flex;
      justify-content: flex-end;
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

  &__status {
    ::v-deep(.ui-tooltip__wrapper) {
      max-width: 250px;
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