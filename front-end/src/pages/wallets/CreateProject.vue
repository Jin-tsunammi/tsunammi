<template>
  <div class="wallet-project">
    <PageLoading v-if="isPageLoading"/>
    <template v-else>
      <div v-if="!projectsStore.allProjects.length" class="wallet-project__empty">
        <SVGFolderOpenDot />
        <div class="paragraph-medium bold">No pools yet</div>
        <p class="paragraph-small regular grey">Create your first pool to generate or import wallets and start managing funds.</p>
        <UIButton
          color_type="outline"
          size="large"
          @cta="openWalletModal({type: 'create-project'})"
        >
          Create new Pool
        </UIButton>
      </div>

      <div v-else class="wallet-project__content">
        <div class="wallet-project__top">
          <UISectionTitleWithBorder>
            {{'Manage Wallets'}}
            <template #actions>
              <UIButton
                color_type="primary"
                size="large"
                @cta="openWalletModal({type: 'create-project'})"
              >
                Create new Pool
              </UIButton>
            </template>
          </UISectionTitleWithBorder>
        </div>
        <div class="wallet-project__table">
          <UITable
            :columns="columns"
            :rows="projects"
            @handle-row-click="openProject"
          >
            <template #project="{ item }">
              <div class="table__project">
                <div class="project-name">
                  <div class="image">
                    <SVGLogoTsunammi color="#FFF"/>
                  </div>
                  <span class="paragraph-small regular">{{ item.name }}</span>
                </div>
              </div>
            </template>
            <template #created_at="{ item }">
              <div class="created_at">
                <div v-if="isSyncing === item.id" class="paragraph-small medium grey">
                  <UISpinner color="#2563EB"/>
                  Syncing...
                </div>

                <div v-else class="paragraph-small regular">
                  {{ formatDate(item.created_at).date }}
                  <span class="paragraph-small regular grey">{{ formatDate(item.created_at).time }}</span>
                </div>
              </div>
            </template>
            <template #lifetime="{ item }">
              <span class="paragraph-small regular">{{ daysSince(item.created_at) }}</span>
            </template>
            <template #wallets_qty="{ item }">
              <span class="paragraph-small regular">{{ item.wallet_count }}</span>
            </template>
            <template #actions="{item}">
              <div class="table__actions">
                <template v-if="item.user_id">
                  <UIDotsMenu
                    :menu="dotsMenu"
                    @handle-option-select="handleMenuOptionClick($event, item)"
                  />
                  <button class="table__action_btn" @click.stop="openProject(item)">
                    <SVGSmallArrowDown color="#030712"/>
                  </button>
                </template>
              </div>
            </template>
          </UITable>
        </div>
        <div v-if="totalPages > 1" class="wallet-project__pagination">
          <Pagination :current-page="currentPage" :total="totalPages" @cta="handlePageChange"/>
        </div>
      </div>

      <MobileAdaptsNotification class="wallet-project__mobile" />
    </template>
    <Modals>
      <ConfirmationModal
        class="create-confirmation"
        v-if="modalsStore.modalData.type === 'delete-project'"
        main-text="This action will permanently delete the pool and all associated wallets."
        additional-text="This action cannot be undone."
        :confirmation-btn-style="'destructive'"
        :confirmation-btn-text="isLoading ? 'Deleting...' : 'Delete'"
        @handle-confirmation="handleProjectDelete"
        :is-loading="isLoading"
      />
      <ConfirmationModal
        class="create-confirmation"
        v-if="modalsStore.modalData.type === 'create-confirmation'"
        additional-text="Add wallets to begin managing balances and automated operations."
        confirmation-btn-text="Manage Pool"
        cancellation-btn-text="Ok"
        header-color="success"
        @handle-confirmation="openProject(modalsStore.modalData.item)"
      />

      <ModalCreateEditProject v-if="modalsStore.modalData.type === 'create-project'" />
      <ModalImportWallets v-if="modalsStore.modalData.type === 'import-wallets'" />

      <template #custom-content>
        <ModalCreateWallets v-if="modalsStore.modalData.type === 'create-wallets'" />
      </template>
    </Modals>
  </div>
