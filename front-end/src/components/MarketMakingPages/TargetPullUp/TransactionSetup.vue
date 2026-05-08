<template>
  <div class="transaction-setup">
    <div class="transaction-setup__inner">
      <UISectionTitleWithBorder>Transaction setup</UISectionTitleWithBorder>

      <div class="transaction-setup__inputs">
        <UISelect
          class="transaction-setup__input"
          v-model="isProjectOptionsOpen"
          :selected="selectedProject?.name || ''"
          label="Select Wallet Pool"
          placeholder="Select a project"
          size="large"
          :error-message="errors?.project_id || ''"
          :is_disabled="isEditMode"
        >
          <template v-if="selectedProject" #right-icon>
            <span class="monospaced-small grey">{{ `${projectBalance(selectedProject)}` }}</span>
          </template>
          <template #dropdown>
            <UIDropdown
              :is-open="isProjectOptionsOpen"
              :selected-option="selectedProject?.name || ''"
              :options="projects"
              label="name"
              @handle-option-select="handleProjectSelect"
            >
              <template #custom-option="{item}">
                <div class="custom-option paragraph-small regular">
                  {{ item?.name || '' }}
                  <span class="monospaced-small grey">{{ projectBalance(item) }}</span>
                </div>
              </template>
              <template #custom-dropdown>
                <div class="custom-dropdown-content" @click.stop>
                  <div class="paragraph-small bold">You do not have wallet pools yet.</div>
                  <p class="paragraph-mini regular">Create wallet pool to start funding.</p>
                  <UIButton
                    color_type="ghost"
                    size="large"
                    @cta="openPage('project')"
                  >
                    Create wallet pool

                    <template #left-icon>
                      <SVGPlus/>
                    </template>
                  </UIButton>
                </div>
              </template>
            </UIDropdown>
          </template>
        </UISelect>
        <UIBaseInput
          class="transaction-setup__input"
          :label="`Set budget for Price ${campaignAction === 'pull-up' ? 'Boost' : 'Drop'}`"
          type="number"
          size="large"
          v-model="campaignStore.campaign.budget"
          :error-message="errors?.budget || ''"
          @handle-input="handleInput($event, 'budget')"
          @handle-blur="handleBlur($event, 'budget')"
        >
          <template #icon-right>
            <UIGhostSelector
              :selected-option="selectedAmountType.budget"
              :options="amountTypes"
              @handle-option-select="handlePriceTypeSelect($event, 'budget')"
            />
          </template>
          <template #bottom-right>
            <UIGhostButtonsGroup
              :options="budgetOptions"
              @handle-option-select="handleMaxBudget"
            />
          </template>
        </UIBaseInput>

        <div class="transaction-setup__slippage transaction-setup__input">
          <div class="transaction-setup__slippage_top">
            <div class="paragraph-small medium transaction-setup__input_label">Slippage Settings</div>
            <UIRoundToggle label="Custom" v-model:is-active="isSlippageCustom"/>
          </div>
          <UIToggleGroup
            v-if="!isSlippageCustom"
            :options="slippageOptions"
            label="label"
            size="large"
            @handle-option-select="handleSlippageSelect"
            :selected-option="String(campaignStore.campaign.slippage)"
          />
          <UIBaseInput
            v-else
            size="large"
            type="number"
            placeholder="Custom"
            v-model="campaignStore.campaign.slippage"
            :error-message="errors?.slippage || ''"
            @handle-input="handleInput($event, 'slippage')"
            @handle-blur="handleBlur($event, 'slippage')"
          >
            <template #icon-right><span>%</span></template>
          </UIBaseInput>
        </div>
        <div class="transaction-setup__speed transaction-setup__input">
          <div class="paragraph-small medium transaction-setup__input_label">Priority Fee</div>
          <UITabs size="large">
            <UITab
              v-for="k in validatorTipOptions"
              :key="k"
              :is_active="campaignStore.campaign.priority_fee === priorityFees?.[k]"
              @click="setPriorityFee(k)"
            >
              <template #default>
                <span class="paragraph-mini medium">{{ k }}</span>
              </template>
              <template #add-info>
                <span class="jito-data">{{ priorityFees?.[k] ? toDynamicFix(priorityFees?.[k]) : '--' }}</span>
              </template>
            </UITab>
          </UITabs>
          <span v-if="errors?.transaction_speed"
                class="paragraph-mini transaction-setup__error">{{ errors.transaction_speed }}</span>
        </div>

        <UIBaseInput
          v-for="timeEl in timeArray"
          :key="timeEl.type"
          class="transaction-setup__input"
          :label="timeEl.label"
          size="large"
          placeholder="sec"
          type="number"
          :top-add-text="timeEl.add_text"
          :is_dot_allowed="false"
          v-model="selectedTime[timeEl.type]"
          :error-message="errors?.[timeEl.type] || ''"
          @handle-input="handleInput($event, timeEl.type)"
          @handle-blur="handleBlur($event, timeEl.type)"
        >
          <template #icon-right>
            <div class="transaction-setup__input_controls">
              <UINumberIncreaseDecrease
                @handle-increase="handleTimeChangeByStep({action: 'increase', type: timeEl.type})"
                @handle-decrease="handleTimeChangeByStep({action: 'decrease', type: timeEl.type})"
              />
            </div>
          </template>
          <template #bottom-right>
            <UIGhostButtonsGroup
              :options="timeOptions"
              :selected-option="String(selectedTime[timeEl.type])"
              @handle-option-select="handleTimeSelect({type: timeEl.type, val: $event})"
            />
          </template>
        </UIBaseInput>

        <UIBaseInput
          label="Transactions Amount Range"
          class="transaction-setup__input budget-min"
          size="large"
          top-add-text="Min"
          type="number"
          placeholder="min budget range"
          v-model="campaignStore.campaign.min_transactions_budget"
          :error-message="errors?.min_transactions_budget || ''"
          @handle-input="handleInput($event, 'min_transactions_budget')"
          @handle-blur="handleBlur($event, 'min_transactions_budget')"
        >
          <template #icon-right>
            <UIGhostSelector
              :selected-option="selectedAmountType.min_amount"
              :options="amountTypes"
              @handle-option-select="handlePriceTypeSelect($event, 'min_amount')"
            />
          </template>
        </UIBaseInput>

        <UIBaseInput
          class="transaction-setup__input budget-max"
          size="large"
          top-add-text="Max"
          type="number"
          placeholder="max budget range"
          v-model="campaignStore.campaign.max_transactions_budget"
          :error-message="errors?.max_transactions_budget || ''"
          @handle-blur="handleBlur($event, 'max_transactions_budget')"
        >
          <template #icon-right>
            <UIGhostSelector
              :selected-option="selectedAmountType.max_amount"
              :options="amountTypes"
              @handle-option-select="handlePriceTypeSelect($event, 'max_amount')"
            />
          </template>
        </UIBaseInput>
      </div>
    </div>
  </div>
