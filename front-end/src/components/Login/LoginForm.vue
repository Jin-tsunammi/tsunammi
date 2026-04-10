<template>
  <div class="login-form">
    <BaseStage
      v-show="stage === 'base'"
      @handle-code-request="handleCodeRequest"
      @handle-page-change="changePage"
      v-model="email"
      :is-email-already-exists="isEmailAlreadyExists"
      :page="page"
    />
    <CodeStage
      v-if="stage === 'code'"
      :email="email"
      @handle-stage-change="handleStageChange"
      @handle-code-request="handleCodeRequest"
      :page="page"
    />
  </div>
</template>
<script setup>
import {ref} from "vue";
import BaseStage from "./BaseStage.vue";
import CodeStage from "./CodeStage.vue";
import {CheckEmail, GetCodeByEmail} from "../../api/api.js";
import {useRoute} from "vue-router";

const page = ref('login'); // login | signup
const stage = ref('base'); // base | code
const email = ref('');
const isEmailAlreadyExists = ref(false);

const handleStageChange = (val) => {
  stage.value = val;
}

const handleCodeRequest = async() => {
  isEmailAlreadyExists.value = false;
  try {
    if (page.value === 'signup') {
      const resp = await CheckEmail({email: email.value});

      if (resp?.data?.exist) {
        isEmailAlreadyExists.value = true;
        return;
      }
    }
    await GetCodeByEmail({email: email.value});
    handleStageChange('code');
  } catch (e) {
    console.error(e);
  }
}

const changePage = () => {
  if (page.value === 'login') {
    page.value = 'signup';
  } else {
    page.value = 'login';
  }
}
</script>
<style scoped lang="scss">
.login-form {
  display: flex;
  flex-direction: column;
  padding: 24px 40px;
  border-radius: 12px;
  background: #FFF;
  border: 1px solid #E5E7EB;
  align-items: center;
}

@media (max-width: 1200px) {
  .login-form {
    padding: 24px;
  }
}
</style>