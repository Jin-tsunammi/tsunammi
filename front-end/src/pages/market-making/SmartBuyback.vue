<template>
  <div ref="smartBuyBack" :class="['smart-buy-back']">
    <PageLoading v-if="isPageLoading"/>
    <div v-show="!isPageLoading" class="smart-buy-back__inner">
      <div class="smart-buy-back__desktop">
        <div class="smart-buy-back__desktop_left">
          <SmartBuyBackTop
            v-model="selectedDex.label"
            v-model:search="searchToken"
            :is-edit-mode="isEditMod"
            :tokens="tokensList"
            :projects="projects"
            :errors="errors"
            :is-page-loading="isPageLoading"
            @handle-error-clear="handleGeneralErrors"
          />
          <Targets
            class="smart-buy-back__block transaction-setup"
            :jito-data="jitoData"
            :projects="projects"
            :errors="errors"
            :is-edit-mode="isEditMod"
            :estimate-error="estimateError"
            :solana-price-u-s-d="solanaPriceUSD"
            :is-page-loading="isPageLoading"
            @handle-error-clear="handleTargetErrorClear"
            @clear-estimate-error="clearEstimateError"
            @set-estimate-error="setEstimateError"
          />
          <UIAlert
            class="smart-buy-back__alert"
            text="The campaign will close automatically when the budget is spent or the target is reached."
            status="blue"
            :icon="SVGAlertInfo"
          />
          <div class="smart-buy-back__btns">
            <UIButton
              class="smart-buy-back__start"
              size="large"
              color_type="primary"
              @cta="handleStartCampaignClick"
              :is_disabled="(isEditMod && !smartCampaignStore.isCampaignDataChanged) || isChangesSaving || isPoolNotFound"
            >
              {{ startButtonText }}
            </UIButton>
            <UIButton
              v-if="isEditMod && smartCampaignStore.isCampaignDataChanged && !isChangesSaving"
              class="smart-buy-back__start"
              size="large"
              color_type="outline"
              @cta="handleCancelClick"
            >
              Cancel
            </UIButton>
          </div>
        </div>
        <div v-if="!isEditMod" class="smart-buy-back__desktop_right">
          <div class="smart-buy-back__active-campaigns">
            <UISectionTitleWithBorder>Active campaign status</UISectionTitleWithBorder>
            <div class="smart-buy-back__desktop_campaigns">
              <UIEmptyState
                v-if="!activeCampaign?.length"
                :icon="SVGMonitorDot"
                :main_text="'No active campaigns'"
                :add_text="'Active campaigns will show here'"
              />
              <div class="list" v-else>
                <ProfileCampaign
                  v-for="campaign in activeCampaign"
                  :key="campaign.id"
                  :campaign="campaign"
                  @handle-add-budget="openModal({type: 'add-budget', campaign})"
                  @handle-stop="openModal({type: 'stop-campaign', campaign})"
                  @handle-edit="openCampaign(campaign)"
                />
              </div>
            </div>
          </div>
          <div class="smart-buy-back__all-campaigns">
            <UISectionTitleWithBorder>All campaign status</UISectionTitleWithBorder>
            <div class="smart-buy-back__desktop_campaigns">
              <div class="list">
                <CompletedCampaign
                  v-for="campaign in smartCampaignStore.allSmartCampaigns"
                  :key="campaign.id"
                  :campaign="campaign"
                />
              </div>

              <router-link class="see-all paragraph-small medium" :to="{name: 'MarketHistory', query: {type: 'smart'}}">See all history</router-link>
            </div>
          </div>
        </div>
      </div>
      <MobileAdaptsNotification class="smart-buy-back__mobile"/>
    </div>

    <Modals>
      <template #title-icon-left>
        <component class="modal-title-icon" v-if="modalsStore.modalData.icon" :is="modalsStore.modalData.icon"/>
      </template>
      <template #default>
        <ConfirmationModal
          class="create-confirmation"
          v-if="modalsStore.modalData.type === 'create-confirmation'"
          :additional-text="`Your Smart Buy/Sell campaign has been created. \n The algorithm is now monitoring the market for the best execution.`"
          cancellation-btn-text="Ok"
          :is-confirmation-btn="false"
          header-color="success"
        />
        <ModalAddBudget
          v-if="modalsStore.modalData.type === 'add-budget'"
          v-model="modalAddNewBudget"
        />
        <ConfirmationModal
          class="modal-stop-campaign"
          v-if="modalsStore.modalData.type === 'stop-campaign'"
          header-color="error"
          :main-text="`You are about to stop the campaign`"
          additional-text="All running transactions will be halted. Remaining budget will stay in your project wallet."
          confirmation-btn-style="destructive"
          confirmation-btn-text="Stop campaign"
          @handle-confirmation="handleStopCampaign"
        />
      </template>
    </Modals>
  </div>
