<template>
  <div class="selected-cex-top-up">
    <div class="selected-cex-top-up__content_label top heading-3">Top up wallets from CEX</div>
    <div class="selected-cex-top-up__inner">
      <div class="selected-cex-top-up__grid">
        <UIBaseInput
          class="block block-1"
          size="large"
          v-model="selectedToken"
          :is_readonly="true"
          label="Select Solana Network Token for Top up"
        >
          <template #icon-left>
            <div class="image">
              <img  src="../../../../public/images/solana-icon.webp" alt="">
            </div>
          </template>
        </UIBaseInput>

        <UISelect
          class="block block-2"
          size="large"
          v-model="dropDownStatus.cex"
          :selected="selectedCEX?.name || ''"
          label="Select CEX for Top up"
          :placeholder="cexInputPlaceholder"
          :error-message="fieldErrors.cex_api"
        >
          <template #dropdown>
            <UIDropdown
              :is-open="dropDownStatus.cex"
              :selected-option="selectedCEX?.name || ''"
              :options="cexAPI"
              label="name"
              @handle-option-select="handleProjectCEXSelect($event, 'cex')"
            >
              <template #custom-dropdown>
                <div class="custom-dropdown-content" @click.stop>
                  <div class="paragraph-small bold">No exchange connected.</div>
                  <p class="paragraph-mini regular">Connect your CEX API to enable wallet top-ups.</p>
                  <UIButton
                    color_type="ghost"
                    size="large"
                    @cta="openPage('cex')"
                  >
                    Connect CEX API

                    <template #left-icon>
                      <SVGPlus/>
                    </template>
                  </UIButton>
                </div>
              </template>
            </UIDropdown>
          </template>
        </UISelect>

        <UISelect
          class="block block-3"
          size="large"
          v-model="dropDownStatus.project"
          :placeholder="projectInputPlaceholder"
          :selected="selectedProject?.name || ''"
          label="Select Wallet Pool"
          :error-message="fieldErrors.project"
        >
          <template #dropdown>
            <UIDropdown
              :is-open="dropDownStatus.project"
              :selected-option="selectedProject?.name || ''"
              :options="projects"
              label="name"
              @handle-option-select="handleProjectCEXSelect($event, 'project')"
            >
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

        <div class="quantity-block numbers block block-4">
          <UIBaseInput
            size="large"
            v-model="cexData.quantity"
            label="Q-ty of wallets"
            type="number"
            @handle-input="handleWalletsQuantityInput"
            @handle-blur="handleWalletsQuantityBlur"
            :error-message="fieldErrors.wallets_qty"
          >
            <template #bottom-right>
              <UIGhostButtonsGroup
                :options="walletsOptions"
                :selected-option="quantityOptionSelected"
                @handle-option-select="handleWalletsQuantitySelect"
              />
            </template>
            <template v-if="selectedProject" #bottom-left>
              <span class="paragraph-mini medium grey">Total wallets: {{selectedProject.wallet_count}}</span>
            </template>
          </UIBaseInput>
        </div>
        <div class="numbers block-5">
          <UIBaseInput
            size="large"
            v-model="cexData.min_deposit"
            label="Min deposit per wallet"
            type="number"
            placeholder="Min: 0.016"
            :error-message="fieldErrors.min_deposit_amount"
            @handle-input="handleErrorClear('min_deposit_amount')"
            @handle-blur="handleInputBlur"
          >
            <template #icon-right>
              <span class="sol monospaced-small">Sol</span>
            </template>
          </UIBaseInput>
          <UIBaseInput
            size="large"
            v-model="cexData.max_deposit"
            label="Max deposit per wallet"
            type="number"
            :error-message="fieldErrors.max_deposit_amount"
            @handle-input="handleErrorClear('max_deposit_amount')"
          >
            <template #icon-right>
              <span class="sol monospaced-small">Sol</span>
            </template>
          </UIBaseInput>
        </div>
      </div>
      <div class="divider"></div>
      <div class="selected-cex-top-up__total">
        <div class="selected-cex-top-up__total_left">
          <span class="label paragraph-small">Sum for deposit</span>
          <div class="sum monospaced-large">{{totalAmountText}} Sol</div>
          <span class="commission paragraph-mini">Total (Including Commission): {{totalCommission}} Sol</span>
        </div>
        <UIButton color_type="primary" size="large" @cta="handleTopUp">Top up and pay</UIButton>
      </div>
    </div>
  </div>
