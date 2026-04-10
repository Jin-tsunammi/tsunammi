<template>
  <div class="history">
    <div class="history__content">
      <PageLoading v-if="isPageLoading"/>
      <div class="history__inner" v-show="!isPageLoading">
        <HistoryDesktop :rows="history" :columns="columns" @handle-dates-change="handleDatesRangeChange"/>
        <MobileAdaptsNotification class="history__mobile"/>

        <div v-if="totalPages > 1" class="history__pagination">
          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>
        </div>
      </div>
      <Modals>
      </Modals>
    </div>
  </div>
</template>
<script setup>
import {computed, onMounted, ref, watch} from 'vue';
import Modals from "../../components/UI/Modals.vue";
import HistoryDesktop from "../../components/Wallets/History/HistoryDesktop.vue";
import {GetUserHistory} from "../../api/api.js";
import {useToastStore} from "../../store/toastStore.js";
import {errorToast} from "../../helpers/index.js";
import PageLoading from "../../components/UI/PageLoading.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import Pagination from "../../components/UI/Pagination.vue";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const toastStore = useToastStore();
const userStore = useUserStore();
const isPageLoading = ref(true);
const history = ref([]);
const currentPage = ref(1);
const totalItems = ref(0);
const itemsOnPage = 12;
const columns = [
  {label: 'Date', field: 'created_at'},
  {label: 'Action', field: 'action'},
  {label: 'Value', field: 'value'},
];

const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsOnPage) || 0;
})

const handlePageChange = async (page) => {
  if (!page) return;
  currentPage.value = page;
  isPageLoading.value = true;
  await getHistory(false, getParams());
  isPageLoading.value = false;
}

const getHistory = async (isRefresh = false, params = null) => {
  isPageLoading.value = true;
  try {
    if (userStore.isUserAuth) {
      const resp = await GetUserHistory(params);

      history.value = resp?.data?.user_history || [];
      totalItems.value = resp.data?.total || 0;

      if (isRefresh) {
        toastStore.success({text: 'Page data refreshed'});
      }
    }
  } catch (error) {
    errorToast(error.response?.data);
  } finally {
    isPageLoading.value = false;
  }
}

const getParams = (datesRange = []) => {
  let params = {};

  params.page = currentPage.value;
  params.pageSize = itemsOnPage;

  if (datesRange && datesRange.length > 1) {
    const from = new Date(datesRange[0]);
    const to = new Date(datesRange[1]);

    from.setHours(0, 0, 0, 0);
    to.setHours(23, 59, 59, 999);

    params.from = from.toISOString();
    params.to = to.toISOString();
  }

  return params;
}

const handleDatesRangeChange = async (datesRange) => {
  await getHistory(false, getParams(datesRange));
};

watch(() => userStore.isUserAuth, async(newVal) => {
  if (newVal) {
    await getHistory(false, getParams());
  }
})

useHeaderRefresh(() => getHistory(true, getParams()));

onMounted(async () => {
  try {
    await getHistory(false, getParams());
  } catch (error) {
  }
})
</script>
<style scoped lang="scss">
.history {
  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
    min-height: fit-content;
    height: 100%;

    &_label {
      position: relative;
      color: #030712;

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
    display: flex;
    flex-direction: column;
    height: 100%;
    gap: 10px;
  }

  &__pagination {
    width: 100%;
    margin-top: auto;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &__mobile {
    display: none;
  }

  &__header {
    display: flex;
    align-items: center;
    gap: 12px;

    ::v-deep(.ui-select__input) {
      width: 177px;
    }
  }
}

@media (max-width: 1200px) {
  .history {
    &__mobile {
      margin: 16px auto;
      padding: 0 16px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>