</template>
<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {
  CreateSmartBuyBack, CreateSmartBuyBackTarget,
  GetAllProjectsNameOnly, StopSmartBuyBackTarget, UpdateSmartBuyBackTarget,
} from "../../api/api.js";
import UIButton from "../../components/UI/UIButton.vue";
import UIAlert from "../../components/UI/UIAlert.vue";
import SVGAlertInfo from "../../components/SVG/SVGAlertInfo.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import ProfileCampaign from "../../components/Profile/ProfileCampaign.vue";
import {useRoute, useRouter} from "vue-router";
import UIEmptyState from "../../components/UI/UIEmptyState.vue";
import SVGMonitorDot from "../../components/SVG/SVGMonitorDot.vue";
import {errorToast, trackGoogleTagEvent} from "../../helpers/index.js";
import Modals from "../../components/UI/Modals.vue";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import ModalAddBudget from "../../components/MarketMakingPages/TargetPullUp/Modals/ModalAddBudget.vue";
import PageLoading from "../../components/UI/PageLoading.vue";
import {cloneDeep, debounce} from "lodash";
import {useToastStore} from "../../store/toastStore.js";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useTokensStore} from "../../store/tokensStore.js";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";
import Targets from "../../components/MarketMakingPages/SmartBuyBack/Targets.vue";
import SmartBuyBackTop from "../../components/MarketMakingPages/SmartBuyBack/SmartBuyBackTop.vue";
import {useSmartCampaignsStore} from "../../store/smartCampaignsStore.js";
import CompletedCampaign from "../../components/MarketMakingPages/CompletedCampaign.vue";
import {SOLANA_MINT} from "../../constants/const.js";

const DEFAULT_TARGET_ERROR = {
  id: 1,
  budget: '',
  max_time_between_transactions: '',
  max_transaction_amount: '',
  min_time_between_transactions: '',
  min_transaction_amount: '',
  slippage: '',
  start_at: '',
  target_price: '',
}

const route = useRoute();
const router = useRouter();
const toastStore = useToastStore();
const smartCampaignStore = useSmartCampaignsStore();
const tokensStore = useTokensStore();
const modalsStore = useModalsStore();
const userStore = useUserStore();
const smartBuyBack = ref(null);
const modalAddNewBudget = ref(0);
const isChangesSaving = ref(false);
const isPoolNotFound = ref(false);
const projects = ref([]);
const searchToken = ref('');
const estimateError = ref('');
const isPageLoading = ref(true)
const solanaPriceUSD = ref(null);
let campaignsInterval = null
const selectedDex = ref({label: 'Raydium', val: 'raydium'});
const tokensList = ref([]);
const errors = ref({
  token_mint: '',
  project_id: '',
  targets: [cloneDeep(DEFAULT_TARGET_ERROR)],
});
const jitoData = ref({
  default: 0.000005,
  fast: 0.00001,
  extra: 0.0002,
});

