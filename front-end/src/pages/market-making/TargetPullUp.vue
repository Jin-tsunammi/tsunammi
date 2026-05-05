<template>
  <div ref="targetPullUpRef" :class="['target-pull-up', {'pull-down': isTokenNotSelected}]">
    <PageLoading v-if="isPageLoading"/>
    <div v-show="!isPageLoading" class="target-pull-up__inner">
      <div class="target-pull-up__desktop">
        <div class="target-pull-up__desktop_left">
          <ExchangeSettings
            ref="ExchangeSettingsRef"
            v-model="selectedDex.label"
            v-model:search="searchToken"
            :is-edit-mode="isEditMod"
            :tokens="tokensList"
            :errors="errors"
            :campaign-action="campaignAction"
          />
          <TransactionSetup
            class="target-pull-up__block transaction-setup"
            :jito-data="jitoData"
            :projects="projects"
            :errors="errors"
            :is-edit-mode="isEditMod"
            :campaign-action="campaignAction"
            :priority-fees="priorityFees"
            :is-route-changed="isRouteChanged"
          />
          <TargetsAndStopTriggers
            class="target-pull-up__block target"
            :campaign-action="campaignAction"
            :errors="errors"
          />
          <CampaignEstimate
            :estimate-data="campaignEstimate"
            class="target-pull-up__estimate"
            :campaign-action="campaignAction"
          />
          <div class="target-pull-up__btns">
            <UIButton
              class="target-pull-up__start"
              size="large"
              color_type="primary"
              @cta="handleStartCampaignClick"
              :is_disabled="(isEditMod && !campaignStore.isCampaignDataChanged) || isChangesSaving || isPoolNotFound"
            >
              {{ startButtonText }}
            </UIButton>
            <UIButton
              v-if="isEditMod && campaignStore.isCampaignDataChanged && !isChangesSaving"
              class="target-pull-up__start"
              size="large"
              color_type="outline"
              @cta="handleCancelClick"
            >
              Cancel
            </UIButton>
          </div>
          <UIAlert
            class="target-pull-up__alert"
            text="The campaign will close automatically when the budget is spent or the target is reached."
            status="blue"
            :icon="SVGAlertInfo"
          />
        </div>
        <div v-if="!isEditMod" class="target-pull-up__desktop_right">
          <div class="target-pull-up__desktop_active">
            <UISectionTitleWithBorder>Active campaign status</UISectionTitleWithBorder>
            <div class="target-pull-up__desktop_campaigns">
              <UIEmptyState
                v-if="!campaignStore.activeCampaigns?.length"
                :icon="SVGMonitorDot"
                :main_text="'No active campaigns'"
                :add_text="'Active campaigns will show here'"
              />
              <div class="list" v-else>
                <ProfileCampaign
                  v-for="campaign in campaignStore.activeCampaigns"
                  :key="campaign.campaign_id"
                  :campaign="campaign"
                  :campaign-action="campaignAction"
                  @handle-add-budget="openModal({type: 'add-budget', campaign})"
                  @handle-stop="openModal({type: 'stop-campaign', campaign})"
                  @handle-edit="openCampaign(campaign)"
                />
              </div>
            </div>
          </div>
          <div v-if="campaignsHistory.length" class="target-pull-up__desktop_history">
            <UISectionTitleWithBorder>History</UISectionTitleWithBorder>
            <div class="target-pull-up__desktop_campaigns">
              <div class="list">
                <CompletedCampaign
                  v-for="campaign in campaignsHistory"
                  :key="campaign.campaign_id"
                  :campaign="campaign"
                  :campaign-action="campaignAction"
                />
              </div>
            </div>

            <router-link class="see-all paragraph-small medium" :to="{name: 'MarketHistory'}">See all history</router-link>
          </div>
        </div>
      </div>
      <MobileAdaptsNotification class="target-pull-up__mobile"/>
    </div>

    <Modals>
      <template #title-icon-left>
        <component class="modal-title-icon" v-if="modalsStore.modalData.icon" :is="modalsStore.modalData.icon"/>
      </template>
      <template #default>
        <ModalAddBudget
          v-if="modalsStore.modalData.type === 'add-budget'"
          v-model="modalAddNewBudget"
          :campaign-action="campaignAction"
        />
        <ConfirmationModal
          v-if="modalsStore.modalData.type === 'budget-confirmation'"
          :main-text="`You’re adding ${modalAddNewBudget} ${tokensStore.solTokensData?.[modalsStore.modalData.item?.token_mint_from]?.symbol || ''} to this campaign. This action can’t be undone.`"
          confirmation-btn-style="primary"
          confirmation-btn-text="Confirm"
          @handle-confirmation="handleCampaignChangeBudget"
        />
        <ConfirmationModal
          class="modal-stop-campaign"
          v-if="modalsStore.modalData.type === 'stop-campaign'"
          :is-custom-content="true"
        >
          <template #confirmation-custom-content>
            <ModalStopCampaign @handle-stop-campaign="handleStopCampaign"/>
          </template>
        </ConfirmationModal>
      </template>
    </Modals>
  </div>
