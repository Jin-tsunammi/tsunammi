<template>
  <div class="add-api">
    <div class="add-api__left">
      <div class="add-api__title heading-4">
        {{modalsStore.modalData.title}}
        <SVGClose @click="modalsStore.closeModal" />
      </div>
      <div class="add-api__top">
        <div class="add-api__block">
          <UIBaseInput
            v-model="apiData.name"
            label="Your title for CEX"
            placeholder=""
            size="large"
          />
        </div>

        <div class="add-api__block">
          <UIBaseInput
            v-model="selectedCEX"
            placeholder=""
            label="CEX"
            size="large"
            :is_readonly="true"
          />
        </div>
      </div>

      <div class="add-api__divider"></div>

      <div class="add-api__bottom">
        <div class="add-api__bottom_label paragraph-small">Data from CEX</div>
        <div class="add-api__block">
          <UIBaseInput
            v-model="apiData.api_key"
            placeholder=""
            label="API Key"
            size="large"
            :type="isPasswordVisible ? 'text' : 'password'"
          >
            <template #icon-right>
              <button class="password" @click="isPasswordVisible = !isPasswordVisible">
                <SVGEyeCrossed v-if="!isPasswordVisible"/>
                <SVGEyeOpen v-else/>
              </button>
            </template>
          </UIBaseInput>
        </div>

        <div class="add-api__block">
          <UIBaseInput
            v-model="apiData.secret_key"
            placeholder=""
            label="API secret key"
            size="large"
            :type="isPasswordVisible ? 'text' : 'password'"
          >
            <template #icon-right>
              <button class="password" @click="isPasswordVisible = !isPasswordVisible">
                <SVGEyeCrossed v-if="!isPasswordVisible"/>
                <SVGEyeOpen v-else/>
              </button>
            </template>
          </UIBaseInput>
        </div>

        <div class="add-api__block">
          <UIBaseInput
            v-model="apiData.passphrase"
            placeholder=""
            label="API passphrase"
            size="large"
            :type="isPasswordVisible ? 'text' : 'password'"
          >
            <template #icon-right>
              <button class="password" @click="isPasswordVisible = !isPasswordVisible">
                <SVGEyeCrossed v-if="!isPasswordVisible"/>
                <SVGEyeOpen v-else/>
              </button>
            </template>
          </UIBaseInput>
        </div>

      </div>

      <div :class="['add-api__btns', {disabled: isCreating}]">
        <UIButton
          color_type="outline"
          @cta="modalsStore.closeModal"
        >
          Cancel
        </UIButton>
        <UIButton color_type="primary" @cta="handleImport" :is_disabled="isImportBtnAvailable">
          <template v-if="isCreating" #left-icon>
            <UISpinner/>
          </template>
          {{isCreating ? 'Importing...' : 'Import'}}
        </UIButton>
      </div>
    </div>

    <div class="add-api__right">
      <div class="add-api__right_title heading-4">How to connect your trading account</div>
      <div class="add-api__server-api">
        <span class="add-api__server-api_label paragraph-small medium">IP what you need to add:</span>
        <div class="add-api__server-api_input">
          <span class="paragraph-small regular">{{server_ip}}</span>
          <UICopyText :copy-text="server_ip" />
        </div>
      </div>
      <div class="add-api__right_scroll">
        <div class="add-api__steps">
          <div
            v-for="(step, index) in instructions"
            :key="index"
            class="add-api__step"
          >
            <div class="add-api__step_title heading-5">{{step.label}}</div>
            <div class="add-api__step_content">
              <p v-if="step.text1" class="paragraph-small regular grey">{{step.text1}}</p>
              <div
                v-if="step.lists"
                class=""
                v-for="(list, i) in step.lists"
                :key="i"
              >
                <p class="paragraph-small regular grey">{{list.text}}</p>
                <ul
                  v-for="(order, idx) in list.orders"
                  :key="idx"
                >
                  <li class="paragraph-small regular grey">{{order}}</li>
                </ul>
                <p v-if="list.text2" class="paragraph-small regular grey">{{list.text2}}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="add-api__right_image">
          <img src="../../../../public/images/kucoin_card.webp" alt="Kucoin">
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../../UI/UIButton.vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import UIBaseInput from "../../UI/UIBaseInput.vue";
import {computed, ref} from "vue";
import UISelect from "../../UI/UISelect.vue";
import SVGEyeOpen from "../../SVG/SVGEyeOpen.vue";
import SVGEyeCrossed from "../../SVG/SVGEyeCrossed.vue";
import {useCEXApiStore} from "../../../store/cexStore.js";
import UISpinner from "../../UI/UISpinner.vue";
import SVGClose from "../../SVG/SVGClose.vue";
import UICopyText from "../../UI/UICopyText.vue";

