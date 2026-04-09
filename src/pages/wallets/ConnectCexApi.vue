<template>
  <div class="cex-api">
    <div class="cex-api__content">
      <PageLoading v-if="isPageLoading" />
      <template v-else>
        <DesktopTopUp
          :columns="columns"
          :rows="cexApiStore.allCEXApi"
          @open-c-e-x-modal="openCEXModal"
        />

        <MobileAdaptsNotification class="cex-api__mobile"/>

        <div class="cex-api__pagination">
          <!--          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>-->
        </div>
      </template>
      <Modals>
        <ConfirmationModal
          class="create-confirmation"
          v-if="modalsStore.modalData.type === 'delete-api'"
          main-text="This action will permanently delete this item."
          additional-text="This action cannot be undone."
          :confirmation-btn-style="'destructive'"
          :confirmation-btn-text="isDeleting ? 'Deleting...' : 'Delete'"
          @handle-confirmation="handleCEXApiDelete"
          :is-loading="isDeleting"
        />
        <ModalAddAPI v-if="modalsStore.modalData.type === 'wallet-cex-add-api'" :server_ip="serverIP"/>
      </Modals>
    </div>
  </div>
</template>
<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue';
import {useModalsStore} from "../../store/modalsStore.js";
import Modals from "../../components/UI/Modals.vue";
import Pagination from "../../components/UI/Pagination.vue";
import ModalAddAPI from "../../components/TopUpCEX/Modals/ModalAddAPI.vue";
import DesktopTopUp from "../../components/Wallets/TopUpCex/DesktopTopUp.vue";
import {useCEXApiStore} from "../../store/cexStore.js";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import {GetServerIP} from "../../api/api.js";
import PageLoading from "../../components/UI/PageLoading.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {errorToast} from "../../helpers/index.js";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const modalsStore = useModalsStore();
const cexApiStore = useCEXApiStore();
const userStore = useUserStore();
const currentPage = ref(1);
const totalOrders = ref(300);
const selectedItem = ref(null);
const isDeleting = ref(false);
const isPageLoading = ref(true);
const serverIP = ref('');
const itemsOnPage = 20;
const columns = [
  { label: 'Title', field: 'title' },
  { label: 'Status', field: 'status' },
  { label: 'Date added', field: 'date' },
  { label: 'Lifetime', field: 'lifetime' },
  { label: 'API name', field: 'api' },
  { label: 'Total deposited', field: 'deposited' },
  { label: '', field: 'actions' },
];

const totalPages = computed(() => {
  return Math.ceil(totalOrders.value / itemsOnPage);
})

const openCEXModal = ({type, item=null}) => {
  if (userStore.isOpenLoginModal() || !type) return;

  if (type === 'delete') {
    modalsStore.modalData.type = 'delete-api';
    modalsStore.modalData.action = 'confirmation'
    modalsStore.modalData.title = `Delete ${item.name || ''}?`

    if (item) selectedItem.value = item;
  } else if (type === 'add-api') {
    modalsStore.modalData.title = 'Add new API';
    modalsStore.modalData.type = 'wallet-cex-add-api';
    modalsStore.modalData.action = 'confirmation';
  }

  modalsStore.openModal();
}

const handleCEXApiDelete = async() => {
  isDeleting.value = true;
  await cexApiStore.deleteCEXApi(selectedItem.value.id);
  isDeleting.value = false;
}

const handlePageChange = async (page) => {
  currentPage.value = page;

  // await getUsers();
}

const handlePageRefresh = async(isRefreshing=false) => {
  try {
    isPageLoading.value = false;
    if (userStore.isUserAuth) {
      await cexApiStore.getAllCEXApi();
      const resp = await GetServerIP();

      serverIP.value = resp?.data?.ip || '';

      if (isRefreshing) {
        toastStore.success({text: 'Page data refreshed'});
      }
    }
  } finally {
    isPageLoading.value = false;
  }
}

watch(() => modalsStore.modalData.type, (newVal) => {
  if (!newVal && selectedItem.value) selectedItem.value = null;
})

watch(() => userStore.isUserAuth, async(newVal) => {
  if (newVal) {
    await handlePageRefresh();
  }
})

useHeaderRefresh(() => handlePageRefresh(true));
onMounted(async() => {
  await handlePageRefresh();
})
onBeforeUnmount(() => {
  cexApiStore.clearStore();
})
</script>
<style scoped lang="scss">
.cex-api {
  &__content {
    width: 100%;
    display: flex;
    flex-direction: column;
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

  &__mobile {
    display: none;
  }

  &__pagination {
    margin-inline: auto;
    margin-top: auto;
  }

  &__header {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}
.create-confirmation {
  max-width: 360px;
}

@media (max-width: 1200px) {
  .cex-api {
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