</template>
<script setup>
import TargetsAndStopTriggers from "../../components/MarketMakingPages/TargetPullUp/TargetsAndStopTriggers.vue";
import ExchangeSettings from "../../components/MarketMakingPages/TargetPullUp/ExchangeSettings.vue";
import TransactionSetup from "../../components/MarketMakingPages/TargetPullUp/TransactionSetup.vue";
import {computed, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {
  CreatePumpFunPullDown, CreatePumpFunPullUp,
  CreateRaydiumPullDown,
  CreateRaydiumPullUp,
  GetAllProjectsWithBalance,
  GetJitoInfo, GetProjectWithBalance, GetPumpFunEstimate,
  GetRaydiumEstimate,
  UpdateCampaign
} from "../../api/api.js";
import UIButton from "../../components/UI/UIButton.vue";
import UIAlert from "../../components/UI/UIAlert.vue";
import SVGAlertInfo from "../../components/SVG/SVGAlertInfo.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import ProfileCampaign from "../../components/Profile/ProfileCampaign.vue";
import {useRoute, useRouter} from "vue-router";
import UIEmptyState from "../../components/UI/UIEmptyState.vue";
import SVGMonitorDot from "../../components/SVG/SVGMonitorDot.vue";
import CampaignEstimate from "../../components/MarketMakingPages/TargetPullUp/CampaignEstimate.vue";
import {useCampaignsStore} from "../../store/campaignsStore.js";
import {calculateBudget, errorToast} from "../../helpers/index.js";
import Modals from "../../components/UI/Modals.vue";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import ModalAddBudget from "../../components/MarketMakingPages/TargetPullUp/Modals/ModalAddBudget.vue";
import ModalStopCampaign from "../../components/MarketMakingPages/TargetPullUp/Modals/ModalStopCampaign.vue";
import PageLoading from "../../components/UI/PageLoading.vue";
import {cloneDeep, debounce} from "lodash";
import {useToastStore} from "../../store/toastStore.js";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useTokensStore} from "../../store/tokensStore.js";
import {storeToRefs} from "pinia";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";
import CompletedCampaign from "../../components/MarketMakingPages/CompletedCampaign.vue";