const isEditMod = computed(() => {
  return route.params.campaign_id !== 'create';
})

const startButtonText = computed(() => {
  if (isEditMod.value) {
    return isChangesSaving.value ? 'Saving...' : 'Save changes';
  } else {
    return isChangesSaving.value ? 'Creating...' : 'Start campaign';
  }
})
const activeCampaign = computed(() => {
  return smartCampaignStore.activeSmartCampaigns;
})

const getProjects = async () => {
  try {
    const resp = await GetAllProjectsNameOnly();
    projects.value = resp.data;
  } catch (e) {
    console.error(e)
    errorToast(e.response?.data)
  }
}

const searchSolToken = async () => {
  const url = new URL(`${import.meta.env.VITE_SOLSCAN_URL}/search`);

  url.search = new URLSearchParams({
    keyword: searchToken.value,
    search_by: "combination",
    sort_by: "reputation",
    search_mode: "exact",
    sort_order: "desc",
    page: "1",
    page_size: "10"
  }).toString();
  try {
    const response = await fetch(
      url.toString(),
      {
        method: "GET",
        headers: {
          token: import.meta.env.VITE_SOLSCAN_API_KEY
        }
      }
    );

    const tokensResp = await response.json();
    tokensList.value = tokensResp.data?.items || [];
  } catch (e) {
    console.error(e);
  }
}

const debouncedSearch = debounce((val) => {
  if (val) {
    searchSolToken()
  }
}, 400)

const openModal = ({type, campaign = null}) => {
  modalsStore.modalData.type = type;

  if (type === 'stop-campaign') {
    modalsStore.modalData.title = 'Stop active campaign?'
    modalsStore.modalData.action = 'confirmation'
  }

  if (type === 'create-confirmation') {
    modalsStore.modalData.title = `Smart Strategy Deployed`
    modalsStore.modalData.action = 'confirmation';
  }

  if (campaign) {
    modalsStore.modalData.item = campaign;
  }

  modalsStore.modalData.is_open = true;
}

const handlePageRefresh = async (isRefreshing = false, isAuth = false) => {
  isPageLoading.value = true;

  if (!isAuth) {
    smartCampaignStore.clearStore();
  }

  if (campaignsInterval) {
    clearInterval(campaignsInterval)
  }

  const solanaPrice = await tokensStore.getTokenPrice(SOLANA_MINT);

  if (solanaPrice?.[SOLANA_MINT]) {
    solanaPriceUSD.value = solanaPrice[SOLANA_MINT].usdPrice || 0;
  }

  if (!userStore.isUserAuth) {
    isPageLoading.value = false;

    return
  }

  try {
    if (isEditMod.value) {
      await smartCampaignStore.getSmartCampaign(route.params.campaign_id);
      const sourceToken = [{source_token_mint: smartCampaignStore.smartCampaignData.token_mint}];
      await tokensStore.updateSolTokensData(sourceToken, 'source_token_mint');
    } else {
      await smartCampaignStore.getAllActiveSmartCampaigns();
      await smartCampaignStore.getAllSmartCampaigns({page: 1, pageSize: 3});
    }
    await getProjects();

    if (isRefreshing) {
      toastStore.success({text: "Page is refreshed"})
    }
  } catch (e) {
    errorToast(e.response?.data)
  } finally {
    isPageLoading.value = false;
  }
}

const handleStopCampaign = async () => {
  try {
    await smartCampaignStore.handleStopSmartCampaign(modalsStore.modalData.item?.id);
    await smartCampaignStore.getAllActiveSmartCampaigns();
  } finally {
    modalsStore.closeModal();
  }
}


const openCampaign = (campaign) => {
  if (!campaign) return;

  router.push({name: route.name, params: {campaign_id: campaign.id}});
}

const handleGeneralErrors = (field) => {
  if (!field) return;

  errors.value[field] = '';
}