</template>
<script setup>
import UISelect from "../../UI/UISelect.vue";
import {computed, ref, watch} from "vue";
import UISectionTitleWithBorder from "../../UI/UISectionTitleWithBorder.vue";
import UIBaseInput from "../../UI/UIBaseInput.vue";
import UINumberIncreaseDecrease from "../../UI/UINumberIncreaseDecrease.vue";
import UIToggleGroup from "../../UI/UIToggleGroup.vue";
import UITabs from "../../UI/UITabs.vue";
import UITab from "../../UI/UITab.vue";
import UIRoundToggle from "../../UI/UIRoundToggle.vue";
import UIGhostButtonsGroup from "../../UI/UIGhostButtonsGroup.vue";
import UIDropdown from "../../UI/UIDropdown.vue";
import {useCampaignsStore} from "../../../store/campaignsStore.js";
import SVGPlus from "../../SVG/SVGPlus.vue";
import UIButton from "../../UI/UIButton.vue";
import {useRouter} from "vue-router";
import {calculateBudget, toDynamicFix} from "../../../helpers/index.js";
import {NANO_IN_SECOND} from "../../../constants/const.js";
import UIGhostSelector from "../../UI/UIGhostSelector.vue";

const props = defineProps({
  jitoData: {type: Object, default: null},
  priorityFees: {type: Object, default: null},
  projects: {type: Array, default: []},
  errors: {type: Object, default: () => ({})},
  isEditMode: {type: Boolean, default: false},
  isRouteChanged: {type: Boolean, default: false},
  campaignAction: {type: String, default: ''},
})
const emits = defineEmits(['handleErrorClear'])
const router = useRouter();
const campaignStore = useCampaignsStore();
const isSlippageCustom = ref(false);
const selectedProject = ref(null);
const isProjectOptionsOpen = ref(false);
const selectedTime = ref({
  max_time_between_transactions: 1,
  min_time_between_transactions: 1,
});
const amountTypes = ref([
  {name: 'SOL', val: 'sol'},
  {name: 'USD', val: 'usd'},
]);
const selectedAmountType = ref({
  budget: amountTypes.value[0],
  min_amount: amountTypes.value[0],
  max_amount: amountTypes.value[0]
});
const tokenSymbol = computed(() => {
  if (props.campaignAction === 'pull-up') return 'SOL';
  else return campaignStore.selectedToken?.symbol ? String(campaignStore.selectedToken?.symbol).toUpperCase() : '$TOKEN';
})
const jitoOptions = ['default', 'fast', 'extra'];
const validatorTipOptions = ['low', 'medium', 'high'];
const budgetOptions = [
  {label: '1%', val: 1},
  {label: '5%', val: 5},
  {label: '10%', val: 10},
  {label: '20%', val: 20},
  {label: 'max', val: 100},
];
const timeOptions = [
  {label: '5', val: 5},
  {label: '10', val: 10},
  {label: '30', val: 30},
];
const slippageOptions = [
  {label: '1%', val: 1},
  {label: '2%', val: 2},
  {label: '3%', val: 3},
  {label: '5%', val: 5},
  {label: '10%', val: 10},
];
const timeArray = [
  {label: 'Time Between Transactions', type: 'min_time_between_transactions', add_text: 'Min, s'},
  {label: '', type: 'max_time_between_transactions', add_text: 'Max, s'},
];