const route = useRoute();
const router = useRouter();
const toastStore = useToastStore();
const campaignStore = useCampaignsStore();
const tokensStore = useTokensStore();
const modalsStore = useModalsStore();
const userStore = useUserStore();
const {campaignDataBeforeChange} = storeToRefs(campaignStore);
const targetPullUpRef = ref(null);
const ExchangeSettingsRef = ref(null);
const SOLANA_MINT = 'So11111111111111111111111111111111111111112';
const modalAddNewBudget = ref(0);
const campaignEstimate = ref(null);
const isChangesSaving = ref(false);
const isPoolNotFound = ref(false);
const isRouteChanged = ref(false);
const projects = ref([]);
const campaignsHistory = ref([]);
const searchToken = ref('');
const isPageLoading = ref(true);
let jitoInterval = null
let campaignsInterval = null
const selectedDex = ref({label: 'Raydium', val: 'raydium'});
const tokensList = ref([]);
const errors = ref({
  dest_token_mint: '',
  project_id: '',
  budget: '',
  slippage: '',
  goal_percentage_change: '',
  parallel_transactions_amount: '',
  min_transactions_budget: '',
  max_transactions_budget: '',
  min_time_between_transactions: '',
  max_time_between_transactions: '',
  transaction_speed: '',
});
const jitoData = ref({
  default: 0.000005,
  fast: 0.00001,
  extra: 0.0002,
});
const priorityFees = ref(null);
const isEditMod = computed(() => {
  return route.params.campaign_id !== 'create';
})
const campaignAction = computed(() => {
  if (route.name === 'MarketTargetDrop') {
    return 'pull-down'
  } else {
    return 'pull-up'
  }
})
const activeCampaignsParams = computed(() => {
  return {
    status: 'ACTIVE',
    type: campaignAction.value.replace('-', '_'),
  };
})
const isTokenNotSelected = computed(() => {
  return campaignAction.value === 'pull-down' && !campaignStore.campaign.source_token_mint;
})
const tokenMint = computed(() => {
  if (campaignAction.value === 'pull-up') {
    return 'dest_token_mint'
  } else {
    return 'source_token_mint'
  }
})
const startButtonText = computed(() => {
  if (isEditMod.value) {
    return isChangesSaving.value ? 'Saving...' : 'Save changes';
  } else {
    return isChangesSaving.value ? 'Creating...' : 'Start campaign';
  }
})
const getJitoData = async () => {
  try {
    const resp = await GetJitoInfo();
    jitoData.value = resp.data;

  } catch (error) {
    console.error(error);
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
  } else {
    tokensList.value = [];
  }
}, 400)

const getProject = async(campaign) => {
  let resp = null;
  try {
    resp = await GetProjectWithBalance(campaign.project_id, {mint: campaign.token_mint_from});
    modalsStore.modalData.project = resp.data;
    return resp.data;
  } catch (e) {
    errorToast(e.response.data);
    return resp;
  }
}

const openModal = async({type, campaign = null}) => {
  modalsStore.modalData.type = type;

  if (type === 'add-budget') {
    modalsStore.modalData.title = 'Add budget to campaign';
    const project = await getProject(campaign);

    if (!project) {
      modalsStore.modalData.type = '';
      modalsStore.modalData.title = '';
      return;
    }
  }

  if (type === 'stop-campaign') {
    modalsStore.modalData.title = 'Stop active campaign?'
    modalsStore.modalData.action = 'confirmation'
  }

  if (campaign) {
    modalsStore.modalData.item = campaign;
  }

  modalsStore.modalData.is_open = true;
}

const getProjectsWithBalance = async () => {
  let params = null;

  if (campaignAction.value === 'pull-up') {
    params = {mint: SOLANA_MINT};
  } else {
    params = {mint: campaignStore.campaign[tokenMint.value]};
  }

  try {
    if (!params?.mint) return;

    const projectsResp = await GetAllProjectsWithBalance(params);

    projects.value = projectsResp.data?.projects;
  } catch (e) {
    errorToast(e.response.data);
  }
}

const handlePageRefresh = async (isRefreshing = false, isAuth=false) => {
  isPageLoading.value = true;

  if (!isAuth) {
    campaignStore.clearStore();
  }

  if (jitoInterval) {
    clearInterval(jitoInterval)
  }

  if (campaignsInterval) {
    clearInterval(campaignsInterval)
  }

  if (!userStore.isUserAuth) {
    isPageLoading.value = false;

    return
  }

  try {
    await getJitoData();

    if (isEditMod.value) {
      await campaignStore.getCampaign(route.params.campaign_id);
      const sourceToken = [{source_token_mint: campaignStore.campaign[tokenMint.value]}];
      await tokensStore.updateSolTokensData(sourceToken, 'source_token_mint');
    } else {
      await campaignStore.getAllActiveCampaigns(activeCampaignsParams.value);
      await campaignStore.getAllCampaigns({
        page: 1,
        pageSize: 30,
      });

      if (campaignStore.allCampaigns.length) {
        const onlyCompleted = campaignStore.allCampaigns?.filter(campaign => campaign.status.toLowerCase() !== 'active')
        campaignsHistory.value = onlyCompleted.slice(0, 3);
      }
    }

    await getProjectsWithBalance();

    jitoInterval = setInterval(() => {
      getJitoData()
    }, 10000)

    campaignsInterval = setInterval(() => {
      campaignStore.getAllActiveCampaigns(activeCampaignsParams.value);
    }, 10000)

    if (isRefreshing) {
      toastStore.success({text: "Page is refreshed"})
    }
  } catch (e) {
    errorToast(e.response.data)
  } finally {
    isPageLoading.value = false;
  }
}

