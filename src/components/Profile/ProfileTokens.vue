<template>
  <div class="profile-tokens">
    <div class="profile-tokens__top">
      <div class="profile-tokens__info">
        <div class="profile-tokens__avatar">
<!--          <img src="" alt="Image">-->
        </div>
        <span class="paragraph-medium">Duel Duck</span>
        <span class="paragraph-medium">DDP</span>
      </div>

      <div class="profile-tokens__details">
        <span class="monospaced-medium">$0.11</span>
        <UIBadge type="round" color_type="success">+31%</UIBadge>
      </div>
    </div>
    <div class="profile-tokens__cells">
      <div class="profile-tokens__cell">
        <span class="label paragraph-small">Amount</span>
        <div class="value paragraph-medium">3 232 332</div>
      </div>
      <div class="profile-tokens__cell">
        <span class="label paragraph-small">Your share</span>
        <div class="value paragraph-medium">21%</div>
      </div>
      <div class="profile-tokens__cell">
        <span class="label paragraph-small">MCap</span>
        <div class="value paragraph-medium">$312,5K</div>
      </div>
    </div>
    <UIButton color_type="primary" size="large">Create</UIButton>
  </div>
</template>
<script setup>

import UIButton from "../UI/UIButton.vue";
import UIBadge from "../UI/UIBadge.vue";
</script>
<style scoped lang="scss">
.profile-tokens {
  border-radius: 12px;
  background: #FFF;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 20px;

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  &__cells {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__cell {
    display: flex;
    flex-direction: column;
    height: 69px;
    border-radius: 8px;
    padding: 12px;
    border: 1px solid #D1D5DB;
    background: #F9FAFB;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.12);

    &:nth-child(1) {
      width: 40%;
    }

    &:nth-child(2), &:nth-child(3) {
      width: 30%;
    }

    & .label {
      color: #6B7280;
      font-weight: 400;
    }

    & .value {
      color: #030712;
      font-weight: 700;
    }
  }

  &__details {
    display: flex;
    align-items: center;
    gap: 8px;
    &>span {
      font-weight: 700;
    }
  }

  &__info {
    display: flex;
    align-items: center;

    & span {
      display: block;
      font-weight: 500;

      &:nth-child(2) {
        color: #030712;
      }

      &:nth-child(1) {
        color: #6B7280;
        margin-left: 4px;
      }
    }
  }

  &__avatar {
    min-width: 24px;
    max-width: 24px;
    aspect-ratio: 1/1;
    border-radius: 50%;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.10), 0 1px 2px -1px rgba(0, 0, 0, 0.10);
    background: #F3F4F6;
    margin-right: 12px;
  }
}

@media (max-width: 1200px) {
  .profile-tokens {
    &__cells {
      flex-wrap: wrap;
    }

    &__cell {
      &:nth-child(1) {
        width: 100%;
      }

      &:nth-child(2), &:nth-child(3) {
        width: calc(50% - 4px);
      }

      & .label {
        color: #6B7280;
        font-weight: 400;
      }

      & .value {
        color: #030712;
        font-weight: 700;
      }
    }
  }
}
</style>