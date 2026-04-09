<template>
  <div class="profile-campaign">
    <div class="profile-campaign__top">
      <div class="profile-campaign__info">
        <UIAvatarShow
          class="profile-campaign__logo"
          :mint="campaignMint"
          :src="token?.image"
        />
        <div class="profile-campaign__title">
          <div class="name">
            <span>{{token?.name}}</span>
          </div>
          <span class="symbol">({{token?.symbol}})</span>
        </div>
      </div>

      <div class="profile-campaign__details">
        <div class="profile-campaign__status">
          <div class="color monospaced-small"></div>
          Active
        </div>

        <div class="profile-campaign__spend">
          <div class="paragraph-small label">Spend</div>
          <div class="monospaced-medium info">
            <span>{{ toDynamicFix(campaign?.spent_budget ?? 0) }}</span>/{{`${toDynamicFix(campaign?.budget ?? 0)} ${tokenSymbol}`}}
          </div>
          <div class="progress paragraph-small">
            {{ Math.round(spendPercent) }}%
            <RoundedProgress
              :value="spendPercent"
              :show-text="false"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="profile-campaign__bottom">
      <UIButton color_type="outline" size="large" @cta="emits('handleStop')">
        Stop
      </UIButton>
      <UIButton color_type="primary" size="large" @cta="emits('handleAddBudget')">
        Add Budget
      </UIButton>

      <UIButton class="edit" color_type="ghost" size="large" @cta="emits('handleEdit')">
        <template #left-icon><SVGEdit /></template>
        Edit
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import { computed } from 'vue'
import RoundedProgress from "../UI/RoundedProgress.vue";
import UIButton from "../UI/UIButton.vue";
import UIAvatarShow from "../UI/UIAvatarShow.vue";
import { toDynamicFix } from "../../helpers/index.js";
import {useTokensStore} from "../../store/tokensStore.js";
import SVGEdit from "../SVG/SVGEdit.vue";

const props = defineProps({
  campaign: { type: Object, default: null },
  campaignAction: { type: String, default: ''},
})

const emits = defineEmits(['handleStop', 'handleAddBudget', 'handleEdit']);
const tokensStore = useTokensStore();

const mintType = computed(() => {
  if (props.campaignAction === 'pull-up') return 'token_mint_to';
  return 'token_mint_from';
})
const campaignMint = computed(() => {
  return props.campaign?.[mintType.value] || '';
})
const token = computed(() => {
  if (!props.campaign || !tokensStore.solTokensData[campaignMint.value]) return null;

  return tokensStore.solTokensData[campaignMint.value];
})
const tokenSymbol = computed(() => {
  if (props.campaignAction === 'pull-down') return tokensStore.solTokensData[campaignMint.value]?.symbol || '';
  else return 'Sol';
})
const spendPercent = computed(() => {
  const campaign = props.campaign;
  if (!campaign?.budget || campaign.budget <= 0) return 0;
  const progress = (campaign.spent_budget / campaign.budget) * 100;
  return Math.min(100, Math.max(0, progress))
})
</script>
<style scoped lang="scss">
.profile-campaign {
  display: flex;
  flex-direction: column;
  height: 170px;
  border-radius: 8px;
  border: 1px solid #D1D5DB;

  &__title {
    color: #111827;
    font-size: 18px;
    font-style: normal;
    font-weight: 500;
    line-height: 150%;
    display: flex;
    align-items: center;
    gap: 5px;
    min-width: 0;

    & .name {
      display: flex;
      min-width: 0;
      max-width: 150px;
      overflow: hidden;

      & span {
        display: block;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    & .symbol {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      flex-shrink: 1;
      min-width: 0;
    }
  }

  &__top {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 16px 8px;
  }

  &__info {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    overflow: hidden;
  }

  &__details {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__status {
    color: #030712;
    display: flex;
    align-items: center;
    gap: 9px;

    & .color {
      background: #16A34A;
      aspect-ratio: 1/1;
      min-width: 8px;
      max-width: 8px;
      border-radius: 50%;
      box-shadow: 0 0.75px 0 0 rgba(255, 255, 255, 0.20) inset;
      filter: drop-shadow(0 0 0 #16A34A) drop-shadow(0 1px 2px rgba(22, 163, 74, 0.40));
    }
  }

  &__logo {
    border-radius: 50%;
    min-width: 28px;
    max-width: 28px;
    aspect-ratio: 1/1;
    overflow: hidden;
    display: flex;
    background: #F3F4F6;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);
  }

  &__bottom {
    display: flex;
    align-items: center;
    background: #E5E7EB;
    margin-top: auto;
    height: 72px;
    padding: 16px;
    gap: 8px;

    & .edit {
      margin-left: auto;
    }
  }

  &__spend {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    margin-left: 6px;

    & .label {
      color: #6B7280;
      font-weight: 500;
      flex-shrink: 0;
    }

    & .info {
      color: #6B7280;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      min-width: 0;
      flex: 1 1 0;

      & span {
        color: #111827;
      }
    }

    & .progress {
      height: 32px;
      border-radius: 7px;
      border: 1px solid #E5E7EB;
      background: #F9FAFB;
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 0 8px;
      font-weight: 500;
      color: #16A34A;
      flex-shrink: 0;
    }
  }
}
</style>