</template>
<script setup>
import {computed, onBeforeUnmount, onMounted, ref, watch} from 'vue';
import UIButton from "../../components/UI/UIButton.vue";
import {useModalsStore} from "../../store/modalsStore.js";
import Modals from "../../components/UI/Modals.vue";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import {useProjectsStore} from "../../store/projectsStore.js";
import {daysSince, formatDate} from "../../helpers/index.js";
import UITable from "../../components/UI/UITable.vue";
import SVGDelete from "../../components/SVG/SVGDelete.vue";
import SVGFolderOpenDot from "../../components/SVG/SVGFolderOpenDot.vue";
import {useRouter} from "vue-router";
import UIDotsMenu from "../../components/UI/UIDotsMenu.vue";
import SVGPlus from "../../components/SVG/SVGPlus.vue";
import SVGRefresh from "../../components/SVG/SVGRefresh.vue";
import SVGImport from "../../components/SVG/SVGImport.vue";
import SVGSmallArrowDown from "../../components/SVG/SVGSmallArrowDown.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import ModalCreateEditProject from "../../components/Wallets/Modals/ModalCreateEditProject.vue";
import UISpinner from "../../components/UI/UISpinner.vue";
import {useExcelExport} from "../../composable/useExcelExport.js";
import ModalCreateWallets from "../../components/Wallets/Modals/ModalCreateWallets.vue";
import ModalImportWallets from "../../components/Wallets/Modals/ModalImportWallets.vue";
import Pagination from "../../components/UI/Pagination.vue";
import PageLoading from "../../components/UI/PageLoading.vue";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useToastStore} from "../../store/toastStore.js";
import SVGLogoTsunammi from "../../components/SVG/SVGLogoTsunammi.vue";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";

const router = useRouter();
const modalsStore = useModalsStore();
const projectsStore = useProjectsStore();
const toastStore = useToastStore();
const userStore = useUserStore();
const projects = ref([]);
const isLoading = ref(false);
const isPageLoading = ref(true);
const isFrozenMoneyToolTipVisible = ref(false);
const isSyncing = ref('');
const currentPage = ref(1);
const totalItems = ref(0);
const itemsOnPage = 12;
const sortParams = ref({
  sortBy: 'created_at', // created_at || wallet_count || last_sync || name
  sortOrder: 'asc' //asc || desc
})

const {exportProjectExcel} = useExcelExport()
const columns = [
  { label: 'Wallet Pools', field: 'project' },
  { label: 'Created', field: 'created_at' },
  { label: 'Lifetime', field: 'lifetime' },
  { label: 'Quantity', field: 'wallets_qty' },
  { label: '', field: 'actions' },
];
const dotsMenu = [
  [{label: "Create wallets", icon: SVGPlus, action: "create-wallets"}, {label: "Import wallets", icon: SVGImport, action: "import-wallets"}],
  [{label: "Sync", icon: SVGRefresh, action: "sync"}],
  [{label: "Delete", icon: SVGDelete, action: "delete"}],
]
const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsOnPage) || 0;
})
const openWalletModal = ({type, item=null}) => {
  if (userStore.isOpenLoginModal()) return;

  modalsStore.modalData.type = type;

  if (type === 'create-project') {
    if (item) {
      modalsStore.modalData.title = 'Edit pool';
    } else {
      modalsStore.modalData.title = 'Create new Pool';
    }
  } else if (type === 'delete-project') {
    modalsStore.modalData.type = 'delete-project';
    modalsStore.modalData.title = `Delete ${item?.name || ''}?`;
    modalsStore.modalData.action = 'confirmation';
  } else if (type === 'edit-project') {
    modalsStore.modalData.title = 'Edit pool';
  } else if (type === 'create-wallets') {
    modalsStore.modalData.is_custom = true;
    modalsStore.modalData.title = 'Create wallets';
  } else if (type === 'import-wallets') {
    modalsStore.modalData.title = 'Import wallets';
  }

  if (item) {
    modalsStore.modalData.item = item;
  }

  modalsStore.openModal();
}

const toggleUIKit = () => {
  isFrozenMoneyToolTipVisible.value = !isFrozenMoneyToolTipVisible.value
}

const handlePageChange = async (page) => {
  if (!page) return;
  currentPage.value = page;
  isPageLoading.value = true;
  await getProjects();
  isPageLoading.value = false;
}

const handleProjectDelete = async() => {
  isLoading.value = true;
  await projectsStore.handleProjectDelete(modalsStore.modalData.item);
  isLoading.value = false;
}

const openProject = (project) => {
  if (!project) return;

  if (modalsStore.modalData.is_open) {
    modalsStore.closeModal()
  }
  router.push({name: 'WalletsSelectedProject', params: {project_id: project.id}});
}

const handleMenuOptionClick = async(action, project=null) => {
  if (!project) return;

  switch (action) {
    case 'delete':
      openWalletModal({type: 'delete-project', item: project});
      break;
    case 'edit-project':
      openWalletModal({type: 'create-project', item: project});
      break;

    case 'sync':
      isSyncing.value = project.id;
      await projectsStore.updateProjectData(project.id);
      isSyncing.value = '';

      break;

    case 'export':
      exportProjectExcel({project, wallets: project.wallets});

      break;

    case 'create-wallets':
    case 'import-wallets':
      openWalletModal({type: action, item: project});

      break;
  }
}

