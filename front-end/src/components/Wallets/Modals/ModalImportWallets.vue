<template>
  <div class="import-wallets">
    <UITabs>
      <UITab
        v-for="tab in tabs"
        :key="tab.val"
        :is_active="activeTab.val === tab.val"
        @click="handleTabChange(tab)"
      >
        {{ tab.label }}
      </UITab>
    </UITabs>

    <div class="import-wallets__manual manual">
      <div v-if="activeTab.val === 'manual'" class="manual__textarea">
        <button @click="pasteText" class="manual__textarea_paste paragraph-small">
          <SVGPaste/>
          Paste from clipboard
        </button>
        <UIBaseTextarea placeholder="Address" v-model="manualTabData.keys"/>
      </div>

      <div
        v-else
        class="import-wallets__file file file"
      >
        <div class=" file__template paragraph-small medium">
          <SVGDownload />
          Download template:
          <button type="button" class="paragraph-small medium" @click="handleTemplateDownload('csv')">CSV</button>
          /
          <button type="button" class="paragraph-small medium" @click="handleTemplateDownload('xlsx')">XLSX</button>
        </div>
        <div class="file__inner">
          <input ref="walletFileImportRef" type="file" accept=".txt,.csv,.xlsx,text/plain,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" @change="handleFileChange">
          <div
            v-if="!selectedFile.file && !selectedFileError"
            class="file__empty_wrapper"
            @dragover.prevent
            @dragenter.prevent
            @drop.prevent="handleDrop"
            @click="triggerFileInput"
          >
            <div class="file__empty">
              <div class="black paragraph-medium">Drag and drop a file here or</div>
              <UIButton @cta="triggerFileInput" size="regular" color_type="secondary" class="file__upload">
                <template #left-icon>
                  <SVGUpload/>
                </template>
                Upload file
              </UIButton>
              <span class="grey regular paragraph-small">Upload file with public wallet addresses</span>
              <span class="grey regular paragraph-small">Supported formats: .txt, .csv, .xlsx</span>
            </div>
          </div>
        </div>

        <div v-if="selectedFile.file || selectedFileError" class="file__selected">
          <div class="file__selected_inner">
            <div
              :class="['file__selected_top', {open: isWalletsListOpen}]"
              @click="toggleWalletsList"
            >
              <div :class="['icon', {error: selectedFileError}]">
                <span v-if="!selectedFileError" class="monospaced-mini regular">{{selectedFileFormat}}</span>
                <SVGCircleSlash v-else/>
              </div>
              <div class="file__selected_info">
                <div class="paragraph-small medium">{{ selectedFileText.label }}</div>
                <span class="paragraph-small regular grey">{{selectedFileText.text}}</span>
              </div>
              <div v-if="selectedTabValidationData.valid_count" class="arrow">
                <SVGSmallArrowDown color="#4B5563"/>
              </div>
            </div>

            <Transition name="nested-expand">
              <div v-if="isWalletsListOpen" class="file__selected_wallets">
                <div class="wrapper">
                  <ul>
                    <li
                      v-for="link in selectedTabValidationData.valid_keys"
                      :key="link"
                      class="paragraph-small regular"
                    >
                      <span>{{ formatWalletAddress(link) }}</span>
                    </li>
                  </ul>
                </div>
              </div>
            </Transition>
          </div>
          <div class="file__selected_btns">
            <UIButton
              v-if="!selectedFileError"
              color_type="destructive"
              size="large"
              @cta="handleFileRemove"
            >
              Remove
            </UIButton>
            <UIButton
              @cta="triggerFileInput"
              color_type="outline"
              size="large"
            >
              {{!selectedFileError ? 'Replace file' : 'Try again'}}
            </UIButton>
          </div>
        </div>
      </div>

      <div v-if="isAlertVisible" class="manual__alert">
        <UIAlert :icon="SVGAlertInfo" text="Private keys will be encrypted and stored securely." status="blue"/>
      </div>
      <div class="manual__info">
        <div class="manual__info_value paragraph-small">Valid <span>{{ selectedTabValidationData.valid_count }}</span>
        </div>
        <div class="manual__info_value paragraph-small">Invalid
          <span>{{ selectedTabValidationData.invalid_count }}</span></div>
        <div class="manual__info_value paragraph-small">Duplicates
          <span>{{ selectedTabValidationData.duplicate_count }}</span></div>
      </div>
    </div>

    <div :class="['import-wallets__btns', {disabled: isImporting}]">
      <UIButton
        color_type="outline"
        @cta="modalsStore.closeModal"
      >
        Cancel
      </UIButton>
      <UIButton
        :is_disabled="isImportDisabled"
        color_type="primary"
        @cta="handleWalletsImport"
      >
        {{ importBtnText }}
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import UITabs from "../../UI/UITabs.vue";
import UITab from "../../UI/UITab.vue";
import {computed, ref, watch} from "vue";
import UIButton from "../../UI/UIButton.vue";
import UIAlert from "../../UI/UIAlert.vue";
import SVGAlertInfo from "../../SVG/SVGAlertInfo.vue";
import UIBaseTextarea from "../../UI/UIBaseTextarea.vue";
import SVGPaste from "../../SVG/SVGPaste.vue";
import {useModalsStore} from "../../../store/modalsStore.js";
import {useProjectsStore} from "../../../store/projectsStore.js";
import {debounce} from "lodash";
import {useToastStore} from "../../../store/toastStore.js";
import {ImportSolWallets} from "../../../api/api.js";
import SVGUpload from "../../SVG/SVGUpload.vue";
import bs58 from 'bs58';
import SVGSmallArrowDown from "../../SVG/SVGSmallArrowDown.vue";
import SVGCircleSlash from "../../SVG/SVGCircleSlash.vue";
import {errorToast, formatWalletAddress} from "../../../helpers/index.js";
import * as XLSX from "xlsx";
import SVGDownload from "../../SVG/SVGDownload.vue";
import {useImportWalletTemplate} from "../../../composable/useImportWalletTemplate.js";