const handleStopCampaign = async () => {
  try {
    await campaignStore.handleStopCampaign(modalsStore.modalData.item?.campaign_id, activeCampaignsParams.value);
  } finally {
    modalsStore.closeModal();
  }
}

const clearError = (field) => {
  if (!field) return;
  if (Object.prototype.hasOwnProperty.call(errors.value, field)) {
    errors.value[field] = '';
  }
}

const openCampaign = async (campaign) => {
  if (!campaign) return;

  await router.push({name: route.name, params: {campaign_id: campaign.campaign_id}});
}

const handleCampaignChangeBudget = async () => {
  if (!modalsStore.modalData?.item?.campaign_id) return;

  const newBudget = modalsStore.modalData.item.budget + modalAddNewBudget.value;

  try {
    await UpdateCampaign(modalsStore.modalData.item.campaign_id, {budget: newBudget});
    await campaignStore.getAllActiveCampaigns(activeCampaignsParams.value);
    modalsStore.closeModal();
    toastStore.success({text: 'Changes have been saved.'});
  } catch (e) {
    console.error(e.response.data);
    errorToast(e.response.data);
  }
}

const fetchEstimate = debounce(async (data) => {
  if (Object.values(data).some(v => !v)) {
    campaignEstimate.value = null
    return
  }
  try {
    let resp = null;
    if (ExchangeSettingsRef.value && ExchangeSettingsRef.value.selectedDex?.val === 'pumpfun') {
      resp = await GetPumpFunEstimate(data)
    } else {
      resp = await GetRaydiumEstimate(data)
    }
    campaignEstimate.value = resp.data
    priorityFees.value = resp.data.priority_fees || null;
    isPoolNotFound.value = false;
  } catch (e) {
    console.error(e)
    errorToast(e.response?.data);

    if (e.response?.data === 'pool not found') {
      isPoolNotFound.value = true;
    }

    campaignEstimate.value = null
  }
}, 400)