const handleTargetErrorClear = ({targetId = null, field = ''} = {}) => {
  if (!targetId || !field) return;

  const targetErrorIndex = errors.value.targets.findIndex((targetError) => targetError.id === targetId);
  if (targetErrorIndex === -1) return;
  if (!Object.hasOwn(errors.value.targets[targetErrorIndex], field)) return;

  errors.value.targets[targetErrorIndex][field] = '';
}
const setEstimateError = (error) => {
  if (!error || estimateError.value === error) return;
  estimateError.value = error;

  errorToast(error);
}
const clearEstimateError = () => {
  estimateError.value = '';
}
const buildTargetErrors = (targets = []) => {
  return (targets || []).map((target, index) => ({
    ...cloneDeep(DEFAULT_TARGET_ERROR),
    id: target?.id ?? index + 1,
  }));
}

const validateCampaignBeforeStart = () => {
  if (!userStore.isUserAuth) {
    userStore.isOpenLoginModal();

    return
  }


  if (estimateError.value) {
    errorToast(estimateError.value);

    return;
  }

  const campaign = smartCampaignStore.smartCampaignData || {};
  const targets = Array.isArray(campaign.targets) ? campaign.targets : [];
  const nextErrors = {
    token_mint: '',
    project_id: '',
    targets: buildTargetErrors(targets),
  };

  const tokenMint = campaign.token_mint;
  const projectId = campaign.project_id;

  if (!tokenMint) {
    nextErrors.token_mint = 'Token is required';
  }

  if (!projectId) {
    nextErrors.project_id = 'Project is required';
  }

  if (!targets.length) {
    errors.value = nextErrors;
    return false;
  }

  targets.forEach((target, index) => {
    const targetErrors = nextErrors.targets[index];

    const budget = Number(target?.budget);
    const maxTimeBetweenTransactions = Number(target?.max_time_between_transactions);
    const maxTransactionAmount = Number(target?.max_transaction_amount);
    const minTimeBetweenTransactions = Number(target?.min_time_between_transactions);
    const minTransactionAmount = Number(target?.min_transaction_amount);
    const slippage = Number(target?.slippage);
    const startAt = String(target?.start_at || '').trim();
    const targetPrice = Number(target?.target_price);

    if (!Number.isFinite(budget) || budget <= 0) {
      targetErrors.budget = 'Required';
    }

    if (!Number.isFinite(maxTimeBetweenTransactions) || maxTimeBetweenTransactions <= 0) {
      targetErrors.max_time_between_transactions = 'Required';
    }

    if (!Number.isFinite(minTimeBetweenTransactions) || minTimeBetweenTransactions <= 0) {
      targetErrors.min_time_between_transactions = 'Required';
    }

    if (!Number.isFinite(minTransactionAmount) || minTransactionAmount <= 0) {
      targetErrors.min_transaction_amount = 'Required';
    }

    if (!Number.isFinite(maxTransactionAmount) || maxTransactionAmount <= 0) {
      targetErrors.max_transaction_amount = 'Required';
    }

    if (
      Number.isFinite(minTransactionAmount) &&
      Number.isFinite(maxTransactionAmount) &&
      minTransactionAmount > maxTransactionAmount
    ) {
      targetErrors.min_transaction_amount = 'Must be less than or equal to max';
      targetErrors.max_transaction_amount = 'Must be greater than or equal to min';
    }

    if (Number.isFinite(budget) && Number.isFinite(minTransactionAmount) && minTransactionAmount > budget) {
      targetErrors.min_transaction_amount = 'Must be less than or equal to budget';
    }

    if (Number.isFinite(budget) && Number.isFinite(maxTransactionAmount) && maxTransactionAmount > budget) {
      targetErrors.max_transaction_amount = 'Must be less than or equal to budget';
    }

    if (!Number.isFinite(slippage) || slippage <= 0) {
      targetErrors.slippage = 'Required';
    }

    if (!startAt) {
      targetErrors.start_at = 'Required';
    }

    if (!Number.isFinite(targetPrice) || targetPrice <= 0) {
      targetErrors.target_price = 'Required';
    }
  });

  errors.value = nextErrors;

  return !Object.values(nextErrors).some((value) => {
    if (Array.isArray(value)) {
      return value.some((targetError) =>
        Object.entries(targetError).some(([key, error]) => key !== 'id' && Boolean(error))
      );
    }

    return Boolean(value);
  });
}