const props = defineProps({
  isLocalImport: false,
})
const emits = defineEmits(["handleKeysImport"]);
const {downloadTemplate} = useImportWalletTemplate();
const ALLOWED_FILE_TYPES = [
  'text/plain',
  'text/csv',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
];
const ALLOWED_EXTENSIONS = ['.txt', '.csv', '.xlsx'];

function isFileAllowed(file) {
  const ext = '.' + (file.name.split('.').pop() || '').toLowerCase();
  return ALLOWED_FILE_TYPES.includes(file.type) || ALLOWED_EXTENSIONS.includes(ext);
}

const modalsStore = useModalsStore();
const projectsStore = useProjectsStore();
const toastStore = useToastStore();
const tabs = [
  {
    label: 'Manual',
    val: 'manual',
  },
  {
    label: 'From file',
    val: 'from-file',
  }
]
const activeTab = ref(tabs[0]);
const selectedProject = ref(null);
const isWalletsListOpen = ref(false);
const isImporting = ref(false);
const selectedFile = ref({
  file: null,
  name: '',
});
const manualTabData = ref({
  keys: '',
  valid_keys: [],
  valid_count: 0,
  invalid_count: 0,
  duplicate_count: 0,
})
const fileTabData = ref({
  keys: '',
  valid_keys: [],
  valid_count: 0,
  invalid_count: 0,
  duplicate_count: 0,
})
const selectedFileError = ref('');
const walletFileImportRef = ref(null);
const isImportDisabled = computed(() => {
  if (props.isLocalImport) {
    return !selectedTabValidationData.value.valid_count;
  } else {
    return !selectedProject.value || !selectedTabValidationData.value.valid_count;
  }
})
const selectedTabValidationData = computed(() => {
  if (activeTab.value.val === 'from-file') {
    return fileTabData.value;
  } else if (activeTab.value.val === 'manual') {
    return manualTabData.value;
  }
})
const importBtnText = computed(() => {
  const count = selectedTabValidationData.value.valid_count || 0;

  if (isImporting.value) {
    return 'Importing...'
  } else if (count) {
    return `Import ${count} wallets`
  } else {
    return 'Import'
  }
})
const isAlertVisible = computed(() => {
  if (activeTab.value.val === 'manual') return true;
  else if (activeTab.value.val === 'from-file') {
    return !selectedFile.value.file;
  }
})
const selectedFileText = computed(() => {
  const fields = {
    label: selectedFile.value.name,
    text: `${selectedTabValidationData.value.valid_count
    + selectedTabValidationData.value.duplicate_count
    + selectedTabValidationData.value.invalid_count} addresses detected`
  };

  if (selectedFileError.value) {
    if (selectedFileError.value === 'Incorrect file type') {
      fields.label = 'Unsupported file format.';
      fields.text = 'Please upload .txt, .csv or .xlsx file.'
    } else if (selectedFileError.value === 'Failed to read the file') {
      fields.label = selectedFile.value.name;
      fields.text = selectedFileError.value;
    }
  }

  return fields;
})
const handleTabChange = (tab) => {
  activeTab.value = tab;
}

