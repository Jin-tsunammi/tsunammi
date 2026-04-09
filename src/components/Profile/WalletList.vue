<template>
  <div class="wallet-list">
    <div
      v-for="(wallet, index) in wallets"
      :key="wallet.id"
      class="wallet-list__item"
    >
      <div class="wallet-list__icon">
        <img
          v-if="wallet.logo"
          :src="wallet.logo"
          :alt="wallet.name"
          class="wallet-list__icon-img"
        />
        <div v-else class="wallet-list__icon-placeholder">
          {{ wallet.name.charAt(0) }}
        </div>
      </div>
      
      <div class="wallet-list__info">
        <div class="wallet-list__name">
          <span class="wallet-list__name-text paragraph-medium">{{ wallet.name }}</span>
          <span class="wallet-list__name-id paragraph-medium">{{ wallet.identifier }}</span>
        </div>
      </div>
      
      <div class="wallet-list__separator">/</div>
      
      <div class="wallet-list__balance">
        <span class="wallet-list__balance-value monospaced-medium">{{ formatBalance(wallet.balance) }}</span>
        <span class="wallet-list__balance-unit monospaced-medium">{{ wallet.unit }}</span>
      </div>
    </div>
    
    <div v-if="wallets.length === 0" class="wallet-list__empty">
      <p class="paragraph-medium color-secondary">No connected wallets</p>
    </div>
  </div>
</template>

<script setup>
import { defineProps } from 'vue'

defineProps({
  wallets: {
    type: Array,
    default: () => []
  }
})

const formatBalance = (balance) => {
  if (!balance) return '0'
  
  const num = parseFloat(balance)
  if (isNaN(num)) return balance
  
  if (num >= 1000) {
    return num.toLocaleString('ru-RU', { 
      minimumFractionDigits: 0,
      maximumFractionDigits: 6 
    }).replace(/,/g, ' ')
  }
  
  return num.toLocaleString('ru-RU', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 6
  })
}
</script>

<style scoped lang="scss">
.wallet-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  width: 100%;
  border-radius: 12px;
  border: 1px solid #D1D5DB;
  overflow: hidden;
  background: transparent;
  
  &__empty {
    padding: 24px;
    text-align: center;
  }
  
  &__item {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.10);
    gap: 12px;
    height: 64px;

    &:last-child {
      border-bottom: none;
    }
  }
  
  &__icon {
    width: 24px;
    height: 24px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    overflow: hidden;
    
    &-img {
      width: 100%;
      height: 100%;
      object-fit: contain;
    }
    
    &-placeholder {
      width: 100%;
      height: 100%;
      border-radius: 4px;
      background: #D1D5DB;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      font-weight: 600;
      color: #6B7280;
    }
  }
  
  &__info {
    flex: 1;
    min-width: 0;
  }
  
  &__name {
    display: flex;
    align-items: center;
    gap: 8px;
    
    &-text {
      color: #302F2F;
      font-weight: 500;
    }
    
    &-id {
      color: #6B7280;
    }
  }
  
  &__separator {
    color: #9CA3AF;
    font-size: 14px;
    font-weight: 400;
    margin: 0 4px;
  }
  
  &__balance {
    width: 100%;
    max-width: 167px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
    margin-left: auto;
    
    &-value {
      color: #302F2F;
    }
    
    &-unit {
      color: #302F2F;
    }
  }
}

@media (max-width: 1200px) {
  .wallet-list {
    &__item {
      height: auto;
      min-height: 76px;
    }

    &__separator {
      display: none;
    }

    &__name {
      flex-direction: column;
      align-items: flex-start;
      gap: 0;
    }
  }
}
</style>
