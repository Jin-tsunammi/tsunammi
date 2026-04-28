<template>
  <div class="campaign-estimate">
    <div class="campaign-estimate__inner">
      <div class="campaign-estimate__top">
        <div class="paragraph-medium">Campaign launch estimate</div>
      </div>
      <div class="campaign-estimate__breakdown">
        <div class="campaign-estimate__breakdown_item">
          <div class="paragraph-small medium">Budget</div>
          <span class="monospaced-small regular">{{`${toDynamicFix(estimateData?.budget_sol || 0)} ${tokenSymbol}`}}</span>
        </div>
        <div v-if="campaignAction === 'pull-up'" class="campaign-estimate__breakdown_item">
          <div class="paragraph-small medium" @mouseenter="toggleUIKit('network-fee')" @mouseleave="toggleUIKit('network-fee')">
            Network fee
            <div class="ttip">
              <div class="tooltip paragraph-mini">
                <UIToolTip
                  position="bottom"
                  :is-shown="UIKitVisible === 'network-fee'"
                  text="Includes account rent (ATA creation) and Solana validator tip."
                />
              </div>
              <SVGAlertInfo color="#4B5563"/>
            </div>
          </div>
          <span class="monospaced-small regular">{{`${toDynamicFix(estimateData?.rent_sol || 0)} ${tokenSymbol}`}}</span>
        </div>
        <div v-if="campaignAction === 'pull-up'" class="campaign-estimate__breakdown_item">
          <div class="paragraph-small medium" @mouseenter="toggleUIKit('frozen-money')" @mouseleave="toggleUIKit('frozen-money')">
            Frozen money
            <div class="ttip">
              <div class="tooltip paragraph-mini">
                <UIToolTip
                  position="bottom"
                  :is-shown="UIKitVisible === 'frozen-money'"
                  text="Funds reserved by the network for wallet-related operations, such as associated token account creation."
                />
              </div>
              <SVGAlertInfo color="#4B5563"/>
            </div>
          </div>
          <span class="monospaced-small regular">{{`${toDynamicFix(estimateData?.tip_sol || 0)} ${tokenSymbol}`}} </span>
        </div>
      </div>
      <div class="campaign-estimate__divider"></div>
      <div class="campaign-estimate__total">
        <div class="paragraph-medium medium label">Total on launch</div>
        <span class="monospaced-medium medium">{{`${totalAmount} ${tokenSymbol}`}}</span>
      </div>
      <div class="campaign-estimate__divider"></div>
      <div class="campaign-estimate__bottom">
        <span class="paragraph-small regular grey">Estimated values. Final network fee may slightly vary.</span>
      </div>
    </div>
  </div>
</template>
<script setup>
import {computed, ref} from "vue";
import UIToolTip from "../../UI/UIToolTip.vue";
import SVGAlertInfo from "../../SVG/SVGAlertInfo.vue";
import {toDynamicFix} from "../../../helpers/index.js";
import {useCampaignsStore} from "../../../store/campaignsStore.js";

const props = defineProps({
  estimateData: {type: Object, default: null},
  campaignAction: {type: String, default: ''},
})
const campaignStore = useCampaignsStore();
const UIKitVisible = ref('');
const tokenSymbol = computed(() => {
  if (props.campaignAction === 'pull-up') return 'Sol';
  else return campaignStore.selectedToken?.symbol || '';
})
const totalAmount = computed(() => {
  if (!props.estimateData) {
    return 0;
  } else {
    if (props.campaignAction === 'pull-up') {
      return toDynamicFix(props.estimateData.budget_sol + props.estimateData.rent_sol + props.estimateData.tip_sol + campaignStore.campaign.priority_fee);
    } else {
      return toDynamicFix(props.estimateData.budget_sol);
    }
  }
})

const toggleUIKit = (type) => {
  if (UIKitVisible.value !== type) {
    UIKitVisible.value = type;
  } else {
    UIKitVisible.value = '';
  }
}
</script>
<style scoped lang="scss">
.campaign-estimate {
  &__inner {
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: #fff;
  }

  & .tooltip {
    position: absolute;
    bottom: calc(100% + 10px);
    left: 50%;
    transform: translateX(-50%);
    width: 205px;
    z-index: 5;
    font-weight: 400;

    &-wrapper {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  &__top {
    height: 56px;
    padding: 0 16px;
    display: flex;
    align-items: center;
    background: #F9FAFB;
  }

  &__divider {
    margin: 12px 16px;
    width: 100%;
    height: 1px;
    background: #E5E7EB;
  }

  &__total {
    padding: 0 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;

    & .label {
      color: #EA580C;
    }
  }

  &__breakdown {
    margin-top: 20px;
    display: flex;
    flex-direction: column;
    padding: 0 16px;
    gap: 12px;

    &_item {
      display: flex;
      align-items: center;
      justify-content: space-between;

      & .paragraph-small {
        display: flex;
        align-items: center;
        gap: 6px;
      }

      & .ttip {
        position: relative;
        display: flex;
        align-items: center;
      }
    }
  }

  &__bottom {
    padding: 0 16px 16px 16px;
  }
}
</style>