const handleTemplateDownload = (format) => {
  downloadTemplate(format);
}

const pasteText = async () => {
  try {
    if (!navigator.clipboard) {
      throw new Error('Clipboard API not supported')
    }

    if (navigator.permissions) {
      const permission = await navigator.permissions.query({name: 'clipboard-read'})

      if (permission.state === 'denied') {
        console.warn('Clipboard permission denied')
        return
      }
    }

    const text = await navigator.clipboard.readText()

    if (typeof text === 'string') {
      manualTabData.value.keys = text
    }

  } catch (err) {
    console.error('Paste failed:', err)
  }
}

const handleDrop = (event) => {
  event.preventDefault();
  const droppedFile = event.dataTransfer.files[0];
  handleFiles(droppedFile);
};

const handleFileChange = (event) => {
  const inputFile = event.target.files[0];
  handleFiles(inputFile)
};

function parseCsvToKeys(text) {
  if (!text || !text.trim()) return [];
  return text
    .split(/\r?\n/)
    .flatMap((line) => line.split(/[,;\t]/).map((v) => v.trim()).filter(Boolean))
    .filter(Boolean);
}

function parseXlsxToKeys(arrayBuffer) {
  const workbook = XLSX.read(arrayBuffer, { type: 'array' });
  const firstSheetName = workbook.SheetNames[0];
  if (!firstSheetName) return [];
  const sheet = workbook.Sheets[firstSheetName];
  const rows = XLSX.utils.sheet_to_json(sheet, { header: 1 });
  return rows.flat().filter((cell) => cell != null && String(cell).trim() !== '').map((cell) => String(cell).trim());
}

function getFileExtension(file) {
  return ('.' + (file.name.split('.').pop() || '').toLowerCase());
}

const handleFiles = (file) => {
  if (!file) return;

  if (!isFileAllowed(file)) {
    selectedFileError.value = 'Incorrect file type';
    return;
  }

  selectedFile.value.file = file;
  selectedFile.value.name = file.name;
  selectedFileError.value = '';
  walletFileImportRef.value.value = '';

  const ext = getFileExtension(file);
  const isXlsx = ext === '.xlsx' || file.type === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet';

  if (isXlsx) {
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const keys = parseXlsxToKeys(e.target.result);
        analyzeKeys(keys);
      } catch {
        selectedFileError.value = 'Failed to read the file';
      }
    };
    reader.onerror = () => {
      selectedFileError.value = 'Failed to read the file';
    };
    reader.readAsArrayBuffer(file);
    return;
  }

  const reader = new FileReader();
  reader.onload = (e) => {
    const text = e.target.result;
    const keys = ext === '.csv' ? parseCsvToKeys(text) : formatKeysToArray(text);
    analyzeKeys(keys);
  };
  reader.onerror = () => {
    selectedFileError.value = 'Failed to read the file';
  };
  reader.readAsText(file);
}

