<template>
  <div class="login-form-base">
    <h2 class="login-form-base__title paragraph-medium">{{text.title}}</h2>
    <span class="login-form-base__text paragraph-small">{{text.info}}</span>
    <UIBaseInput
      v-model="vModel"
      label="Email"
      size="large"
      :error-message="emailError"
    />
    <div v-if="route.name === 'SignUp'" class="login-form-base__agreement">
      <UICheckBox
        v-model="isAgreed"
        label="I agree to the Terms of Service and Privacy Policy"
      />
    </div>
    <div class="login-form-base__btns">
      <UIButton
        size="large"
        :is_disabled="isSignUpDisabled"
        @cta="handleEmailSend"
      >
        {{text.email_btn}}
      </UIButton>
      <UIButton
        size="large"
        color_type="outline"
        @cta="handleGoogleAuth"
      >
        {{text.google_btn}}
        <template #left-icon>
          <SVGGoogleIcon />
        </template>
      </UIButton>
      <div class="divider paragraph-small regular">or</div>
      <UIButton
        size="large"
        color_type="outline"
        @cta=""
      >
        {{text.wallet_btn}}
        <template #left-icon>
          <img class="wallet-icon" src="../../../public/images/connect-wallet-image.webp" alt="Image">
        </template>
      </UIButton>
    </div>
    <div class="login-form-base__bottom paragraph-small">
      {{text.bottom_text}}
      <button @click="emits('handlePageChange')">{{text.link_text}}</button>
    </div>
  </div>
</template>
<script setup>
import {useRoute, useRouter} from "vue-router";
import {computed, ref, watch} from "vue";
import UIButton from "../UI/UIButton.vue";
import UIBaseInput from "../UI/UIBaseInput.vue";
import UICheckBox from "../UI/UICheckBox.vue";
import {GoogleAuthProvider, signInWithPopup} from "firebase/auth";
import CookieManager from "../../helpers/cookieManager.js";
import {useToken} from "../../composable/useToken.js";
import auth from "../../firebase/index.js";
import {SignInByGoogle, SignUpByGoogle} from "../../api/api.js";
import {useToastStore} from "../../store/toastStore.js";
import {useUserStore} from "../../store/userStore.js";
import {errorToast} from "../../helpers/index.js";
import SVGGoogleIcon from "../SVG/SVGGoogleIcon.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import ConnectWallet from "../Profile/ConnectWallet.vue";

const props = defineProps({
  isEmailAlreadyExists: {type: Boolean, default: false},
  page: {type: String, default: 'login'}
})
const emits = defineEmits(['handleCodeRequest', 'handlePageChange'])

const route = useRoute();
const toastStore = useToastStore();
const userStore = useUserStore();
const modalStore = useModalsStore();
const vModel = defineModel({type: String, default: ''});
const {setToken} = useToken();
const isAgreed = ref(false);
const emailError = ref('');
const isSignUpDisabled = computed(() => {
  if (route.name === "SignUp") {
    return !isAgreed.value;
  } else {
    return false;
  }
})
const text = computed(() => {
  if (props.page === "login") {
    return {
      title: 'Login to your account',
      info: 'Enter your email below to login to your account',
      email_btn: 'Login',
      google_btn: 'Login with Google',
      bottom_text: 'Don\'t have an account?',
      link_text: 'Sign up',
      wallet_btn: 'Wallet connect'
    }
  } else {
    return {
      title: 'Create your account',
      info: 'Enter your email below to create a new account',
      email_btn: 'Sign up',
      google_btn: 'Sign up with Google',
      bottom_text: 'Already have an account?',
      link_text: 'Log in',
      wallet_btn: 'Wallet connect'
    };
  }
})

const handleEmailSend = () => {
  const emailPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

  if (!emailPattern.test(vModel.value)) {
    emailError.value = 'Invalid email';
    return;
  } else {
    emailError.value = '';
  }

  emits('handleCodeRequest');
}

const handleGoogleAuth = async() => {
  try {
    // in case if need to add other auth providers like Facebook, Apple, etc
    let provider;

    provider = new GoogleAuthProvider();
    provider.setCustomParameters({
      prompt: 'select_account'
    });

    const { user } = await signInWithPopup(auth, provider);
    const {accessToken} = user.stsTokenManager;

    CookieManager.setItem(
      "isOAuth",
      '',
      user.stsTokenManager.expirationTime
    );

    let resp = null;

    if (props.page === 'signup') {
      resp = await SignUpByGoogle(accessToken);
    } else {
      resp = await SignInByGoogle(accessToken)
    }

    if (resp.data) {
      userStore.setUserData(resp.data.user);
      setToken(resp.data.jwt_info?.access_token, resp.data.jwt_info?.refresh_token);
      localStorage.removeItem('connected_wallets');
      userStore.isUserAuth = true;
      modalStore.closeModal();
    }
  } catch(error) {
    console.error("Error during Google sign-in:", error)
    if (error.response?.data === 'user with email already exists') {
      toastStore.error({text: 'User with this email already exists'});
    } else if (error.response?.data === 'user with email not found') {
      toastStore.error({text: 'This email does not register'});
    } else {
      errorToast(error.response.data);
    }
    return false;
  }
}

watch(() => props.isEmailAlreadyExists, (newVal) => {
  if (newVal) {
    emailError.value = 'This email already exists.'
  }
})
</script>
<style scoped lang="scss">
.login-form-base {
  display: flex;
  flex-direction: column;
  align-items: center;

  &__title {
    color: #030712;
    font-weight: 500;
  }

  &__text {
    color: #6B7280;
    display: block;
    margin-top: 4px;
    margin-bottom: 24px;
    font-weight: 400;
  }

  &__agreement {
    margin-top: 28px;
  }

  &__btns {
    margin-top: 28px;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 12px;

    & .wallet-icon {
      width: 15px;
      height: 15px;
    }

    & .divider {
      width: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 10px;

      &::before, &::after {
        content: "";
        height: 1px;
        width: 100%;
        background: #E5E7EB;
      }
    }
  }

  &__bottom {
    margin-top: 28px;
    font-weight: 400;
    color: #030712;
    align-self: center;

    & button {
      text-decoration: underline;
      background: none;
    }
  }
}

@media (max-width: 1200px) {
  .login-form-base {
    &__text {
      text-align: center;
    }
  }
}
</style>