const handleTimeSelect = ({type, val}) => {
  if (!val) return;

  if (type === 'max_time_between_transactions'
    && val < selectedTime.value.min_time_between_transactions) return;

  selectedTime.value[type] = val;
  campaignStore.campaign[type] = val * NANO_IN_SECOND;
}
const setPriorityFee = (tip) => {
  const priorityFees = props.priorityFees;
  if (!tip || !priorityFees) return;

  campaignStore.campaign.priority_fee = priorityFees[tip];
}
const projectBalance = (project) => {
  if (props.campaignAction === 'pull-up') {
    return `${toDynamicFix(project?.total_balance_sol || 0)} SOL`;
  } else {
    return `${toDynamicFix(project?.total_balance || 0)} ${tokenSymbol.value}`;
  }
}
const handleTimeChangeByStep = ({action = '', type = ''}) => {
  const minSec = selectedTime.value.min_time_between_transactions;
  const maxSec = selectedTime.value.max_time_between_transactions;

  if (action === 'increase' && type) {
    const newVal = +selectedTime.value[type] + 1;
    selectedTime.value[type] = newVal;
    campaignStore.campaign[type] = newVal * NANO_IN_SECOND;

    if (newVal > maxSec) {
      selectedTime.value.max_time_between_transactions = newVal + 1;
      campaignStore.campaign.max_time_between_transactions = (newVal + 1) * NANO_IN_SECOND;
    }
  } else if (action === 'decrease' && type) {
    const newVal = +selectedTime.value[type] - 1;
    if (newVal < 1) {
      selectedTime.value[type] = 1
      campaignStore.campaign[type] = NANO_IN_SECOND;

      return;
    } else {
      selectedTime.value[type] = newVal;
      campaignStore.campaign[type] = newVal * NANO_IN_SECOND;
    }

    if (type === 'max_time_between_transactions' && newVal < minSec) {
      selectedTime.value.min_time_between_transactions = newVal;
      campaignStore.campaign.min_time_between_transactions = newVal * NANO_IN_SECOND;
    }
  }
}

