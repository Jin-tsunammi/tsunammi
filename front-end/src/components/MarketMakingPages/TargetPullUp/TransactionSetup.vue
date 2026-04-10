<template>
  <div class="transaction-setup">
    <div class="transaction-setup__inner">
      <UISectionTitleWithBorder>Transaction setup</UISectionTitleWithBorder>

      <div class="transaction-setup__inputs">
        <UISelect
          class="transaction-setup__input"
          v-model="isProjectOptionsOpen"
          :selected="selectedProject?.name || ''"
          label="Select Project wallets"
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
                <div class="custom-option paragraph-small regular">{{ item?.name || '' }} <span
                  class="monospaced-small grey">{{ projectBalance(item) }}</span></div>
              </template>
              <template #custom-dropdown>
                <div class="custom-dropdown-content" @click.stop>
                  <div class="paragraph-small bold">You do not have projects yet.</div>
                  <p class="paragraph-mini regular">Create projects and wallets to start funding.</p>
                  <UIButton
                    color_type="ghost"
                    size="large"
                    @cta="openPage('project')"
                  >
                    Create project

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
          :label="`Set budget for pull-${campaignAction === 'pull-up' ? 'up' : 'down'}`"
          type="number"
          size="large"
          v-model="campaignStore.campaign[selectedBudgetType.val]"
          :error-message="errors?.budget || ''"
          @handle-input="handleInput($event, 'budget')"
        >
          <template v-if="selectedBudgetType?.symbol" #icon-right><span
            class="paragraph-mini regular grey">{{ selectedBudgetType?.symbol }}</span></template>
          <template #bottom-right>
            <UIGhostButtonsGroup
              :options="budgetOptions"
              @handle-option-select="handleMaxBudget"
            />
          </template>
        </UIBaseInput>

        <div class="transaction-setup__divider"></div>

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
          >
            <template #icon-right><span>%</span></template>
          </UIBaseInput>
        </div>
        <div v-show="campaignStore.campaign.using_jito" class="transaction-setup__speed transaction-setup__input">
          <div class="paragraph-small medium transaction-setup__input_label">Speed of transaction</div>
          <UITabs size="large">
            <UITab
              v-for="k in jitoOptions"
              :key="k"
              :is_active="campaignStore.campaign.transaction_speed === k"
              @click="campaignStore.campaign.transaction_speed = k"
            >
              <template #default>
                <span class="paragraph-mini medium">{{ k }}</span>
              </template>
              <template #add-info>
                <span class="jito-data" v-if="jitoData?.[k]">{{ jitoData?.[k].toFixed(7) }}</span>
              </template>
            </UITab>
          </UITabs>
          <span v-if="errors?.transaction_speed"
                class="paragraph-mini transaction-setup__error">{{ errors.transaction_speed }}</span>
        </div>

        <div class="transaction-setup__divider"></div>

        <div class="transaction-setup__input_multiple">
          <UIBaseInput
            class="transaction-setup__input"
            size="large"
            label="Multiple transactions simultaneously"
            type="number"
            v-model="campaignStore.campaign.parallel_transactions_amount"
            :error-message="errors?.parallel_transactions_amount || ''"
            @handle-input="handleInput($event, 'parallel_transactions_amount')"
            :is_dot_allowed="false"
          >

          </UIBaseInput>
        </div>

        <UIBaseInput
          v-for="timeEl in timeArray"
          :key="timeEl.type"
          class="transaction-setup__input"
          :label="timeEl.label"
          size="large"
          type="number"
          :top-add-text="timeEl.add_text"
          :is_dot_allowed="false"
          v-model="selectedTime[timeEl.type]"
          :error-message="errors?.[timeEl.type] || ''"
          @handle-input="handleInput($event, timeEl.type)"
          @handle-blur="checkExecutionIntervalOnBlur(timeEl.type)"
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
          label="Transactions Budget Range"
          class="transaction-setup__input budget-min"
          size="large"
          top-add-text="Min"
          type="number"
          v-model="campaignStore.campaign.min_transactions_budget"
          :error-message="errors?.min_transactions_budget || ''"
          @handle-input="handleInput($event, 'min_transactions_budget')"
        />

        <UIBaseInput
          class="transaction-setup__input budget-max"
          size="large"
          top-add-text="Max"
          type="number"
          v-model="campaignStore.campaign.max_transactions_budget"
          :error-message="errors?.max_transactions_budget || ''"
        />
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
import SVGSmallArrowDown from "../../SVG/SVGSmallArrowDown.vue";
import UIDropdown from "../../UI/UIDropdown.vue";
import {useCampaignsStore} from "../../../store/campaignsStore.js";
import {OnClickOutside} from "@vueuse/components";
import SVGPlus from "../../SVG/SVGPlus.vue";
import UIButton from "../../UI/UIButton.vue";
import {useRouter} from "vue-router";
import {useProjectsStore} from "../../../store/projectsStore.js";
import {calculateBudget, toDynamicFix} from "../../../helpers/index.js";

