<template>
  <div class="empty-page__wrapper">
    <div class="empty-page">
      <img src="../../../../public/images/tools.webp" alt="Icon">
      <h1 >We’re working on this feature.</h1>
      <p class="paragraph-medium">This feature is currently under development and will be available soon.</p>
    </div>
    <span class="text paragraph-medium">Want updates? Follow us on <a href="https://x.com/MMakingtools"></a><SVGTwitter color="#030712"/></span>
  </div>
</template>
<script setup>

import SVGTwitter from "../../SVG/SVGTwitter.vue";
</script>
<style scoped lang="scss">
.empty-page {
  height: 259px;
  width: 100%;
  max-width: 545px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  border: 1px solid #E5E7EB;
  background: #FFF;
  padding: 0 60px;

  &__wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 36px;
  }

  & img {
    width: 30px;
    height: 30px;
    margin-bottom: 20px;
  }

  & h1 {
    color: #030712;
    margin-bottom: 12px;

    font-size: 24px;
    font-style: normal;
    font-weight: 600;
    line-height: 120%;
    letter-spacing: -0.48px;
  }

  & p {
    color: #9CA3AF;
    font-weight: 400;
    text-align: center;
  }

  & .text {
    color: #6B7280;
    font-weight: 400;
    display: flex;
  }
}

@media (max-width: 1200px) {
  .empty-page {
    padding: 15px 40px;

    & h1 {
      font-size: 20px;
      font-style: normal;
      font-weight: 600;
      line-height: 120%;
      letter-spacing: -0.4px;
    }
  }
}
</style>