const validateCampaignBeforeStart = () => {
  const campaign = campaignStore.campaign || {};
  const NANO_IN_SECOND = 1_000_000_000;

  const destTokenMint = String(campaign[tokenMint.value] || '').trim();
  const projectId = Number(campaign.project_id);
  const budget = Number(campaign.budget || campaign.budget_percent);
  const slippage = Number(campaign.slippage);
  const goalPercentageChange = Number(campaign.goal_percentage_change);
  const parallelTransactionsAmount = Number(campaign.parallel_transactions_amount);
  const minTransactionsBudget = Number(campaign.min_transactions_budget);
  const maxTransactionsBudget = Number(campaign.max_transactions_budget);
  const minTimeBetweenTransactionsNs = Number(campaign.min_time_between_transactions);
  const maxTimeBetweenTransactionsNs = Number(campaign.max_time_between_transactions);
  const transactionSpeed = String(campaign.transaction_speed || '').trim();

  const nextErrors = {
    dest_token_mint: '',
    project_id: '',
    budget: '',
    slippage: '',
    goal_percentage_change: '',
    parallel_transactions_amount: '',
    min_transactions_budget: '',
    max_transactions_budget: '',
    min_time_between_transactions: '',
    max_time_between_transactions: '',
    transaction_speed: '',
  };

  if (!destTokenMint) {
    nextErrors.dest_token_mint = 'Token is required';

    if (campaignAction.value === 'pull-down') {
      errors.value = nextErrors;
      return false;
    }
  }

  if (!Number.isFinite(projectId) || !Number.isInteger(projectId) || projectId <= 0) {
    nextErrors.project_id = 'Project is required';
  }

  if (!Number.isFinite(budget) || budget <= 0) {
    nextErrors.budget = 'Budget must be greater than 0';
  }

  if (!Number.isFinite(slippage) || slippage <= 0 || slippage > 100) {
    nextErrors.slippage = 'Slippage must be between 0 and 100';
  }

  if (
    !Number.isFinite(goalPercentageChange) ||
    goalPercentageChange === 0 ||
    goalPercentageChange > 100
  ) {
    nextErrors.goal_percentage_change = 'Target must be greater than 0 and less than 100';
  }

  if (
    !Number.isFinite(parallelTransactionsAmount) ||
    !Number.isInteger(parallelTransactionsAmount) ||
    parallelTransactionsAmount < 1
  ) {
    nextErrors.parallel_transactions_amount = 'Must be at least 1';
  }

  if (!Number.isFinite(minTransactionsBudget) || minTransactionsBudget <= 0) {
    nextErrors.min_transactions_budget = 'Min transaction budget must be greater than 0';
  }

  if (!Number.isFinite(maxTransactionsBudget) || maxTransactionsBudget <= 0) {
    nextErrors.max_transactions_budget = 'Max transaction budget must be greater than 0';
  }

  if (
    Number.isFinite(minTransactionsBudget) &&
    Number.isFinite(maxTransactionsBudget) &&
    minTransactionsBudget > maxTransactionsBudget
  ) {
    nextErrors.min_transactions_budget = 'Must be less than or equal to max';
    nextErrors.max_transactions_budget = 'Must be greater than or equal to min';
  }

  if (
    Number.isFinite(budget) &&
    Number.isFinite(maxTransactionsBudget) &&
    maxTransactionsBudget > budget
  ) {
    nextErrors.max_transactions_budget = 'Must be less than or equal to total budget';
  }

  if (
    Number.isFinite(minTimeBetweenTransactionsNs) &&
    Number.isFinite(maxTimeBetweenTransactionsNs) &&
    minTimeBetweenTransactionsNs > maxTimeBetweenTransactionsNs
  ) {
    nextErrors.min_time_between_transactions = 'Must be less than or equal to max';
    nextErrors.max_time_between_transactions = 'Must be greater than or equal to min';
  }

  if (!['default', 'fast', 'extra'].includes(transactionSpeed)) {
    nextErrors.transaction_speed = 'Transaction speed is required';
  }

  errors.value = nextErrors;
  return Object.values(nextErrors).every(v => !v);
}