const SMART_BUYBACK_TARGET_COMPARE_KEYS = [
  'budget',
  'max_time_between_transactions',
  'max_transaction_amount',
  'min_time_between_transactions',
  'min_transaction_amount',
  'parallel_transactions_amount',
  'priority_fee',
  'slippage',
  'start_at',
  'target_price',
  'transaction_speed',
  'type',
  'using_jito',
];

const normalizeSmartBuyBackStartAtSeconds = (value) => {
  if (value === null || value === undefined || value === '') return 0;
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  const ms = Date.parse(value);

  return Number.isNaN(ms) ? 0 : Number((ms / 1000).toFixed());
};

const normalizeSmartBuyBackTargetFieldForCompare = (key, raw) => {
  if (key === 'start_at') return normalizeSmartBuyBackStartAtSeconds(raw);
  if (['budget', 'max_transaction_amount', 'min_transaction_amount', 'target_price', 'priority_fee'].includes(key)) {
    return String(Number(raw) || 0);
  }
  if (key === 'slippage') return Number(raw);
  if (['max_time_between_transactions', 'min_time_between_transactions', 'parallel_transactions_amount'].includes(key)) {
    return Number(raw);
  }
  if (key === 'using_jito') return Boolean(raw);

  return raw;
};

const isSmartBuyBackTargetChanged = (oldTarget, newTarget) => {
  return SMART_BUYBACK_TARGET_COMPARE_KEYS.some(
    (key) =>
      normalizeSmartBuyBackTargetFieldForCompare(key, oldTarget?.[key])
      !== normalizeSmartBuyBackTargetFieldForCompare(key, newTarget?.[key])
  );
};

const handleTargetsCheck = async(newCampaign) => {
  if (!newCampaign) return;

  for (const oldTarget of smartCampaignStore.smartCampaignDataBeforeChange.targets) {
    const newTarget = newCampaign.targets.find((t) => t.id === oldTarget.id);

    if (newTarget) {
      if (isSmartBuyBackTargetChanged(oldTarget, newTarget)) {
        console.log(newTarget)
        await UpdateSmartBuyBackTarget({id: newCampaign.id, targetID: newTarget.id, data: newTarget});
      }
    } else {
      await StopSmartBuyBackTarget({id: newCampaign.id, targetID: oldTarget.id});
    }
  }

  for (const newTargetId of smartCampaignStore.currentTargetIds) {
    const newTarget = newCampaign.targets.find((t) => t.id === newTargetId);

    if (!smartCampaignStore.oldTargetIds.includes(newTargetId) && newTarget) {
      await CreateSmartBuyBackTarget(newCampaign.id, newTarget);
    }
  }
}
const runStartCampaign = async() => {
  if (userStore.isOpenLoginModal()) return;

  const campaign = cloneDeep(smartCampaignStore.smartCampaignData);
  campaign.targets.forEach(target => {
    if (target.start_at) {
      const millisec = Date.parse(target.start_at);
      const secs = (millisec / 1000).toFixed();
      target.start_at = Number(secs);
    } else {
      target.start_at = 0;
    }

    target.budget = String(target.budget);
    target.max_transaction_amount = String(target.max_transaction_amount);
    target.min_transaction_amount = String(target.min_transaction_amount);
    target.target_price = String(target.target_price);
    target.priority_fee = String(target.priority_fee);
    target.slippage = Number(target.slippage);
  })
  try {
    isChangesSaving.value = true;

    if (isEditMod.value) {
      await handleTargetsCheck(campaign);

      toastStore.success({text: 'Campaign has been updated.'});
    } else {
      await CreateSmartBuyBack(campaign);
      smartCampaignStore.clearStore();

      await smartCampaignStore.getAllActiveSmartCampaigns();
      await smartCampaignStore.getAllSmartCampaigns();

      openModal({type: 'create-confirmation'});
    }

    trackGoogleTagEvent('Start campaign Buyback');

    const scrollContainer = smartBuyBack.value?.parentElement;
    if (scrollContainer) {
      scrollContainer.scrollTo({top: 0, behavior: 'smooth'});
    }
  } catch (e) {
    console.error(e.response.data);
    errorToast(e.response.data);
  } finally {
    isChangesSaving.value = false;
  }
}

