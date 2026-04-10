<template>
  <div class="stop-campaign">
    <div class="stop-campaign__inner">
      <div class="stop-campaign__header">
        <SVGTriangleWarning/>
        <span class="heading-4">{{modalStore.modalData.title}}</span>
      </div>
      <div class="text-1 paragraph-small regular grey">You are about to stop the campaign <span class="paragraph-small bold grey">{{token?.name}}</span>.</div>
      <p class="paragraph-small regular grey">All running transactions will be halted. Remaining budget will stay in your project wallet.</p>
      <span class="paragraph-small regular red">This action cannot be undone.</span>
      <div class="stop-campaign__campaign">
        <div class="stop-campaign__campaign_item">
          <span class="paragraph-small regular">Campaign:</span>
          <span class="paragraph-medium medium">{{token?.name}}</span>
        </div>
        <div class="stop-campaign__campaign_item">
          <span class="paragraph-small regular">Status:</span>
          <div class="stop-campaign__campaign_status">
            <div class="color monospaced-small"></div>
            Active
          </div>
        </div>
        <div class="stop-campaign__campaign_item">
          <span class="paragraph-small regular">Progress:</span>
          <div class="stop-campaign__campaign_spend">
            <div class="monospaced-medium info"><span>{{toDynamicFix(modalStore.modalData.item?.spent_budget || 0)}}</span>/{{toDynamicFix(modalStore.modalData.item?.budget || 0)}} Sol</div>
            <div class="progress paragraph-small">
              {{spendPercent}}%
              <RoundedProgress
                :value="spendPercent"
                :show-text="false"
              />
            </div>
          </div>
        </div>
      </div>
      <div class="stop-campaign__btns">
        <UIButton color_type="outline" @cta="modalStore.closeModal">
          Cancel
        </UIButton>
        <UIButton color_type="destructive" @cta="emits('handleStopCampaign')">
          Stop campaign
        </UIButton>
      </div>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../../../UI/UIButton.vue";
import SVGTriangleWarning from "../../../SVG/SVGTriangleWarning.vue";
import {useModalsStore} from "../../../../store/modalsStore.js";
import RoundedProgress from "../../../UI/RoundedProgress.vue";
import {useTokensStore} from "../../../../store/tokensStore.js";
import {computed} from "vue";
import {toDynamicFix} from "../../../../helpers/index.js";

const modalStore = useModalsStore();
const emits = defineEmits(['handleStopCampaign']);
const tokensStore = useTokensStore();
const token = computed(() => {
  if (!modalStore.modalData.item || !tokensStore.solTokensData[modalStore.modalData.item?.token_mint_from]) return null;

  return tokensStore.solTokensData[modalStore.modalData.item.token_mint_from];
})

const spendPercent = computed(() => {
  const campaign = modalStore.modalData.item;
  if (!campaign?.budget || campaign.budget <= 0) return 0;
  const progress = (campaign.spent_budget / campaign.budget) * 100;
  return Math.floor(Math.min(100, Math.max(0, progress)))
})
</script>
<style scoped lang="scss">
.stop-campaign {
  &__inner {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &__header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__campaign {
    padding: 16px;
    border-radius: 8px;
    border: 1px solid #E5E7EB;
    background: #F9FAFB;
    display: flex;
    flex-direction: column;
    gap: 8px;

    &_item {
      height: 28px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    &_status {
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

    &_spend {
      display: flex;
      align-items: center;
      gap: 8px;

      & .label {
        color: #6B7280;
        font-weight: 500;
      }

      & .info {
        color: #6B7280;

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
      }
    }
  }

  &__btns {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: flex-end;
  }
}
</style>