const handleSlippageSelect = (option) => {
  if (!option) return;

  campaignStore.campaign.slippage = option.val;
}

const handleMaxBudget = (val) => {
  if (!selectedProject.value || !val) return;
  const projectBalanceAmount = (props.campaignAction === 'pull-up' ? selectedProject.value?.total_balance_sol : selectedProject.value?.total_balance) || 0;


  if (val === 100) {
    campaignStore.campaign.budget = projectBalanceAmount || 0;
  } else {
    campaignStore.campaign.budget = calculateBudget(projectBalanceAmount, val);
  }
}

const handlePriceTypeSelect = (option, type='') => {
  if (!option) return;

  selectedAmountType.value[type] = option;
}

const openPage = (type) => {
  if (type === 'project') {
    isProjectOptionsOpen.value = false;
    router.push({name: 'WalletsProjects'});
  }
}

const handleProjectSelect = (data) => {
  if (!data) return;

  selectedProject.value = data;
  campaignStore.campaign.project_id = data.id;
  isProjectOptionsOpen.value = false;
  emits('handleErrorClear', 'project')
}

function handleInput(event, field) {
  const raw = String(event.target.value);
  const cleaned = raw.replace(/[^\d.]/g, '');
  const val = Number(cleaned || 0);

  if (cleaned.length === 0) return;

  if (isNaN(val)) {
    campaignStore.campaign[field] = 0;
    return;
  }
  if (field === 'budget') {
    if (!selectedProject.value) {
      campaignStore.campaign.budget = 0;
    }

    const projectBalance = (props.campaignAction === 'pull-up' ? selectedProject.value?.total_balance_sol : selectedProject.value?.total_balance) || 0;

    if (val > projectBalance) {
      campaignStore.campaign[field] = projectBalance;
    }
  }

  const timeFields = ['min_time_between_transactions', 'max_time_between_transactions'];

  if (timeFields.includes(field)) {
    if (Number(val) < 1) {
      selectedTime.value[field] = 1;
      campaignStore.campaign[field] = NANO_IN_SECOND;
    } else {
      selectedTime.value[field] = val;
      campaignStore.campaign[field] = val * NANO_IN_SECOND;
    }

    return;
  }

  if (Number(val) > 100) {
    const fields = ['slippage', 'goal_percentage_change', 'parallel_transactions_amount'];

    if (fields.includes(field)) {
      campaignStore.campaign[field] = 100;
    }
  }

  emits('handleErrorClear', field);
}

function handleBlur(event, field) {
  const val = String(event.target.value);
  const campaign = campaignStore.campaign;

  if (!val) {
    switch (field) {
      case "budget":
        campaign[field] = 0;
        break;
      case "min_time_between_transactions":
        if (!val) {
          selectedTime.value[field] = 1;
          campaign[field] = NANO_IN_SECOND;
        }
        break;
      case "max_time_between_transactions":
        const min = selectedTime.value.min_time_between_transactions;
        const max = selectedTime.value[field];

        if (min > max) {
          selectedTime.value.max_time_between_transactions = min;
          campaign[field] = min * NANO_IN_SECOND;
        }

        break;
      case "min_transactions_budget":
        campaign[field] = campaign.budget;
        break;
      case "max_transactions_budget":
        campaign[field] = campaign.min_transactions_budget;
        break;
      case "parallel_transactions_amount":
        campaign[field] = 1;
        break;
      default:
        campaign[field] = 0;
    }
  } else {
    if (field === 'min_time_between_transactions') {
      const min = selectedTime.value.min_time_between_transactions;
      const max = selectedTime.value.max_time_between_transactions;

      if (min > max) {
        selectedTime.value.max_time_between_transactions = (min + 1);
        campaign.max_time_between_transactions = (min + 1) * NANO_IN_SECOND;
      }
    }
  }
}

