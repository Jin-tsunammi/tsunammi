<template>
  <div class="profile-history">
    <div class="profile-history__top">
      <div class="profile-history__logo">
<!--        <img src="" alt="Image">-->
      </div>
      <span class="profile-history__title">Default campaign (DDP)</span>
    </div>

    <div class="profile-history__bottom">
      <div class="profile-history__ended">
        <div class="paragraph-small">Ended <span class="">22.10.2025</span></div>
      </div>

      <div class="profile-history__spend">
        <div class="paragraph-small label">Spend</div>
        <div class="monospaced-medium info"><span>9</span>/23 Sol</div>
        <div class="progress paragraph-small">
          34%
          <RoundedProgress
            :value="34"
            :show-text="false"
          />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>

import RoundedProgress from "../UI/RoundedProgress.vue";
</script>
<style scoped lang="scss">
.profile-history {
  display: flex;
  flex-direction: column;
  gap: 12px;

  &__title {
    color: #111827;
    font-size: 18px;
    font-style: normal;
    font-weight: 500;
    line-height: 150%;
  }

  &__top {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__logo {
    border-radius: 50%;
    min-width: 28px;
    max-width: 28px;
    aspect-ratio: 1/1;
    overflow: hidden;
    display: flex;
    background: #F3F4F6;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);
  }

  &__bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__ended {
    display: flex;
    align-items: center;

    & div {
      color: #030712;
      font-weight: 500;
    }

    & span {
      color: #6B7280;
      font-weight: 400;
    }
  }

  &__spend {
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
</style>