const selectedFileFormat = computed(() => {
  const name = selectedFile.value?.name || '';
  const lastDot = name.lastIndexOf('.');
  if (lastDot === -1) return '';
  return name.slice(lastDot + 1).toLowerCase();
});

const handleFileRemove = () => {
  selectedFile.value = {file: null, name: ''}
}

const triggerFileInput = () => {
  walletFileImportRef.value?.click();
};

const toggleWalletsList = () => {
  if (!selectedTabValidationData.value.valid_count) return;

  isWalletsListOpen.value = !isWalletsListOpen.value;
}

function getNumberFromString(str) {
  const num = parseInt(str, 10);
  return Number.isNaN(num) ? null : num;
}

const handleWalletsImport = async () => {
  const data = {
    private_keys: [],
    project_ids: []
  };
  let newFormData = new FormData();

  data.private_keys = selectedTabValidationData.value.valid_keys;
  data.project_ids.push(selectedProject.value?.id || '');

  if (activeTab.value.val === 'from-file') {
    newFormData.append('file', selectedFile.value.file);
    newFormData.append('project_ids', JSON.stringify(data.project_ids));
  }

  try {
    isImporting.value = true;

    if (!props.isLocalImport) {
      await ImportSolWallets(data);

      await projectsStore.updateProjectData(selectedProject.value.id);
    } else {
      emits('handleKeysImport', data.private_keys)
    }

    modalsStore.closeModal();
  } catch (error) {
    if (error.response.status === 409 && error.response.data.includes('already exist')) {
      const failedWallets = getNumberFromString(error.response.data);

      if (data.private_keys.length > failedWallets) {
        await projectsStore.updateProjectData(selectedProject.value.id);
      } else {
        toastStore.error({text: "Imported wallets already exist"});
      }

      modalsStore.closeModal();
    } else {
      console.error(error);
      errorToast(error.response.data);
    }

  } finally {
    isImporting.value = false;
  }
}

function formatKeysToArray(str) {
  if (!str) return [];

  return str
    .split(/\r?\n/)
    .map(v => v.trim())
    .filter(Boolean);
}

function analyzeKeys(keys) {
  if (!Array.isArray(keys)) return;
  const seen = new Set()
  const BASE58_REGEX = /^[1-9A-HJ-NP-Za-km-z]+$/;
  const MIN_LEN = 34;

  let valid = 0
  let invalid = 0
  let duplicates = 0

  selectedTabValidationData.value.valid_keys = [];

  for (const key of keys) {
    if (!BASE58_REGEX.test(key) || key.length < MIN_LEN) continue

    if (seen.has(key)) {
      duplicates++
      continue
    }

    seen.add(key)

    try {
      const decoded = bs58.decode(key)

      if (decoded.length === 64) {
        selectedTabValidationData.value.valid_keys.push(key)
        valid++
      } else {
        invalid++
      }
    } catch {
      invalid++
    }
  }

  selectedTabValidationData.value.valid_count = valid
  selectedTabValidationData.value.invalid_count = invalid
  selectedTabValidationData.value.duplicate_count = duplicates
}

const analyzeKeysDebounced = debounce(analyzeKeys, 400)

watch(() => manualTabData.value.keys, (val) => {
  const keys = formatKeysToArray(val);

  analyzeKeysDebounced(keys)
})