</template>
<script setup>
import UISelect from "../../UI/UISelect.vue";
import UIBaseInput from "../../UI/UIBaseInput.vue";
import UIButton from "../../UI/UIButton.vue";
import {computed, ref} from "vue";
import UIDropdown from "../../UI/UIDropdown.vue";
import SVGPlus from "../../SVG/SVGPlus.vue";
import {useRouter} from "vue-router";
import UIGhostButtonsGroup from "../../UI/UIGhostButtonsGroup.vue";
import {useToastStore} from "../../../store/toastStore.js";
import {CreateDeposit} from "../../../api/api.js";
import {formatText, toDynamicFix} from "../../../helpers/index.js";

const props = defineProps({
  projects: {type: Array, default: []},
  cexAPI: {type: Array, default: []},
})
const emits = defineEmits(['openModal'])
const toastStore = useToastStore();
const router = useRouter();
const COMMISSION = 0.008;
const  walletsOptions = [
  {label: 'half', val: 'half'},
  {label: 'all', val: 'all'}
]

const selectedProject = ref(null);
const selectedCEX = ref(null);
const selectedToken = ref('SOL'); //currently only SOL
const fieldErrors = ref({
  cex_api: '',
  project: '',
  wallets_qty: '',
  min_deposit_amount: '',
  max_deposit_amount: ''
})
const dropDownStatus = ref({
  cex: false,
  project: false,
})
const cexData = ref({
  quantity: 0,
  min_deposit: null,
  max_deposit: null,
})
const quantityOptionSelected = ref('') // half | all
const totalAmount = computed(() => {
  let averageAmount  = 0;

  if (cexData.value.min_deposit && cexData.value.max_deposit) {
    averageAmount = (cexData.value.min_deposit + cexData.value.max_deposit) / 2;
  } else {
    averageAmount = cexData.value.min_deposit + cexData.value.max_deposit
  }
  return cexData.value.quantity * averageAmount || 0;
})
const totalAmountText = computed(() => {
  return !totalAmount.value ? 0 : `~${toDynamicFix(totalAmount.value)}`;
})
const totalCommission = computed(() => {
  const sum = totalAmount.value + (COMMISSION * (cexData.value.max_deposit ? cexData.value.quantity : 0));
  return !sum ? 0 : `~${toDynamicFix(sum)}`;
})
const cexInputPlaceholder = computed(() => {
  if (!props.cexAPI.length) return 'No CEX connected'
  else return 'Select CEX'
})
const projectInputPlaceholder = computed(() => {
  if (!props.projects.length) return 'No Projects created'
  else return 'Select Project'
})

function handleWalletsQuantityInput(event) {
  if (!selectedProject.value || !selectedProject.value.wallet_count) return;
  const raw = String(event.target.value);
  const cleaned = raw.replace(/\D+/g, '');
  const val = Number(cleaned || 0);
  cexData.value.quantity = val;

  const totalWallets = selectedProject.value.wallet_count;
  const half = Math.floor(selectedProject.value.wallet_count / 2);

  if (isNaN(val) || val < 0) {
    cexData.value.quantity = 0;
    return;
  }

  if (Number(val) > totalWallets) {
    cexData.value.quantity = totalWallets;
    quantityOptionSelected.value = 'all';
  } else if (Number(val) === half) {
    quantityOptionSelected.value = 'half';
  } else {
    quantityOptionSelected.value = '';
  }

  handleErrorClear('wallets_qty');
}