const handleStartCampaignClick = async () => {
  const isValid = validateCampaignBeforeStart();
  if (!isValid) return;

  await runStartCampaign();
}

const handleCancelClick = () => {
  smartCampaignStore.returnCampaignChanged();
}
useHeaderRefresh(() => handlePageRefresh(true));

watch(() => searchToken.value, (newVal) => {
  debouncedSearch(newVal)
})

watch(
  () => (smartCampaignStore.smartCampaignData.targets || []).map((target) => target.id).join(','),
  () => {
    const campaignTargets = smartCampaignStore.smartCampaignData.targets || [];
    const existingErrorsById = new Map((errors.value.targets || []).map((targetError) => [targetError.id, targetError]));

    errors.value.targets = campaignTargets.map((target, index) => ({
      ...cloneDeep(DEFAULT_TARGET_ERROR),
      ...existingErrorsById.get(target.id),
      id: target?.id ?? index + 1,
    }));
  },
  {immediate: true}
)

watch(() => modalsStore.modalData.is_open, (newVal) => {
  if (!newVal) {
    modalAddNewBudget.value = 0;
  }
});

watch(() => [route.name, route.params.campaign_id], async () => {
  smartCampaignStore.clearStore();

  await handlePageRefresh();
})

watch(() => userStore.isUserAuth, async (newVal) => {
  if (newVal) {
    await handlePageRefresh(false, true);
  }
})

onMounted(async () => {
  if (route.params?.campaign_id !== 'create' && !userStore.isUserAuth) {
    await router.push({params: {campaign_id: 'create'}});
  }

  await handlePageRefresh();
});
onBeforeUnmount(() => {
  smartCampaignStore.clearStore();
})
</script>
<style scoped lang="scss">
.smart-buy-back {
  &__block {
    margin-top: 32px;
  }

  &__btns {
    display: flex;
    gap: 12px;

  }

  &__back {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #374151;
    cursor: pointer;
  }

  &__start {
    margin: 32px 0;
  }

  &__estimate {
    margin-top: 32px;
  }

  &__desktop {
    width: 100%;
    display: flex;
    gap: 10px;
    justify-content: space-between;

    &_left {
      width: 100%;
      max-width: 669px;
    }

    &_right {
      width: 100%;
      max-width: 371px;
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    &_campaigns {
      margin-top: 20px;
      display: flex;
      flex-direction: column;

      & .list {
        display: flex;
        flex-direction: column;
        gap: 12px;
      }

      & .see-all {
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 30px auto 0;
        color: #374151;
      }
    }
  }

  &__mobile {
    display: none;
  }

  &.pull-down {
    & .smart-buy-back {
      &__block {
        &.transaction-setup, &.target {
          opacity: .5;
          pointer-events: none;
        }
      }
    }
  }
}

.modal-stop-campaign {
  max-width: 480px;
}

@media (max-width: 1200px) {
  .smart-buy-back {
    &__desktop {
      display: none;
    }

    &__mobile {
      margin: 16px auto;
      padding: 0 16px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>