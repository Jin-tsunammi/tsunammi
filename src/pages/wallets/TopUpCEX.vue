<template>
  <div class="selected-cex">
    <div class="selected-cex__content">
      <PageLoading v-if="isPageLoading" />
      <template v-else>
        <div class="selected-cex__desktop">
          <TopUpBlock
            :projects="projects"
            :cex-a-p-i="cexApiStore.allCEXApi"
            @open-modal="openModal"
          />
          <HistoryDesktop :columns="columns" :rows="depositHistory"/>
        </div>
        <MobileAdaptsNotification class="selected-cex__mobile"/>

        <!--        <div class="selected-cex__mobile">-->
        <!--          <UITabs>-->
        <!--            <UITab-->
        <!--              v-for="tab in mobileTabs"-->
        <!--              :key="tab.val"-->
        <!--              :is_active="selectedTab.val === tab.val"-->
        <!--              @click="selectedTab = tab"-->
        <!--            >-->
        <!--              {{tab.label}}-->
        <!--            </UITab>-->
        <!--          </UITabs>-->

        <!--          <TopUpBlock v-show="selectedTab.val === 'top-up'" />-->
        <!--          <HistoryMobile v-show="selectedTab.val === 'history'" :rows="tableData"/>-->
        <!--        </div>-->
      </template>

      <Modals>
        <ConfirmationModal
          v-if="modalsStore.modalData.type === 'top-up-confirmation'"
          main-text="Are you sure that want to proceed with pop up? Action can not be undone."
          confirmation-btn-style="primary"
          confirmation-btn-text="Confirm"
          @handle-confirmation="handleTopUp"
        />

        <ConfirmationModal
          class="top-up-success"
          v-if="modalsStore.modalData.type === 'top-up-success'"
          main-text="The funding request has been successfully submitted to the network."
          header-color="success"
          confirmation-btn-style="outline"
          confirmation-btn-text="Ok"
          :is-cancel="false"
          @handle-confirmation="modalsStore.closeModal"
        />
      </Modals>
    </div>
  </div>
</template>
<script setup>
import {onBeforeUnmount, onMounted, ref, watch} from 'vue';
import {useModalsStore} from "../../store/modalsStore.js";
import Modals from "../../components/UI/Modals.vue";
import HistoryDesktop from "../../components/Wallets/SelectedCEX/HistoryDesktop.vue";
import TopUpBlock from "../../components/Wallets/SelectedCEX/TopUpBlock.vue";
import {useCEXApiStore} from "../../store/cexStore.js";
import {useToastStore} from "../../store/toastStore.js";
import {GetAllProjectsNameOnly, GetDepositHistory, ProcessDepositOrder} from "../../api/api.js";
import {errorToast, formatText} from "../../helpers/index.js";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import PageLoading from "../../components/UI/PageLoading.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const userStore = useUserStore();
const cexApiStore = useCEXApiStore();
const modalsStore = useModalsStore();
const toastStore = useToastStore();
const columns = [
  { label: 'Project/Address', field: 'project_name' },
  { label: 'Date', field: 'created_at' },
  { label: 'Status', field: 'status' },
  { label: 'Sum of deposit', field: 'total_sum_sol' },
  { label: 'Result', field: 'result' },
  { label: '', field: 'actions' },
];
const mobileTabs = [
  {label: 'Top up', val: 'top-up'},
  {label: 'History', val: 'history'},
]
const isPageLoading = ref(true);
const depositHistory = ref([]);
const projects = ref([]);
const selectedTab = ref(mobileTabs[0]);

const openModal = ({type, order_id=null}) => {
  if (type === 'confirmation') {
    modalsStore.modalData.title = 'Top up confirmation'
    modalsStore.modalData.type = 'top-up-confirmation'
    modalsStore.modalData.action = 'confirmation'
  }

  if (type === 'success') {
    modalsStore.modalData.title = 'Funding initiated'
    modalsStore.modalData.type = 'top-up-success'
    modalsStore.modalData.action = 'confirmation'
  }

  if (order_id) {
    modalsStore.modalData.item = order_id;
  }

  modalsStore.modalData.is_open = true;
}

const handleTopUp = async() => {
  if (!modalsStore.modalData.item) return;

  try {
    await ProcessDepositOrder(modalsStore.modalData.item);
    openModal({type: 'success'});

    setTimeout(() => {
      updateDepositHistory();
    }, 1500);
  } catch (e) {
    console.error(e)
    toastStore.error({text: formatText(e?.response?.data)})
  }
}

const getProjects = async() => {
  try {
    const resp = await GetAllProjectsNameOnly();
    projects.value = resp.data;
  } catch (e) {
    console.error(e)
    errorToast(e.response.data)
  }
}

const updateDepositHistory = async(isHeaderRefresh=false, isAuth=false) => {
  try {
    if (!isAuth) {
      isPageLoading.value = true;
    }

    if (userStore.isUserAuth) {
      await getProjects();
      await cexApiStore.getAllCEXApi();
      const resp = await GetDepositHistory();
      depositHistory.value = resp.data?.deposits;

      if (isHeaderRefresh) {
        toastStore.success({text: "Page data refreshed"})
      }
    }
  } catch (e) {
    console.error(e)
    errorToast(e.response.data)
  } finally {
    isPageLoading.value = false;
  }
}

watch(() => userStore.isUserAuth, async(newVal) => {
  if (newVal) {
    await updateDepositHistory(false, true);
  }
})


useHeaderRefresh(() => updateDepositHistory(true));

onMounted(async() => {
  await updateDepositHistory();
})
onBeforeUnmount(() => {
  cexApiStore.clearStore();
})
</script>
<style scoped lang="scss">
.selected-cex {
  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: fit-content;
    flex-grow: 1;
  }

  &__header {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__desktop {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
  }

  &__mobile {
    display: none;
  }
}

.top-up-success {
  ::v-deep(.ui-button) {
    width: fit-content;
    margin: 0 auto;
  }
}

@media (max-width: 1200px) {
  .selected-cex {
    &__desktop {
      display: none;
    }

    &__mobile {
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 16px;
    }
  }
}
</style>