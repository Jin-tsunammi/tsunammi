<template>
  <div class="campaigns">
    <PageLoading v-if="isPageLoading"/>
    <template v-else>
      <div v-if="!campaigns.length" class="campaigns__empty">
        <SVGFolderOpenDot/>
        <div class="paragraph-medium bold">No campaigns yet</div>
        <p class="paragraph-small regular grey">Create your first campaign.</p>
      </div>

      <div v-else class="campaigns__content">
        <div class="campaigns__top">
          <UITabs size="large">
            <UITab
              v-for="tab in tabs"
              :key="tab.val"
              :is_active="selectedTab.val === tab.val"
              @click="handleTabSelect(tab)"
            >
              {{ tab.label }}
            </UITab>
          </UITabs>
        </div>
        <div class="campaigns__table">
          <UITable
            :columns="columns"
            :rows="campaigns"
            @handle-row-click="openCampaign"
          >
            <template #token_mint_from="{ item }">
              <div class="table__token">
                <div class="name"><span class="paragraph-small regular">{{ getTokenDetails(item) }}</span></div>
              </div>
            </template>
            <template #status="{ item }">
              <UIStatus :status="normilizeCampaignStatus(item.status)" />
            </template>
            <template #budget="{ item }">
              <div class="table__price monospaced-small">
                <span class="price">{{ '0' }}</span>
                <span class="token grey">{{ 'SOL' }}</span>
              </div>
            </template>
            <template #created_at="{ item }">
              <div class="created_at paragraph-small regular">
                <span>{{ formatDate(item.created_at || '').date }}</span>
                <span class="paragraph-small regular grey">{{ formatDate(item.created_at || '').time }}</span>
              </div>
            </template>
          </UITable>
        </div>
        <div v-if="totalPages > 1" class="campaigns__pagination">
          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>
        </div>
      </div>
      <MobileAdaptsNotification class="campaigns__mobile"/>
    </template>

    <Modals />
  </div>
</template>
<script setup>
import PageLoading from "../../components/UI/PageLoading.vue";
import {
  errorToast,
  formatDate,
  normilizeCampaignStatus,
} from "../../helpers/index.js";
import SVGFolderOpenDot from "../../components/SVG/SVGFolderOpenDot.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import UITable from "../../components/UI/UITable.vue";
import {useRoute, useRouter} from "vue-router";
import {useModalsStore} from "../../store/modalsStore.js";
import {computed, onMounted, ref, watch} from "vue";
import {GetAllCampaigns, GetSmartBuyBackHistory} from "../../api/api.js";
import {useTokensStore} from "../../store/tokensStore.js";
import Pagination from "../../components/UI/Pagination.vue";
import {useToastStore} from "../../store/toastStore.js";
import {useUserStore} from "../../store/userStore.js";
import Modals from "../../components/UI/Modals.vue";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";
import UITab from "../../components/UI/UITab.vue";
import UITabs from "../../components/UI/UITabs.vue";
import UIStatus from "../../components/UI/UIStatus.vue";

const columns = [
  {label: 'Date', field: 'created_at'},
  {label: 'Status', field: 'status'},
  {label: 'Token', field: 'token_mint_from'},
  {label: 'Budget', field: 'budget'},
  {label: 'MKT CAP', field: 'mkt_cap'},
];
const tabs = [
  {
    label: 'Price Boost',
    val: 'pull_up',
  },
  {
    label: 'Price Drop',
    val: 'pull_down',
  },
  {
    label: 'Smart Buy/Sell',
    val: 'smart',
  }
]

const router = useRouter();
const route = useRoute();
const modalsStore = useModalsStore();
const tokensStore = useTokensStore();
const toastStore = useToastStore();
const userStore = useUserStore();
const campaigns = ref([]);
const selectedTab = ref(tabs[0]);
const isPageLoading = ref(true);
const currentPage = ref(1);
const itemsOnPage = 12;
const totalItems = ref(0);

const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsOnPage) || 0;
})
const campaignType = computed(() => {
  return route.query.type || selectedTab.value.val;
})
const openCampaign = (campaign) => {
  if (!campaign) return;

  if (modalsStore.modalData.is_open) {
    modalsStore.closeModal()
  }

  const nextPageName = campaignType.value === 'smart' ? 'SmartBuyBackTransactions' : 'MarketTransactions';
  router.push({name: nextPageName, params: {campaign_id: campaign.id}, query: {type: route.query.type}});
}

const handleTabSelect = async(tab) => {
  if (!tab) return;

  selectedTab.value = tab;
  await router.push({query: {type: tab.val}});
  await getCampaign();
}