const handlePageRefresh = async(isRefresh=false) => {
  try {
    isPageLoading.value = true;

    if (userStore.isUserAuth) {
      await getProjects();
    }
  } finally {
    isPageLoading.value = false;
  }

  if (isRefresh) {
    toastStore.success({text: 'Page data refreshed'});
  }
}

const getProjects = async() => {
  const params = {
    page: currentPage.value,
    pageSize: itemsOnPage,
  }

  const resp = await projectsStore.getAllProjects(params);

  totalItems.value = resp.total;
}

watch(() => projectsStore.allProjects, (newVal) => {
  projects.value = newVal;
}, {immediate: true, deep: true})

watch(() => userStore.isUserAuth, async(newVal) => {
  if (newVal) {
    await handlePageRefresh();
  }
})

useHeaderRefresh(() => handlePageRefresh(true));

onMounted(async() => {
  await handlePageRefresh()
})

onBeforeUnmount(() => {
  projectsStore.setAllProjects([]);
})
</script>
<style scoped lang="scss">
.wallet-project {
  height: 100%;
  display: flex;
  flex-direction: column;

  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;
  }

  &__mobile {
    display: none;
  }

  &__content {
    display: flex;
    flex-direction: column;
    min-height: fit-content;
    height: 100%;
  }

  &__pagination {
    width: 100%;
    margin-top: auto;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &__empty {
    display: flex;
    height: 309px;
    padding: 32px 0;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    flex-shrink: 0;
    align-self: stretch;
    border-radius: 12px;
    background: #FFF;

    & div {
      margin-top: 12px;
    }

    & p {
      max-width: 339px;
      text-align: center;
      margin: 4px 0 24px;
    }
  }

  &__table {
    width: 100%;
    max-width: 1400px;
    margin-bottom: 20px;

    ::v-deep(.table__header_col) {
      &.project {
        width: calc((530 / 1163) * 100%);
      }

      &.created_at {
        width: calc((170 / 1163) * 100%);
      }

      &.lifetime {
        width: calc((150 / 1163) * 100%);
      }

      &.wallets_qty {
        width: calc((150 / 1163) * 100%);
      }

      &.actions {
        width: calc((163 / 1163) * 100%);
      }
    }

    ::v-deep(.table__row) {
      transition: 0.3s ease;
      &:hover {
        cursor: pointer;
        background: rgba(229, 231, 235, .5);
      }
    }

    ::v-deep(.table__row_cell) {
      &.project {
        width: calc((530 / 1163) * 100%);
      }

      &.created_at {
        width: calc((170 / 1163) * 100%);
      }

      &.lifetime {
        width: calc((150 / 1163) * 100%);
      }

      &.wallets_qty {
        width: calc((150 / 1163) * 100%);
      }

      &.actions {
        width: calc((163 / 1163) * 100%);
      }
    }
  }
}

.create-confirmation {
  max-width: 360px;
}

.table {
  &__project {
    display: flex;
    align-items: center;
    gap: 8px;

    & .project-name {
      display: flex;
      align-items: center;
      gap: 8px;
      overflow: hidden;

      & span {
        text-wrap: nowrap;
        display: block;
        text-overflow: ellipsis;
        overflow: hidden;
      }
    }

    & .image {
      aspect-ratio: 1/1;
      max-width: 32px;
      min-width: 32px;
      border-radius: 50%;
      overflow: hidden;
      background: #000;
      display: flex;
      align-items: center;
      justify-content: center;

      & svg {
        width: 50%;
        height: 50%;
      }
    }
  }

  &__head {
    display: flex;
    align-items: center;
    gap: 8px;

    & .tooltip {
      position: absolute;
      bottom: calc(100% + 10px);
      left: 50%;
      transform: translateX(-50%);
      width: 205px;
      z-index: 5;
      font-weight: 400;

      &-wrapper {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: center;
      }
    }
  }

  & .bold {
    font-weight: 600;
  }

  & .medium {
    font-weight: 500;
  }

  & .regular {
    font-weight: 400;
  }

  & .grey {
    color: #6B7280;
  }

  &__empty {
    padding: 60px 0;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    &_title {
      margin: 12px 0 4px;
    }

    & p {
      text-align: center;
      max-width: 379px;
    }

    &_btns {
      margin-top: 24px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  &__actions {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: flex-end;
  }

  &__action_btn {
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    transition: .3s ease;
    border-radius: 8px;
    min-width: 32px;
    min-height: 32px;

    & svg {
      transform: rotate(-90deg);
    }

    &:hover{
      background: rgba(0, 0, 0, 0.05);
    }
  }

  &__map_btn {
    min-width: 112px;
  }
}

@media (max-width: 1200px) {
.wallet-project {
  &__content {
    display: none;
  }
  &__mobile {
    margin-top: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
}
</style>