<template>
  <div class="add-budget">

    <div class="add-budget__qnty">
      <UIBaseInput
        v-model="budget"
        label="Amount"
        placeholder="0"
        type="number"
        size="large"
        @handle-blur="handleInputBlur"
      >
        <template #bottom-right>
          <UIGhostButtonsGroup
            :options="budgetOptions"
            :selected-option="selectedBudgetOption"
            @handle-option-select="handleAddBudget"
          />
        </template>
      </UIBaseInput>
    </div>

    <div :class="['add-budget__btns']">
      <UIButton
        color_type="outline"
        @cta="modalsStore.closeModal"
      >
        Cancel
      </UIButton>
      <UIButton color_type="primary" @cta="handleSaveBudget">
        Add budget
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import {useModalsStore} from "../../../../store/modalsStore.js";
import UIButton from "../../../UI/UIButton.vue";
import UIBaseInput from "../../../UI/UIBaseInput.vue";
import UIGhostButtonsGroup from "../../../UI/UIGhostButtonsGroup.vue";
import {computed, ref, watch} from "vue";
import {toDynamicFix} from "../../../../helpers/index.js";

const props = defineProps({
  modelValue: {type: Number, default: 0},
  campaignAction: {type: String, default: ''},
})
const emit = defineEmits(['update:modelValue'])
const modalsStore = useModalsStore();
const selectedBudgetOption = ref('');
const projectData = computed(() => {
  return modalsStore.modalData.project;
})
const remainingBudget = computed(() => {
  if (!modalsStore.modalData.item && !projectData.value) return 0;

  const projectBalance = props.campaignAction === 'pull-up' ? projectData.value.total_balance_sol : projectData.value.total_balance;

  return projectBalance - modalsStore.modalData.item.budget;
})

const budget = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', Number(val))
})

const budgetOptions = computed(() => {
  const base = [
    {label: '1%', val: 1},
    {label: '10%', val: 10},
    {label: '15%', val: 15},
    {label: '20%', val: 20},
    {label: 'max', val: 100},
  ];
  const remaining = remainingBudget.value;
  return base.map(opt => ({
    ...opt,
    budget: toDynamicFix((remaining * Math.min(opt.val, 100)) / 100),
  }));
});

function syncSelectedBudgetOption() {
  const current = Number(props.modelValue);
  if (!Number.isFinite(current) || current === 0) {
    selectedBudgetOption.value = '';
    return;
  }
  const match = budgetOptions.value.find(
    opt => Math.abs(Number(opt.budget) - current) < 1e-9
  );
  selectedBudgetOption.value = match ? String(match.val) : '';
}

const handleInputBlur = () => {
  const current = Number(props.modelValue);
  if (!Number.isFinite(current)) return;
  if (current > remainingBudget.value) {
    emit('update:modelValue', toDynamicFix(remainingBudget.value));
  }
}

const handleSaveBudget = () => {
  modalsStore.modalData.type = 'budget-confirmation';
  modalsStore.modalData.title = 'Confirm budget'
  modalsStore.modalData.action = 'confirmation'
}

const handleAddBudget = (val) => {
  const percent = Number(val);
  if (!Number.isFinite(percent) || percent <= 0) return;
  selectedBudgetOption.value = String(val);
  const value = (remainingBudget.value * Math.min(percent, 100)) / 100;
  emit('update:modelValue', toDynamicFix(value));
}

watch(() => props.modelValue, syncSelectedBudgetOption, { immediate: true });
</script>
<style scoped lang="scss">

.add-budget {
  width: 360px;

  &__top {
    display: block;
    margin-bottom: 16px;
  }

  &__btns {
    margin-top: 16px;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;

    &.disabled {
      opacity: .5;
      pointer-events: none;
    }
  }
}

@media (max-width: 1200px) {
  .add-budget {
    width: 100%;
  }
}
</style>