const props = defineProps({
  jitoData: {type: Object, default: null},
  projects: {type: Array, default: []},
  errors: {type: Object, default: () => ({})},
  isEditMode: {type: Boolean, default: false},
  campaignAction: {type: String, default: ''},
})
const emits = defineEmits(['handleErrorClear'])
const NANO_IN_SECOND = 1000000000;
const router = useRouter();
const campaignStore = useCampaignsStore();
const isSlippageCustom = ref(false);
const selectedProject = ref(null);
const isProjectOptionsOpen = ref(false);
const isBudgetOptionsOpen = ref(false);
const jitoOptions = ['default', 'fast', 'extra'];
const tokenSymbol = computed(() => {
  if (props.campaignAction === 'pull-up') return 'Sol';
  else return campaignStore.selectedToken?.symbol || '';
})
const budgetOptions = [
  {label: '1%', val: 1},
  {label: '5%', val: 5},
  {label: '10%', val: 10},
  {label: '20%', val: 20},
  {label: 'max', val: 100},
];
const budgetDropdown = computed(() => {
  return [
    {label: "Price", val: 'budget', symbol: tokenSymbol.value},
    {label: "Percentage", val: 'budget_percent', symbol: '%'},
  ]
});
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
  {label: 'Task Execution Interval', type: 'min_time_between_transactions', add_text: 'Min, s'},
  {label: '', type: 'max_time_between_transactions', add_text: 'Max, s'},
];
const selectedTime = ref({
  max_time_between_transactions: 1,
  min_time_between_transactions: 1,
});
const selectedBudgetType = ref(budgetDropdown.value[0]);

const handleTimeSelect = ({type, val}) => {
  if (!val) return;

  if (type === 'max_time_between_transactions'
    && val < selectedTime.value.min_time_between_transactions) return;

  selectedTime.value[type] = val;
  campaignStore.campaign[type] = val * NANO_IN_SECOND;
}
const projectBalance = (project) => {
  if (props.campaignAction === 'pull-up') {
    return `${toDynamicFix(project?.total_balance_sol || 0)} Sol`;
  } else {
    return `${toDynamicFix(project?.total_balance || 0)} ${tokenSymbol.value}`;
  }
}
const handleTimeChangeByStep = ({action = '', type = ''}) => {
  const min = 1;

  if (action === 'increase' && type) {
    const newVal = +selectedTime.value[type] + 1;
    selectedTime.value[type] = newVal;
    campaignStore.campaign[type] = newVal * NANO_IN_SECOND;
  } else if (action === 'decrease' && type) {
    const newVal = +selectedTime.value[type] - 1;
    if (newVal < min) {
      selectedTime.value[type] = min
      campaignStore.campaign[type] = min * NANO_IN_SECOND;
    } else {
      selectedTime.value[type] = newVal;
      campaignStore.campaign[type] = newVal * NANO_IN_SECOND;
    }
  }
}

const handleBudjetTypeSelect = (data) => {
  if (!data) return;

  selectedBudgetType.value = data;
  campaignStore.campaign.budget = 0;
  campaignStore.campaign.budget_percent = 0;
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
  const raw = String(event.target.value).replace(/,/g, '.');
  const cleaned = raw.replace(/[^\d.]/g, '');
  const val = Number(cleaned || 0);

  if (isNaN(val) || val < 0) {
    campaignStore.campaign[field] = 0;
    return;
  }
  if (field === 'budget') {
    if (!selectedProject.value) {
      campaignStore.campaign.budget = 0;
    }

    const projectBalance = props.campaignAction === 'pull-up' ? selectedProject.value.total_balance_sol : selectedProject.value.total_balance;

    if (selectedBudgetType.value.val === 'price') {
      if (val > projectBalance) {
        campaignStore.campaign[field] = projectBalance;
      }
    } else if (selectedBudgetType.value.val === 'percentage') {
      if (val > 100) {
        campaignStore.campaign[field] = 100;
      }
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

const checkExecutionIntervalOnBlur = (field) => {
  if (field === 'min_time_between_transactions') return;

  const min = selectedTime.value.min_time_between_transactions;
  const max = selectedTime.value.max_time_between_transactions;

  if (min > max) {
    selectedTime.value.max_time_between_transactions = min;
    campaignStore.campaign.max_time_between_transactions = min * NANO_IN_SECOND;
  }
}

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

// watch(selectedTime.value, () => {
//   const min = selectedTime.value.min_time_between_transactions;
//   campaignStore.campaign.min_time_between_transactions = min * NANO_IN_SECOND;
// }, {deep: true});
watch(() => props.projects, (newVal) => {
  if (newVal.length && props.isEditMode) {
    selectedProject.value = props.projects.find(project => project.id === campaignStore.campaign.project_id);
  }
}, {immediate: true, deep: true});
watch(() => campaignStore.selectedToken, (newVal) => {
  if (newVal && selectedBudgetType.value.val === 'price') {
    selectedBudgetType.value = budgetDropdown.value[0];
  }
}, {deep: true});
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