defineProps({
  server_ip: {type: String, default: ''},
})

const modalsStore = useModalsStore();
const cexApiStore = useCEXApiStore();
const isPasswordVisible = ref(false)
const isCreating = ref(false);

const AVAILABLE_CEX = [
  {name: 'Kucoin', exchange_id: 1}
]

const apiData = ref({
  name: '',
  exchange_id: AVAILABLE_CEX[0].exchange_id,
  api_key: '',
  passphrase: '',
  secret_key: '',
})
const selectedCEX = computed(() => {
  const cex = AVAILABLE_CEX.find(e => e.exchange_id === apiData.value.exchange_id);

  return cex.name || '';
})
const isImportBtnAvailable = computed(() => {
  return !apiData.value.name.length ||
    !apiData.value.api_key.length ||
    !apiData.value.passphrase.length ||
    !apiData.value.secret_key.length;
})
const instructions = [
  {
    label: "Step 1. Log in",
    text1: "Log in to your KuCoin account and open API Management from your profile menu.",

  },
  {
    label: "Step 2. Create API Key",
    text1: "Click Create API and choose Create API Key.",
    lists: [
      {
        text: "Fill in:",
        orders: [
          "API Name – any name (for example: Trading Bot)",
          "Passphrase – create and save it (you will need it together with API Key & Secret)"
        ]
      },
      {
        text: "Set permissions:",
        orders: [
          "Enable General",
          "Enable Withdraw",
        ]
      },
    ],
  },
  {
    label: "Step 3. Save your credentials",
    lists: [
      {
        text: "After creation, KuCoin will show:",
        orders: [
          "API Key",
          "Secret Key",
          "Passphrase",
        ],
        text2: "Copy all three and store them safely — Secret and Passphrase are shown only once."
      }
    ],
  },
  {
    label: "Step 5. Connect your account",
    lists: [
      {
        text: "Return to this page and fill in:",
        orders: [
          "API Key",
          "API secret key",
          "API passphrase",
        ],
        text2: "Then click \"Import\"."
      }
    ],
  },
]

const handleImport = async() => {
  try {
    isCreating.value = true;
    await cexApiStore.createNewCEXApi(apiData.value);
  } catch (e) {
  } finally {
    isCreating.value = false;
  }
}
</script>
<style scoped lang="scss">
.add-api {
  //width: max-content;
  display: flex;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid #E5E7EB;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.10), 0 4px 6px -4px rgba(0, 0, 0, 0.10);
  max-height: 637px;
  width: 808px;

  &__left {
    max-width: 404px;
    width: 100%;
    display: flex;
    flex-direction: column;
    padding: 16px 24px;
    background: #FFF;
  }

  &__right {
    max-width: 404px;
    width: 100%;
    display: flex;
    flex-direction: column;
    padding: 16px 24px;
    background: #FAFAFA;
    border-left: 1px solid #E5E7EB;
    overflow: scroll;

    &_title {
      margin-bottom: 40px;
    }

    &_image {
      margin-top: 24px;
      border-radius: 12px;
      overflow: hidden;
      width: auto;
      height: 188px;
    }

    &_scroll {
      display: flex;
      flex-direction: column;
      height: fit-content;
    }
  }

  &__server-api {
    display: flex;
    flex-direction: column;
    gap: 6px;

    &_input {
      height: 40px;
      border-radius: 8px;
      border: 1px solid #E5E7EB;
      background: #FFF;
      box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
      padding: 0 16px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  &__steps {
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  &__step {
    &_title {
      margin-bottom: 8px;
    }

    &_content {
      display: flex;
      flex-direction: column;
      gap: 12px;

      & ul {
        padding-left: 8px;
        list-style: inside;
      }
    }
  }

  &__title {
    margin-bottom: 40px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__top {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  &__bottom {
    display: flex;
    flex-direction: column;
    gap: 12px;

    &_label {
      color: #6B7280;
      font-weight: 500;
    }
  }

  &__divider {
    margin: 24px 0;
    width: 100%;
    height: 1px;
    background: #E5E7EB;
  }

  &__block {
    & .password {
      width: 20px;
      height: 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: transparent;

      & svg {
        width: 100%;
        height: 100%;
      }
    }
  }

  &__btns {
    margin-top: 40px;
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
</style>