const handleWalletsQuantityBlur = (event) => {
  const val = event.target.value;
  if (!val || !selectedProject.value || !selectedProject.value?.wallet_count) {
    cexData.value.quantity = 0;
  } else if (Number(val) > selectedProject.value.wallet_count) {
    cexData.value.quantity = selectedProject.value.wallet_count;
  }
}

const openPage = (type) => {
  if (type === 'project') {
    dropDownStatus.value.project = false;
    router.push({name: 'WalletsProjects'});
  } else if (type === 'cex') {
    dropDownStatus.value.cex = false;
    router.push({name: 'WalletsConnectCexApi'});
  }
}
const handleWalletsQuantitySelect = (type) => {
  if (!selectedProject.value) {
    return 0;
  }

  if (type === 'all') {
    cexData.value.quantity = selectedProject.value.wallet_count;
  } else if (type === 'half') {
    cexData.value.quantity = Math.floor(selectedProject.value.wallet_count / 2);
  }

  quantityOptionSelected.value = type;
}

const handleInputBlur = () => {
  if (cexData.value.min_deposit < 0.016) {
    cexData.value.min_deposit = 0.016;
  }
}

const handleProjectCEXSelect = (data, type='') => {
  if (!data) return;

  if (type === 'project') {
    selectedProject.value = data;
    dropDownStatus.value.project = false;
    handleErrorClear('project')
  } else if (type === 'cex') {
    dropDownStatus.value.cex = false;
    selectedCEX.value = data;
    handleErrorClear('cex_api')
  }
}

const handleErrorClear = (field) => {
  fieldErrors.value[field] = '';
}

function isFieldsError() {
  const emptyError = 'Field is required'

  if (!selectedCEX.value) {
    fieldErrors.value.cex_api = emptyError;
  }

  if (!selectedProject.value) {
    fieldErrors.value.project = emptyError;
  }

  if (!cexData.value.min_deposit) {
    fieldErrors.value.min_deposit_amount = emptyError;
  }

  if (cexData.value.min_deposit && cexData.value.min_deposit < 0.016) {
    fieldErrors.value.min_deposit_amount = 'Min is 0.016';
  }

  if (cexData.value.max_deposit && cexData.value.max_deposit < cexData.value.min_deposit) {
    fieldErrors.value.max_deposit_amount = 'Max should be greater than Min';
  }

  if (!cexData.value.quantity) {
    fieldErrors.value.wallets_qty = 'Add at least one wallet';
  }

  return Object.keys(fieldErrors.value).some(k => fieldErrors.value[k]);
}