watch(() => props.isRouteChanged, (newVal) => {
  if (newVal) {
    selectedAmountType.value = {
      budget: amountTypes.value[0],
      min_amount: amountTypes.value[0],
      max_amount: amountTypes.value[0]
    };
  }
})

watch(() => campaignStore.campaign, (val) => {
    const min = campaignStore.campaign.min_time_between_transactions / NANO_IN_SECOND;
    const max = campaignStore.campaign.max_time_between_transactions / NANO_IN_SECOND;

    if (min !== selectedTime.value.min_time_between_transactions) {
      selectedTime.value.min_time_between_transactions =
        Math.floor((val.min_time_between_transactions || NANO_IN_SECOND) / NANO_IN_SECOND);
    }

    if (max !== selectedTime.value.max_time_between_transactions) {
      selectedTime.value.max_time_between_transactions =
        Math.floor((val.max_time_between_transactions || NANO_IN_SECOND) / NANO_IN_SECOND);
    }

    if (props.isEditMode) {
      const slippageOption = slippageOptions.find(opt => opt.val === campaignStore.campaign.slippage);

      if (!slippageOption) isSlippageCustom.value = true;
    }

    if (!val.project_id && selectedProject.value) {
      selectedProject.value = null;
      isSlippageCustom.value = false;
    }
  },
  {immediate: true, deep: true}
);

watch(() => campaignStore.campaign.budget, (newVal) => {
  if (newVal) {
    campaignStore.campaign.min_transactions_budget = toDynamicFix(calculateBudget(newVal, 20));
    campaignStore.campaign.max_transactions_budget = toDynamicFix(calculateBudget(newVal, 60));
  } else {
    campaignStore.campaign.min_transactions_budget = 0;
    campaignStore.campaign.max_transactions_budget = 0;
  }
})

watch(() => props.projects, (newVal) => {
  if (newVal.length && props.isEditMode) {
    selectedProject.value = props.projects.find(project => project.id === campaignStore.campaign.project_id);
  }
}, {immediate: true, deep: true});
</script>
<style scoped lang="scss">
.transaction-setup {
  &__inputs {
    margin-top: 20px;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  &__divider {
    height: 1px;
    width: 100%;
    background: #D1D5DB;
    display: none;
  }

  &__slippage {
    display: flex;
    flex-direction: column;
    gap: 4px;

    & span {
      color: #6B7280;
      font-size: 14px;
      font-style: normal;
      font-weight: 400;
      line-height: 150%;
    }

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  &__speed {
    display: flex;
    flex-direction: column;
    gap: 4px;

    & span {
      text-transform: capitalize;
    }

    & .jito-data {
      color: #6B7280;
      font-size: 10px;
      font-style: normal;
      font-weight: 500;
      line-height: 150%;
      letter-spacing: 0.05px;
    }
  }

  &__input {
    width: calc(50% - 6px);

    &:nth-child(-n+4) {
      margin-bottom: 8px;
    }

    & .custom-option {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    &_multiple {
      width: 100%;

      display: flex;
      gap: 12px;
      align-items: center;
    }

    &.budget-min {
      padding-top: 0;
    }

    &.budget-max {
      padding-top: 26px;
    }

    &_controls {
      display: flex;
      gap: 12px;
      align-items: center;

      & span {
        display: block;
        width: max-content;
      }
    }

    & .budjet-dropdown {
      position: relative;
      margin-left: auto;

      &__selected {
        display: flex;
        align-items: center;
        background: transparent;
        padding: 0 8px;
        gap: 16px;

        & svg {
          transition: .3s ease;
        }

        &.open {
          & svg {
            transform: rotate(180deg);
          }
        }
      }

      &__options {
        position: absolute;
        z-index: 10;
        top: 100%;
        right: 24px;
        width: 133px;
      }
    }
  }

  &__error {
    color: #DC2626;
    font-weight: 400;
    margin-top: 6px;
  }
}

.custom-dropdown-content {
  height: 165px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;

  & p {
    color: #000;
    margin-bottom: 8px;
    max-width: 209px;
    text-align: center;
  }
}
</style>