const getCampaign = async (isRefreshing=false) => {
  try {
    isPageLoading.value = true;

    if (userStore.isUserAuth) {
      const isSmartBuyBackSelected = selectedTab.value?.val === 'smart';

      const params = {
        page: currentPage.value,
        pageSize: itemsOnPage,
        type: isSmartBuyBackSelected ? null : selectedTab.value.val
      }

      let resp;

      if (isSmartBuyBackSelected) {
        resp = await GetSmartBuyBackHistory(params);
      } else {
        resp = await GetAllCampaigns(params);
      }

      campaigns.value = resp.data.campaigns;
      totalItems.value = resp.data.total;
      const tokens = [];
      resp.data.campaigns.forEach(campaign => {
        if (isSmartBuyBackSelected) {
          tokens.push({token_mint: campaign.token_mint});
        } else {
          if (campaign.type?.name?.toLowerCase() === 'pull up') {
            tokens.push({token_mint: campaign.token_mint_to});
          } else {
            tokens.push({token_mint: campaign.token_mint_from});
          }
        }
      });

      await tokensStore.updateSolTokensData(tokens, 'token_mint');

      if (isRefreshing) {
        toastStore.success({text: 'Page has been refreshed.'})
      }
    }
  } catch (e) {
    errorToast(e.response.data);
  } finally {
    isPageLoading.value = false;
  }
}

const getCampaignTokenMint = (campaign) => {
  if (!campaign) return '';
  const type = campaign.type?.name?.toLowerCase() || '';
  if (type === 'pull up') {
    return campaign.token_mint_to;
  } else {
    return campaign.token_mint_from;
  }
}

const getTokenDetails = (campaign) => {
  if (!campaign) return '';

  let mintType;
  switch (campaignType.value) {
    case 'smart':
      mintType = 'token_mint'
          break;
    case 'pull_up':
      mintType = 'token_mint_to';
      break;
    default:
      mintType = 'token_mint_from'
  }
  const mint = campaign[mintType];
  const tokenDetails = tokensStore.solTokensData[mint];
  const tokenName = tokenDetails?.name || '';
  const tokenSymbol = tokenDetails?.symbol || '';
  return `${tokenName} (${tokenSymbol})`;
}

const handlePageChange = async (page) => {
  if (!page) return;
  currentPage.value = page;
  await getCampaign();
}

useHeaderRefresh(() => getCampaign(true));

watch(() => userStore.isUserAuth, async(newVal) => {
  if (newVal) {
    await getCampaign();
  }
})
onMounted(async () => {
  if (route.query.type) {
    const findTab = tabs.find(tab => tab.val === route.query.type);

    if (findTab) {
      selectedTab.value = findTab;
    }
  } else {
    await router.push({query: {type: selectedTab.value.val}})
  }

  await getCampaign();
})
</script>
<style scoped lang="scss">
.campaigns {
  display: flex;
  flex-direction: column;
  height: 100%;

  &__content {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
    border-bottom: 1px solid #D1D5DB;
    padding-bottom: 12px;

    & .ui-tab {
      min-width: 136px;
    }

    & .ui-tabs.large {
      padding: 2px;
    }
  }

  &__mobile {
    display: none;
  }

  &__pagination {
    width: 100%;
    margin-top: auto;
    display: flex;
    align-items: center;
    justify-content: center;
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
      &.created_at, &.status {
        width: calc((220 / 1163) * 100%);
      }

      &.token_mint_from, &.budget, &.mkt_cap {
        width: calc((241 / 1163) * 100%);
      }
    }

    ::v-deep(.table__row) {
      height: 55px;
      transition: 0.3s ease;

      &:hover {
        cursor: pointer;
        background: rgba(229, 231, 235, .5);
      }
    }

    ::v-deep(.table__row_cell) {
      &.created_at, &.status {
        width: calc((220 / 1163) * 100%);
      }

      &.token_mint_from, &.budget, &.mkt_cap {
        width: calc((241 / 1163) * 100%);
      }
    }
  }
}

.table {
  &__token {
    display: flex;
    flex-direction: column;
    overflow: hidden;
    text-overflow: ellipsis;

    & .name,
    & .address {
      overflow: hidden;
      text-overflow: ellipsis;
      text-wrap: nowrap;

      & span {
        text-overflow: ellipsis;
        text-wrap: nowrap;
        overflow: hidden;
      }
    }
  }

  & .created_at {
    display: flex;
    gap: 5px;

    & .price {
      color: #030712;
    }
  }

  &__status {
    text-transform: capitalize;
  }

  &__status {
    display: flex;
    align-items: center;
    gap: 7px;

    & .indicator {
      aspect-ratio: 1/1;
      min-width: 6px;
      max-width: 6px;
      border-radius: 50%;
    }

    &.stopped {
      & .indicator {
        background: #DC2626;
      }
    }

    &.active, &.done {
      & .indicator {
        background: #16A34A;
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