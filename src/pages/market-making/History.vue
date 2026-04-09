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
          <UISectionTitleWithBorder>
            {{ 'Campaigns' }}
          </UISectionTitleWithBorder>
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
                <div class="address"><span
                  class="paragraph-small regular grey">{{
                    formatWalletAddress(getCampaignTokenMint(item), 7)
                  }}</span>
                </div>
              </div>
            </template>
            <template #status="{ item }">
              <div :class="['table__status paragraph-small regular', campaignStatus(item)]">
                <div class="indicator"></div>
                <span>{{ campaignStatus(item) }}</span>
              </div>
            </template>
            <template #type="{ item }">
              <div :class="['table__type paragraph-small regular']">
                <span>{{ campaignType(item) }}</span>
              </div>
            </template>
            <template #project_id="{ item }">
              <span class="paragraph-small regular">{{ item.project_id }}</span>
            </template>
            <template #budget="{ item }">
              <div class="table__price monospaced-small">
                <span class="price">{{ toDynamicFix(item.budget) }}</span>
                <span class="token grey">{{ token(item, true)?.symbol || '' }}</span>
              </div>
            </template>
            <template #goal_percent_change="{ item }">
              <span class="monospaced-small">{{ toDynamicFix(item.goal_percent_change || 0) }}%</span>
            </template>
            <template #goal_price="{ item }">
              <div class="table__price  monospaced-small">
                <span class="price">{{ toDynamicFix(item.goal_price) }}</span>
                <span class="token grey">{{ token(item)?.symbol || '' }}</span>
              </div>
            </template>
            <template #current_price="{ item }">
              <div class="table__price  monospaced-small">
                <span class="price">{{ toDynamicFix(item.current_price) }}</span>
                <span class="token grey">{{ token(item)?.symbol || '' }}</span>
              </div>
            </template>
            <template #created_at="{ item }">
              <div class="paragraph-small regular">{{ formatDate(item.created_at || '').date }} <span
                class="paragraph-small regular grey">{{ formatDate(item.created_at || '').time }}</span></div>
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
import {errorToast, formatDate, formatWalletAddress, toDynamicFix} from "../../helpers/index.js";
import SVGFolderOpenDot from "../../components/SVG/SVGFolderOpenDot.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import UITable from "../../components/UI/UITable.vue";
import {useRouter} from "vue-router";
import {useModalsStore} from "../../store/modalsStore.js";
import {computed, onMounted, ref, watch} from "vue";
import {GetAllCampaigns} from "../../api/api.js";
import {useTokensStore} from "../../store/tokensStore.js";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import Pagination from "../../components/UI/Pagination.vue";
import {useToastStore} from "../../store/toastStore.js";
import {useUserStore} from "../../store/userStore.js";
import Modals from "../../components/UI/Modals.vue";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const router = useRouter();
const modalsStore = useModalsStore();
const tokensStore = useTokensStore();
const toastStore = useToastStore();
const userStore = useUserStore();
const campaigns = ref([]);
const isPageLoading = ref(true);
const currentPage = ref(1);
const itemsOnPage = 12;
const totalItems = ref(0);
const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsOnPage) || 0;
})

const columns = [
  {label: 'Token', field: 'token_mint_from'},
  {label: 'Project', field: 'project_id'},
  {label: 'Type', field: 'type'},
  {label: 'Budget', field: 'budget'},
  {label: 'Target', field: 'goal_percent_change'},
  {label: 'Goal price', field: 'goal_price'},
  {label: 'Current price', field: 'current_price'},
  {label: 'Status', field: 'status'},
  {label: 'Created', field: 'created_at'},
];

const openCampaign = (campaign) => {
  if (!campaign) return;

  if (modalsStore.modalData.is_open) {
    modalsStore.closeModal()
  }
  router.push({name: 'MarketTransactions', params: {campaign_id: campaign.id}});
}

const token = (campaign, isBudget=false) => {
  if (!campaign) return null;
  const type = campaign.type?.name.toLowerCase();

  if (type === 'pull up' && isBudget) return {symbol: 'Sol'};

  const mint = getCampaignTokenMint(campaign);

  return tokensStore.solTokensData[mint];
}

const getCampaign = async (isRefreshing=false) => {
  try {
    isPageLoading.value = true;

    if (userStore.isUserAuth) {
      const params = {
        page: currentPage.value,
        pageSize: itemsOnPage,
      }

      const resp = await GetAllCampaigns(params);
      campaigns.value = resp.data.campaigns;
      totalItems.value = resp.data.total;
      const tokens = [];
      resp.data.campaigns.forEach(campaign => {
        if (campaign.type?.name?.toLowerCase() === 'pull up') {
          tokens.push({token_mint: campaign.token_mint_to});
        } else {
          tokens.push({token_mint: campaign.token_mint_from});
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
  const tokenDetails = token(campaign);
  const tokenName = tokenDetails?.name || '';
  const tokenSymbol = tokenDetails?.symbol || '';
  return `${tokenName} (${tokenSymbol})`;
}

const campaignStatus = (campaign) => {
  if (!campaign) return '';
  const status = campaign.status;

  switch (status) {
    case 'in_use':
      return 'active';
    case 'stop':
      return 'stopped';
    case 'done':
      return 'done';
    default:
      return '';
  }
}

const campaignType = (campaign) => {
  if (!campaign) return '';

  return campaign.type?.name?.toLowerCase() || '';
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
  await getCampaign();
})
</script>
<style scoped lang="scss">
.campaigns {
  display: flex;
  flex-direction: column;
  height: fit-content;

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
      &.token_mint_from {
        width: calc((200 / 1163) * 100%);
      }

      &.created_at {
        width: calc((135 / 1163) * 100%);
      }

      &.project_id, &.goal_percent_change {
        width: calc((88 / 1163) * 100%);
      }

      &.type, &.status {
        width: calc((88 / 1163) * 100%);
      }

      &.budget, &.goal_price, &.current_price {
        width: calc((160 / 1163) * 100%);
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
      &.token_mint_from {
        width: calc((200 / 1163) * 100%);
      }

      &.created_at {
        width: calc((135 / 1163) * 100%);
      }

      &.project_id, &.goal_percent_change {
        width: calc((88 / 1163) * 100%);
      }

      &.type, &.status {
        width: calc((88 / 1163) * 100%);
      }

      &.budget, &.goal_price, &.current_price {
        width: calc((160 / 1163) * 100%);
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

  &__price {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 5px;
    min-width: 0;

    & .price {
      flex-shrink: 0;
    }

    & .token {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      text-wrap: nowrap;
    }
  }

  &__type, &__status {
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