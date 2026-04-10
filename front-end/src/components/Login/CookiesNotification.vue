<template>
  <div class="cookies">
    <h2 class="heading-4">We use cookies</h2>
    <p class="paragraph-small">We use cookies to enhance your experience and analyze website traffic.</p>
    <div class="cookies__btns">
      <UIButton color_type="ghost">Learn more</UIButton>
      <UIButton color_type="outline" @cta="handleBtnClick('decline')">Decline</UIButton>
      <UIButton color_type="primary" @cta="handleBtnClick('accept')">Accept all</UIButton>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../UI/UIButton.vue";

const emits = defineEmits(["closeCookies"]);

const handleBtnClick = (action) => {
  if (action === 'learn') {

  } else if (action === 'decline' || action === 'accept') {
    if (action === 'accept') {
      localStorage.setItem('is_cookies', JSON.stringify('accept-all'))
    } else {
      localStorage.setItem('is_cookies', JSON.stringify('decline'))
    }

    emits('closeCookies')
  }
}
</script>
<style scoped lang="scss">
.cookies {
  position: fixed;
  z-index: 10000;
  bottom: 20px;
  right: 20px;
  border-radius: 10px;
  border: 1px solid #E5E7EB;
  background: #FFF;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.10), 0 4px 6px -4px rgba(0, 0, 0, 0.10);
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  & p {
    font-weight: 400;
  }

  &__btns {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
  }
}

@media (max-width: 1200px) {
  .cookies {
    right: 0;
    bottom: 8px;
  }
}
</style>