watch(() => modalsStore.modalData, (newVal) => {
  if (newVal && newVal.item) {
    selectedProject.value = newVal.item;
  } else if (projectsStore.selectedProject) {
    selectedProject.value = projectsStore.selectedProject;
  }
}, {immediate: true})
</script>
<style scoped lang="scss">
.import-wallets {
  width: 704px;

  & .grey {
    color: #6B7280;
  }

  & .black {
    color: #111827;
  }

  & .regular {
    font-weight: 400;
  }

  &__btns {
    margin-top: 40px;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;

    &.disabled {
      opacity: .5;
      pointer-events: none;
    }
  }
}

.manual {
  &__info {
    margin-top: 20px;
    display: flex;
    align-items: center;
    gap: 20px;

    &_value {
      color: #030712;
      font-weight: 500;

      & span {
        color: #6B7280;
      }
    }
  }

  &__alert {
    margin-top: 10px;
  }

  &__textarea {
    margin-top: 10px;
    display: flex;
    flex-direction: column;

    ::v-deep(.base-textarea__input) {
      height: 108px;
    }

    &_paste {
      display: flex;
      align-items: center;
      gap: 6px;
      font-weight: 500;
      background: transparent;
      align-self: flex-end;
      height: 32px;
    }
  }
}

.file {
  margin-top: 16px;
  display: flex;
  flex-direction: column;


  &__inner {
    display: flex;
    flex-direction: column;
    position: relative;
    align-items: center;
    justify-content: center;
  }

  &__template {
    display: flex;
    align-items: center;
    padding: 12px 5.5px;
    gap: 4px;

    & svg {
      width: 14px;
      height: 14px;
    }

    & button {
      background: transparent;
      color: #3B82F6;
    }
  }

  &.selected {
    align-items: flex-start;
    height: 60px;
  }

  &__empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    &_wrapper {
      width: 100%;
      display: flex;
      flex-direction: column;
      position: relative;
      align-items: center;
      justify-content: center;
      height: 235px;
      border-radius: 8px;
      border: 1px dashed #D1D5DB;
      background: #FFF;
      padding: 0 20px;
    }
  }

  &__selected {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-direction: column;
    gap: 10px;

    &_btns {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-left: auto;
    }

    &_inner {
      border-radius: 8px;
      border: 1px solid #D1D5DB;
      background: #FFF;
      width: 100%;
    }

    &_top {
      width: 100%;
      display: flex;
      padding: 20px;

      &.open {
        & .arrow {
          transform: rotate(180deg);
        }
      }

      & .icon {
        margin-top: 3px;
        height: 16px;
        min-width: 20px;
        background: #16A34A;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 3px;
        color: #FFF;

        & span {
          padding: 0 3px;
        }

        &.error {
          background: transparent;
        }
      }

      & .arrow {
        margin-left: auto;
        height: 21px;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: .3s ease;
      }
    }

    &_info {
      margin-left: 8px;
    }

    &_wallets {
      border-top: 1px solid #D1D5DB;
      max-height: 220px;
      padding: 20px;
      overflow: hidden;
      display: flex;
      flex-direction: column;

      & .wrapper {
        overflow: scroll;
      }

      & ul {
        display: flex;
        flex-direction: column;
      }

      & li {
        height: 36px;
        display: flex;
        align-items: center;
      }

      & span {
        text-overflow: ellipsis;
        overflow: hidden;
      }
    }
  }

  &__upload {
    margin: 12px 0 24px;
  }

  & input {
    user-select: none;
    pointer-events: none;
    visibility: hidden;
    position: absolute;
    top: 0;
    left: 0;
  }
}

@media (max-width: 1200px) {
  .import-wallets {
    width: 100%;
  }
}

.nested-expand-enter-active,
.nested-expand-leave-active {
  transition: max-height 0.3s ease,
  opacity 0.25s ease,
  padding 0.25s ease;
  overflow: hidden;
}

.nested-expand-enter-from,
.nested-expand-leave-to {
  max-height: 0;
  opacity: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.nested-expand-enter-to,
.nested-expand-leave-from {
  max-height: 220px;
  opacity: 1;
}
</style>