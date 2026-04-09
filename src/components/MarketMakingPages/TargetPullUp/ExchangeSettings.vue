<template>
  <div class="exchange-settings">
    <div class="exchange-settings__inner">
      <UISectionTitleWithBorder>Exchange Settings</UISectionTitleWithBorder>
      <div class="exchange-settings__inputs">
        <UISelect
          :selected="tokenName"
          v-model="isDropDownOpen.token"
          :label="label"
          placeholder="Select an item"
          size="large"
          :is_disabled="isEditMode"
          :error-message="errors?.dest_token_mint || ''"
          class="exchange-settings__input"
        >
          <template #dropdown>
            <UIDropdown
              :is-open="isDropDownOpen.token"
              :is-search="true"
              v-model:search="vModelSearch"
              :selected-option="selectedToken?.name || ''"
              :options="tokens"
              label="name"
              @handle-option-select="handleTokenSelect"
            >
              <template #custom-option="{item}">
                <div class="custom-option">
                  <span class="paragraph-small regular">{{ item.name }}</span>
                  <span class="paragraph-mini regular grey">{{ formatWalletAddress(item.address) }}</span>
                </div>
              </template>
            </UIDropdown>
          </template>
        </UISelect>
        <UISelect
          :selected="selectedDex.name"
          v-model="isDropDownOpen.dex"
          label="Select DEX"
          placeholder="Select DEX"
          size="large"
          :is_disabled="isEditMode"
          class="exchange-settings__input"
        >
          <template v-if="selectedDex?.image" #left-icon>
            <UIAvatarShow
              class="icon"
              :mint="selectedDex?.image || ''"
              :src="selectedDex?.image || ''"
              :is-token="false"
            />
          </template>
          <template #dropdown>
            <UIDropdown
              :is-open="isDropDownOpen.dex"
              :selected-option="selectedDex.name || ''"
              :options="dexList"
              label="name"
              @handle-option-select="handleDexSelect"
            />
          </template>
        </UISelect>
        <UISelect
          :selected="selectedValidator.name"
          v-model="isDropDownOpen.validator"
          label="Select Validator"
          placeholder="Select Validator"
          size="large"
          :is_disabled="isEditMode"
          class="exchange-settings__input"
        >
          <template v-if="selectedValidator?.image" #left-icon>
            <UIAvatarShow
              class="icon"
              :mint="selectedValidator?.image || ''"
              :src="selectedValidator?.image || ''"
              :is-token="false"
            />
          </template>
          <template #dropdown>
            <UIDropdown
              :is-open="isDropDownOpen.validator"
              :selected-option="selectedValidator.name || ''"
              :options="validators"
              label="name"
              @handle-option-select="handleValidatorSelect"
            />
          </template>
        </UISelect>
      </div>
    </div>
  </div>
</template>
<script setup>
import UISectionTitleWithBorder from "../../UI/UISectionTitleWithBorder.vue";
import UISelect from "../../UI/UISelect.vue";
import {computed, ref, watch} from "vue";
import UIDropdown from "../../UI/UIDropdown.vue";
import {useCampaignsStore} from "../../../store/campaignsStore.js";
import {formatWalletAddress} from "../../../helpers/index.js";
import {useTokensStore} from "../../../store/tokensStore.js";
import UIAvatarShow from "../../UI/UIAvatarShow.vue";
import Raydium from "../../../../public/images/raydium-icon.webp";
import PumpFun from "../../../../public/images/pumpfun-icon.webp";

const props = defineProps({
  isEditMode: {type: Boolean, default: false},
  tokens: {type: Array, default: []},
  errors: {type: Object, default: () => ({})},
  campaignAction: {type: String, default: ''},
})
const vModel = defineModel();
const vModelSearch = defineModel('search');
const campaignStore = useCampaignsStore();
const tokensStore = useTokensStore();
const isDropDownOpen = ref({
  token: false,
  dex: false,
  validator: false,
})
const selectedToken = ref(null);
const validators = [
  {name: 'Jito', val: 'jito', image: 'https://www.jito.wtf/jitoBig.png'},
  {name: 'Public Solana Node', val: 'custom', image: 'https://solana.com/src/img/branding/solanaLogoMark.svg'},
]
const dexList = [
  {name: 'Raydium', id: 1, val: 'raydium', image: Raydium},
  {name: 'PumpFun', id: 2, val: 'pumpfun', image: PumpFun},
]
const selectedDex = ref(dexList[0]);
const selectedValidator = ref(validators[1]);

const tokenName = computed(() => {
  if (selectedToken.value) {
    return `${selectedToken.value?.name} (${selectedToken.value?.symbol})`;
  } else {
    return '';
  }
})
const tokenMint = computed(() => {
  if (props.campaignAction === 'pull-up') {
    return 'dest_token_mint'
  } else {
    return 'source_token_mint'
  }
})
const label = computed(() => {
  if (props.campaignAction === 'pull-up') {
    return 'Select Token to pull up or enter mint address'
  } else {
    return 'Select Token to pull down or enter mint address'
  }
})
const handleTokenSelect = (token) => {
  isDropDownOpen.value.token = false;
  if (!token) return;

  selectedToken.value = token;
  campaignStore.campaign[tokenMint.value] = token.address;
  campaignStore.setSelectedToken(token);
}

const handleValidatorSelect = (validator) => {
  isDropDownOpen.value.validator = false;
  if (!validator) return;

  selectedValidator.value = validator;
  campaignStore.campaign.using_jito = validator.val === 'jito';
}

const handleDexSelect = (dex) => {
  isDropDownOpen.value.dex = false;
  if (!dex) return;

  selectedDex.value = dex;
  campaignStore.campaign.provider_id = dex.id;
}

watch(() => [campaignStore.campaign, tokensStore.solTokensData], (newVal) => {
  const tokenAddress = newVal[0][tokenMint.value];
  if (props.isEditMode && tokenAddress && newVal[1][tokenAddress]) {
    const token = newVal[1][tokenAddress];
    selectedToken.value = token;
    campaignStore.setSelectedToken(token);
  } else if (!tokenAddress && selectedToken.value) {
    selectedToken.value = null;
    campaignStore.setSelectedToken(null);
  }

  selectedDex.value = dexList.find(v => v.id === newVal[0].provider_id) || dexList[0];

  if (newVal[0].using_jito) {
    selectedValidator.value = validators[0];
  } else {
    selectedValidator.value = validators[1];
  }
}, {immediate: true, deep: true});

defineExpose({
  selectedDex,
})
</script>
<style scoped lang="scss">
.exchange-settings {
  &__inputs {
    margin-top: 20px;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;

    & .custom-option {
      display: flex;
      flex-direction: column;
      padding: 2px 0;
    }

    & .dex-image {
      aspect-ratio: 1/1;
      min-width: 20px;
      max-width: 20px;
      display: flex;
      border-radius: 50%;
      margin-right: 8px;
    }
  }

  &__input {
    width: calc(50% - 6px);

    & .icon {
      aspect-ratio: 1/1;
      min-width: 20px;
      min-height: 20px;
      max-width: 20px;
      max-height: 20px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #000;

     ::v-deep(img) {
       width: 70%;
       height: 70%
     }
    }
  }
}
</style>