const runStartCampaign = async () => {
  if (userStore.isOpenLoginModal()) return;

  const campaign = cloneDeep(campaignStore.campaign);

  if (campaignAction.value === 'pull-up') {
    delete campaign.source_token_mint;
  } else {
    delete campaign.dest_token_mint;
  }

  if (!campaign.budget) delete campaign.budget;
  else if (!campaign.budget_percent) delete campaign.budget_percent;

  const fieldToNumb = ['budget', 'goal_percentage_change', 'max_transactions_budget', 'min_transactions_budget', 'slippage'];

  fieldToNumb.forEach(k => campaign[k] = Number(campaign[k]));

  try {
    if (isEditMod.value) {
      isChangesSaving.value = true;
      await UpdateCampaign(route.params.campaign_id, campaign);
      campaignDataBeforeChange.value = cloneDeep(campaignStore.campaign);
      toastStore.success({text: 'Changes have been saved.'});
    } else {
      isChangesSaving.value = true;
      const isPumpFun = ExchangeSettingsRef.value && ExchangeSettingsRef.value.selectedDex?.val === 'pumpfun';

      if (campaignAction.value === 'pull-up') {
        if (isPumpFun) {
          await CreatePumpFunPullUp(campaign);
        } else {
          await CreateRaydiumPullUp(campaign);
        }
      } else {
        if (isPumpFun) {
          await CreatePumpFunPullDown(campaign);
        } else {
          await CreateRaydiumPullDown(campaign);
        }
      }
      campaignEstimate.value = null;
      campaignStore.clearStore();
      await campaignStore.getAllActiveCampaigns(activeCampaignsParams.value);
      toastStore.success({text: 'Campaign has been created.'});
    }

    const scrollContainer = targetPullUpRef.value?.parentElement;
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
  campaignStore.returnCampaignChanged();
}
useHeaderRefresh(() => handlePageRefresh(true));
watch(() => ({
    dest_token_mint: campaignStore.campaign.dest_token_mint,
    project_id: campaignStore.campaign.project_id,
    budget: campaignStore.campaign.budget,
    slippage: campaignStore.campaign.slippage,
    goal_percentage_change: campaignStore.campaign.goal_percentage_change,
    parallel_transactions_amount: campaignStore.campaign.parallel_transactions_amount,
    min_transactions_budget: campaignStore.campaign.min_transactions_budget,
    max_transactions_budget: campaignStore.campaign.max_transactions_budget,
    min_time_between_transactions: campaignStore.campaign.min_time_between_transactions,
    max_time_between_transactions: campaignStore.campaign.max_time_between_transactions,
    transaction_speed: campaignStore.campaign.transaction_speed,
  }), (next, prev) => {
    if (!prev) return;

    for (const key of Object.keys(next)) {
      if (next[key] !== prev[key]) {
        clearError(key);
      }
    }
  },
  {deep: true}
)

watch(() => searchToken.value, (newVal) => {
  debouncedSearch(newVal)
})

watch(() => ({
    budget: +campaignStore.campaign.budget,
    source_token_mint: campaignAction.value === 'pull-up' ? SOLANA_MINT : campaignStore.campaign[tokenMint.value],
    dest_token_mint: campaignAction.value === 'pull-up' ? campaignStore.campaign[tokenMint.value] : SOLANA_MINT,
    project_id: campaignStore.campaign.project_id,
    slippage: +campaignStore.campaign.slippage,
    transaction_speed: campaignStore.campaign.transaction_speed,
    dex: ExchangeSettingsRef.value?.selectedDex || {},
  }),
  (data) => {
    if (!data.budget || data.source_token_mint === data.dest_token_mint) return;
    delete data.dex;

    fetchEstimate(data);
  },
  {deep: true, immediate: true}
)

watch(() => campaignStore.campaign[tokenMint.value], async (newVal) => {
  if (newVal && campaignAction.value === 'pull-down') {
    await getProjectsWithBalance();
  }
})

watch(() => modalsStore.modalData.is_open, (newVal) => {
  if (!newVal) {
    modalAddNewBudget.value = 0;
  }
});
watch(() => [route.name, route.params.campaign_id], async () => {
  isRouteChanged.value = true;
  campaignStore.clearStore();
  campaignEstimate.value = null;

  await handlePageRefresh();

  setTimeout(() => {
    isRouteChanged.value = false;
  }, 1000)
})

watch(() => userStore.isUserAuth, async(newVal) => {
  if (newVal) {
    await handlePageRefresh(false, true);
  }
})

onMounted(async () => {
  if (route.params?.campaign_id !== 'create' && !userStore.isUserAuth) {
    await router.push({ params: { campaign_id: 'create' } });
  }
  await handlePageRefresh();
});
onBeforeUnmount(() => {
  clearInterval(jitoInterval);
  clearInterval(campaignsInterval);
  fetchEstimate.cancel();
  campaignStore.clearStore();
})
</script>
<style scoped lang="scss">
.target-pull-up {
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
      max-width: 669px;
    }

    &_history {
      & .see-all {
        display: flex;
        align-items: center;
        justify-content: center;
        margin: 30px auto 0;
        color: #374151;
      }
    }

    &_right {
      width: 100%;
      max-width: 371px;
      display: flex;
      flex-direction: column;
      gap: 24px;
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
    }
  }

  &__mobile {
    display: none;
  }

  &.pull-down {
    & .target-pull-up {
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
  .target-pull-up {
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