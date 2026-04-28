<template>
  <div class="triggers">
    <div class="triggers__inner">
      <UISectionTitleWithBorder>Targets and Stop triggers</UISectionTitleWithBorder>
      <div class="triggers__inputs">
        <UIBaseInput
          size="large"
          placeholder="Custom"
          label="Targets"
          type="number"
          v-model="campaignStore.campaign.goal_percentage_change"
          :error-message="errors?.goal_percentage_change || ''"
          @handle-input="handleInput"
          @handle-blur="handleBlur($event, 'goal_percentage_change')"
        >
          <template #icon-left>
            <span>{{campaignAction === 'pull-up' ? '+' : '-'}}</span>
          </template>
          <template #icon-right>
            <span>%</span>
          </template>
        </UIBaseInput>
      </div>
    </div>
  </div>
</template>
<script setup>
import UISectionTitleWithBorder from "../../UI/UISectionTitleWithBorder.vue";
import UIBaseInput from "../../UI/UIBaseInput.vue";
import {useCampaignsStore} from "../../../store/campaignsStore.js";

defineProps({
  campaignAction: {type: String, default: ''},
  errors: {type: Object, default: () => ({})},
})
const emits = defineEmits(['handleErrorClear'])
const campaignStore = useCampaignsStore();

function handleInput(event) {
  const raw = String(event.target.value).replace(/,/g, '.');
  const cleaned = raw.replace(/[^\d.]/g, '');
  const val = Number(cleaned || 0);

  if (!cleaned.length) return;

  if (isNaN(val) || val < 0) {
    campaignStore.campaign.goal_percentage_change = 0;
    return;
  }

  if (val > 100) {
    campaignStore.campaign.goal_percentage_change = 100;
  } else {
    campaignStore.campaign.goal_percentage_change = val;
  }

  emits('handleErrorClear', 'goal_percentage_change');
}

const handleBlur = (event, field) => {
  if (!field) return;
  const val = String(event.target.value).replace(/,/g, '.');

  if (!val) {
    campaignStore.campaign[field] = 0;
  }
}
</script>
<style scoped lang="scss">
.triggers {
  &__inner {
    display: flex;
    flex-direction: column;
  }

  &__inputs {
    margin-top: 20px;
    display: flex;
    gap: 12px;

    & .base-input {
      width: calc(50% - 6px)
    }

    & span {
      color: #6B7280;
      font-size: 14px;
      font-style: normal;
      font-weight: 400;
      line-height: 150%;
    }
  }
}
</style>