const handleTopUp = async() => {
  if (isFieldsError()) return;

  try {
    const data = {
      account_id: selectedCEX.value.id,
      min_amount: Number(cexData.value.min_deposit),
      max_amount: Number(cexData.value.max_deposit) ? Number(cexData.value.max_deposit) : Number(cexData.value.min_deposit),
      project_id: selectedProject.value.id,
      quantity: Number(cexData.value.quantity)
    }
    const resp = await CreateDeposit(data);

    emits("openModal", {type: 'confirmation', order_id: resp?.data?.order_id})

  } catch (e) {
    console.error(e);
    toastStore.error({text: formatText(e?.response?.data), timeout: 4000})
  }
}
</script>
<style scoped lang="scss">
.selected-cex-top-up {
  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: 100%;

    &_label {
      position: relative;
      color: #030712;

      &.top {
        max-width: 668px;
      }

      &.history {
        margin-top: 32px;
      }

      &::after {
        content: '';
        position: absolute;
        bottom: -12px;
        left: 0;
        width: 100%;
        height: 1px;
        background: #030712;
      }
    }
  }

  &__inner {
    margin-top: 32px;
    display: flex;
    flex-direction: column;
    max-width: 668px;

    & .divider {
      display: flex;
      height: 1px;
      width: 100%;
      max-width: 668px;
      margin: 20px 0;
      background: #D1D5DB;
    }

    &_block {
      display: flex;
      gap: 12px;

      & .sol {
        color: #6B7280;
      }

      &.numbers {
        ::v-deep(.base-input__input) {
          & input {
            font-family: "Geist Mono", sans-serif;
            font-size: 14px;
            font-style: normal;
            font-weight: 400;
            line-height: 150%;
          }
        }
      }

      & .ui-select {
        width: calc(50% - 6px);
      }

      & .base-input {
        width: calc(50% - 6px);
      }

      & .quantity-block {
        width: calc(50% - 6px);
        display: flex;
        flex-direction: column;
        gap: 4px;

        & button {
          background: transparent;
          font-weight: 500;
          height: 24px;
          padding: 0 8px;
        }

        & .buttons {
          align-self: flex-end;
          display: flex;
          align-items: center;
        }

        & .base-input {
          width: 100%;
        }
      }
    }
  }

  &__grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;

    & .block {
      height: fit-content;
    }

    & .block-1 { grid-area: 1 / 1 / 2 / 2; }
    & .block-2 { grid-area: 1 / 2 / 2 / 3; }
    & .block-3 { grid-area: 2 / 1 / 3 / 2; }
    & .block-4 { grid-area: 3 / 1 / 4 / 2; }
    & .block-5 { grid-area: 3 / 2 / 4 / 3; }

    & .image {
      min-width: 20px;
      min-height: 20px;
      max-height: 20px;
      max-width: 20px;
      margin-right: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #000;
      border-radius: 50%;

      & img {
        width: 70%;
        height: 70%;
      }
    }

    & .sol {
      color: #6B7280;
    }

    & .block-5 {
      display: flex;
      gap: 12px;
    }

    & .numbers {
      position: relative;
      ::v-deep(.base-input__input) {
        & input {
          font-family: "Geist Mono", sans-serif;
          font-size: 14px;
          font-style: normal;
          font-weight: 400;
          line-height: 150%;
        }
      }
    }

    & .quantity-block {
      display: flex;
      flex-direction: column;
      gap: 4px;

      & button {
        background: transparent;
        font-weight: 500;
        height: 24px;
        padding: 0 8px;
        color: #4B5563;
        transition: .3s ease;

        &:hover, &.active {
          color: #030712;
        }
      }

      & .buttons {
        position: absolute;
        bottom: 0;
        right: 0;
        z-index: 2;
        align-self: flex-end;
        display: flex;
        align-items: center;
      }

      & .base-input {
        width: 100%;
      }
    }
  }

  &__total {
    border-radius: 8px;
    border: 1px solid #D1D5DB;
    background: #F9FAFB;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.12);
    padding: 12px;
    display: flex;
    align-items: flex-end;
    justify-content: space-between;

    & .label {
      display: block;
      margin-bottom: 8px;
      color: #030712;
    }

    & .sum {
      color: #030712;
      font-weight: 600;
    }

    & .commission {
      color: #6B7280;
      font-weight: 400;
    }
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


@media (max-width: 1200px) {
  .selected-cex-top-up {
    &__content {
      width: 100%;
      display: flex;
      flex-direction: column;
      height: 100%;

      &_label {
        display: none;
      }
    }

    &__inner {
      margin-top: 20px;
    }

    &__grid {
      grid-template-columns: 1fr;
      grid-template-rows: auto;

      & .block-1 { grid-area: 1 / 1 / 2 / 2; }
      & .block-2 { grid-area: 2 / 1 / 3 / 2; }
      & .block-3 { grid-area: 3 / 1 / 4 / 2; }
      & .block-4 { grid-area: 4 / 1 / 5 / 2; }
      & .block-5 { grid-area: 5 / 1 / 6 / 2; }
    }
  }
}
</style>