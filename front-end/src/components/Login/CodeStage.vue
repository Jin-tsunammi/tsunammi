<template>
  <div class="login-form-code">
    <h2 v-if="!route.path.includes('/dashboard/')" class="login-form-code__title paragraph-medium">Login to your account</h2>
    <div class="login-form-code__code">
      <div class="top">
        <span class="paragraph-small">{{codeLabel}}</span>
        <button class="reload paragraph-small" :disabled="seconds > 0" @click="handleResendCode">
          <SVGReload v-if="seconds === 0" />
          {{resendText}}
        </button>
      </div>
      <div class="login-form-code__inputs">
        <v-otp-input
          :class="['otp-class', {'error': code.isError}]"
          :num-inputs="6"
          :should-auto-focus="true"
          @on-complete="handleCodeSubmit"
          v-model:value="code.val"
          @on-change="removeCodeError"
        />
      </div>
      <span v-if="code.isError" class="error paragraph-mini">Wrong code</span>
    </div>
    <div v-if="route.name !== 'DashboardProfile'" class="login-form-code__next">
      <UIButton
        size="large"
        @cta="handleCodeSubmit"
      >
        Next
      </UIButton>
    </div>
    <button v-if="route.name !== 'DashboardProfile'" class="login-form-code__return" @click="changeEmail">
      <span class="return paragraph-small">
        <SVGArrowPrevious />
        Back to change
      </span>
      <span class="email paragraph-small">{{email}}</span>
    </button>
  </div>
</template>
<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from "vue";
import UIButton from "../UI/UIButton.vue";
import VOtpInput from "vue3-otp-input"
import SVGReload from "../SVG/SVGReload.vue";
import SVGArrowPrevious from "../SVG/SVGArrowPrevious.vue";
import {ChangeUserEmail, SignInByEmail, SignUpByEmail} from "../../api/api.js";
import {useToken} from "../../composable/useToken.js";
import {useRoute, useRouter} from "vue-router";
import {useToastStore} from "../../store/toastStore.js";
import {useUserStore} from "../../store/userStore.js";
import {useModalsStore} from "../../store/modalsStore.js";
import {formatText} from "../../helpers/index.js";

const props = defineProps({
  email: {type: String, required: true},
  page: {type: String, default: 'login'}
})
const emits = defineEmits(['handleCodeRequest', 'handleStageChange', 'update:code'])
const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const toastStore = useToastStore();
const modalStore = useModalsStore();

let timer = null;
const {setToken} = useToken();
const seconds = ref(30);
const code = ref({
  val: '',
  isError: false,
});
const resendText = computed(() => {
  if (seconds.value > 0) {
    return `Resend in ${seconds.value}s`
  } else {
    return 'Resend'
  }
})
const codeLabel = computed(() => {
  if (route.name === 'DashboardProfile') {
    return 'Verification code'
  } else {
    return 'One time pass'
  }
})

const countDown = () => {
  if (seconds.value === 0) {
    clearInterval(timer);

    return;
  }

  seconds.value--;
}

const changeEmail = () => {
  emits('handleStageChange', 'base')
}

const handleResendCode = async() => {
  clearInterval(timer);
  seconds.value = 30;
  timer = setInterval(countDown, 1000);
  emits('handleCodeRequest')
}

const handleCodeSubmit = async() => {
  const data = {email: props.email, code: code.value.val};
  try {
    if (!code.value.val) {
      code.value.isError = true;

      return;
    }

    let codeResp = null;

    if (props.page === 'signup') {
      codeResp = await SignUpByEmail(data);
    } else if (props.page === 'login') {
      codeResp = await SignInByEmail(data)
    } else if (route.name === 'DashboardProfile') {
      codeResp = await ChangeUserEmail(data)
    }

    if (codeResp && codeResp?.data) {
      if (route.name !== 'DashboardProfile') {
        userStore.setUserData(codeResp.data.user);
        setToken(codeResp.data.jwt_info?.access_token, codeResp.data.jwt_info?.refresh_token);
        userStore.isUserAuth = true;
      } else {
        await userStore.getUserData();
      }
      modalStore.closeModal();
      localStorage.removeItem('connected_wallets');
    }

  } catch (e) {
    console.error('code error', e);
    if (e.response?.data === 'code is invalid') {
      code.value.isError = true;
    } else {
      toastStore.error({text: formatText(e.response.data)});
    }
  }
}

const removeCodeError = () => {
  if (code.value.isError) {
    code.value.isError = false;
  }
}

watch(() => code.value.val, (newVal) => {
  emits('update:code', newVal);
})

onMounted(() => {
  timer = setInterval(countDown, 1000);
})

onBeforeUnmount(() => {
  clearInterval(timer)
})

defineExpose({
  handleCodeSubmit,
})
</script>
<style scoped lang="scss">
.login-form-code {
  display: flex;
  flex-direction: column;
  align-items: center;

  &__code {
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 6px;

    & .top {
      display: flex;
      align-items: center;
      justify-content: space-between;

      & span {
        font-weight: 500;
      }
    }

    & .reload {
      font-weight: 500;
      background: transparent;
      display: flex;
      align-items: center;
      gap: 8px;
      color: #111827;

      &:disabled {
        color: #6B7280;
        cursor: not-allowed;
      }
    }

    & .error {
      font-weight: 400;
      color: #DC2626;
    }
  }

  &__next {
    margin-top: 28px;
  }

  &__inputs {
    ::v-deep(.otp-class) {
      gap: 6px;

      & input {
        width: 100%;
        height: 55px;
        background: transparent;
        border: 1px solid #E5E7EB;
        border-radius: 8px;
        text-align: center;

        color: #030712;

        font-size: 14px;
        font-style: normal;
        font-weight: 400;
        line-height: 150%;
        letter-spacing: 0.07px;
      }
    }

    ::v-deep(.otp-class.error) {
      & input {
        border: 1px solid #EF4444;
      }
    }
  }

  &__return {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
    background: transparent;
    overflow: hidden;
    width: 100%;

    & span {
      color: #030712;
      font-weight: 400;
      display: flex;
      align-items: center;
    }

    & .return {
      width: max-content;
    }

    & .email {
      display: block;
      font-weight: 500;
      margin-left: 3px;
      text-overflow: ellipsis;
      overflow: hidden;
      max-width: 160px;
    }

    & svg {
      margin-right: 8px;
    }
  }
}
</style>