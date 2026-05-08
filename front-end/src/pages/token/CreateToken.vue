<template>
  <div :class="['token-create']">
    <PageLoading v-if="isPageLoading"/>
    <div v-show="!isPageLoading" class="token-create__inner">
      <div class="token-create__desktop">
        <div class="token-create__desktop_left">
          <CreateTokenTop
            :data="tokenData"
            :errors="errors"
            @handle-image-save="handleLogoChoose"
            @handle-wallet-connect="handleWalletConnect"
            @clear-error-message="clearErrorMessage"
            @handle-image-error="handleImageError"
          />
          <div class="token-create__estimate paragraph-mini regular">
            Estimated costs: 0.01862SOL
          </div>
          <UIAlert
            class="token-create__alert"
            text="You can view and manage your token inside your wallet."
            status="blue"
            :icon="SVGAlertInfo"
          />
          <div class="token-create__btns">
            <UIButton
              class="token-create__start"
              size="large"
              color_type="primary"
              @cta="handleTokenCreate"
              :is_disabled="(isChangesSaving)"
            >
              {{ startButtonText }}
            </UIButton>
          </div>
        </div>
        <div class="token-create__desktop_right">
          <UISectionTitleWithBorder>Token Custom Setting</UISectionTitleWithBorder>
          <div class="token-create__desktop_settings">
            <div
              v-for="item in settings"
              :key="item.key"
              class="setting"
            >
              <div class="setting__top">
                <div class="setting__left">
                  <span class="paragraph-small medium">{{ item.label }}</span>
                  <p class="paragraph-mini regular grey">{{ item.text }}</p>
                </div>
                <UIRoundToggle
                  v-model:is-active="tokenData[item['key']]"
                />
              </div>

              <transition name="pumpfun-socials">
                <div v-if="item.key === 'social_links_toggle' && tokenData.social_links_toggle"
                     class="setting__social-links">
                  <UIBaseInput
                    v-for="key in Object.keys(socialLinksInfo)"
                    :key="key"
                    size="large"
                    v-model="tokenData.social_links[key]"
                    :label="socialLinksInfo[key].label"
                    :placeholder="socialLinksInfo[key].placeholder"
                    class="social-links__item"
                  >
                    <template #icon-left>
                      <component class="icon" v-if="socialLinksInfo[key].icon" :is="socialLinksInfo[key].icon"/>
                    </template>
                  </UIBaseInput>
                </div>
              </transition>
            </div>
          </div>
        </div>
      </div>
      <MobileAdaptsNotification class="token-create__mobile"/>

      <Modals>
        <ConfirmationModal
          class="create-confirmation"
          v-if="modalsStore.modalData.type === 'create-confirmation'"
          additional-text="Check your wallet for token details or visit Solscan"
          confirmation-btn-text="View on Solscan"
          cancellation-btn-text="Ok"
          header-color="success"
          @handle-confirmation="openSolscan(modalsStore.modalData.item)"
        />
      </Modals>
    </div>
  </div>
</template>
<script setup>
import {computed, markRaw, onMounted, ref, watch} from "vue";
import UIButton from "../../components/UI/UIButton.vue";
import UIAlert from "../../components/UI/UIAlert.vue";
import SVGAlertInfo from "../../components/SVG/SVGAlertInfo.vue";
import UISectionTitleWithBorder from "../../components/UI/UISectionTitleWithBorder.vue";
import {useRoute, useRouter} from "vue-router";
import {errorToast, trackGoogleTagEvent} from "../../helpers/index.js";
import PageLoading from "../../components/UI/PageLoading.vue";
import {useToastStore} from "../../store/toastStore.js";
import MobileAdaptsNotification from "../../components/UI/MobileAdaptsNotification.vue";
import {useUserStore} from "../../store/userStore.js";
import {useHeaderRefresh} from "../../composable/useHeaderRefresh.js";
import {useSmartCampaignsStore} from "../../store/smartCampaignsStore.js";
import CreateTokenTop from "../../components/Token/CreateToken/CreateTokenTop.vue";
import UIRoundToggle from "../../components/UI/UIRoundToggle.vue";
import UIBaseInput from "../../components/UI/UIBaseInput.vue";
import SVGTwitter from "../../components/SVG/SVGTwitter.vue";
import SVGDiscord from "../../components/SVG/SVGDiscord.vue";
import SVGWebsite from "../../components/SVG/SVGWebsite.vue";
import SVGTelegram from "../../components/SVG/SVGTelegram.vue";
import {useWallet} from "../../composable/useWallet.js";
import {Connection, Keypair, SystemProgram, Transaction} from "@solana/web3.js";
import {
  createInitializeMintInstruction,
  MINT_SIZE,
  TOKEN_PROGRAM_ID,
  createAssociatedTokenAccountInstruction,
  getAssociatedTokenAddress,
  createMintToInstruction,
  createSetAuthorityInstruction, AuthorityType,
} from "@solana/spl-token";
import {cloneDeep} from "lodash";
import { createUmi } from '@metaplex-foundation/umi-bundle-defaults';
import {updateV1, createV1, TokenStandard} from '@metaplex-foundation/mpl-token-metadata';
import { fromWeb3JsPublicKey, toWeb3JsInstruction } from '@metaplex-foundation/umi-web3js-adapters';
import { none, some, signerIdentity, createNoopSigner } from '@metaplex-foundation/umi';
import {UploadImage, UploadMetadata} from "../../api/api.js";
import Modals from "../../components/UI/Modals.vue";
import ConfirmationModal from "../../components/UI/Modals/ConfirmationModal.vue";
import {useModalsStore} from "../../store/modalsStore.js";

const route = useRoute();
const router = useRouter();
const toastStore = useToastStore();
const modalsStore = useModalsStore();
const smartCampaignStore = useSmartCampaignsStore();
const userStore = useUserStore();
const {isLoading, connectHandler, walletProvider, publicKey, address} = useWallet();
const tokenData = ref({
  name: '',
  ticker: '',
  decimals: 9,
  supply: 0,
  logo: '',
  description: '',
  ownership: '',
  social_links: {
    twitter: '',
    discord: '',
    website: '',
    telegram: '',
  },
  fixed_supply: true,
  revoke_freeze: true,
  immutable: true,
  social_links_toggle: false,
})
const isChangesSaving = ref(false);
const isPageLoading = ref(true)
const defaultErrors = {
  name: '',
  ticker: '',
  supply: '',
  ownership: '',
  logo: '',
}
const errors = ref(cloneDeep(defaultErrors))
const startButtonText = computed(() => {
  return isLoading.value ? 'Creating...' : 'Create token';
})
const socialLinksInfo = {
  twitter: {label: "X (Twitter)", placeholder: "Add X", icon: markRaw(SVGTwitter)},
  discord: {label: "Discord", placeholder: "Add Discord", icon: markRaw(SVGDiscord)},
  website: {label: "Website", placeholder: "Add Website", icon: markRaw(SVGWebsite)},
  telegram: {label: "Telegram", placeholder: "Add Telegram", icon: markRaw(SVGTelegram)},
}
const settings = [
  {
    key: 'fixed_supply',
    label: 'Fixed Supply',
    text: 'Permanently disables minting of new tokens. Total supply becomes fixed and cannot be increased by anyone, including the creator.'
  },
  {
    key: 'revoke_freeze',
    label: 'Revoke Freeze',
    text: 'Removes the ability to freeze token accounts. Ensures holders can always transfer or sell their tokens without restrictions.'
  },
  {
    key: 'immutable',
    label: 'Immutable',
    text: 'Locks token metadata forever. Name, symbol, logo, and description cannot be changed after this action is executed.'
  },
  {
    key: 'social_links_toggle',
    label: 'Social Links',
    text: 'Add Twitter, Telegram, and website URLs to token metadata for community discovery and verification.'
  },
];

const handlePageRefresh = async (isRefreshing = false, isAuth = false) => {
  isPageLoading.value = true;

  if (!isAuth) {
    smartCampaignStore.clearStore();
  }

  if (!userStore.isUserAuth) {
    try {
    } catch (e) {
      errorToast(e.response.data)
    } finally {
      isPageLoading.value = false;
    }

    return
  }

  try {

    if (isRefreshing) {
      toastStore.success({text: "Page is refreshed"})
    }
  } catch (e) {
    errorToast(e.response.data)
  } finally {
    isPageLoading.value = false;
  }
}

const handleImageError = (errorMessage) => {
  errors.value.logo = errorMessage;
}

const openSolscan = (data) => {
  modalsStore.closeModal();
  const url = 'https://solscan.io/token/'
  if (data && data?.mint) {
    window.open(url + data.mint, '_blank');
  }
}

const handleLogoChoose = (data) => {
  if (!data) return;

  tokenData.value.logo = data.url || '';
  tokenData.value.logo_file = data.file || null;

  clearErrorMessage('logo')
}

const handleWalletConnect = async () => {
  try {
    await connectHandler();

    clearErrorMessage('ownership')
  } catch (e) {
    console.log('wallet connection error:', e);
  }
}

const clearErrorMessage = (field) => {
  if (errors.value[field]) {
    errors.value[field] = '';
  }
}

const areFieldsValid = () => {
  Object.keys(errors.value).forEach((key) => {
    const val = tokenData.value[key];
    const maxSolanaSupply = 18446744073.709551615;

    if (['name', 'ticker'].includes(key) && !val.length) {
      errors.value[key] = 'Minimum 3 characters.';
    } else if (key === 'supply') {
      if (Number(val) > maxSolanaSupply) {
        errors.value[key] = 'The max number of Solana chain tokens is \n18 446 744 073.709551615';
      } else if (Number(val) === 0) {
        errors.value[key] = 'Enter token amount';
      } else if (Number(val) < 1000000) {
        errors.value[key] = 'The min number of Solana chain tokens is 1 000 000';
      }
    }
    // else if (key === 'ownership' && !val.length) {
    //   errors.value[key] = 'Connect Ownership wallet';
    // }
  })

  return Object.keys(errors.value).every((key) => !errors.value[key].length)
}

const handleTokenCreate = async () => {
  if (!areFieldsValid()) return;
  try {
    if (!address.value?.length) await connectHandler();
    const decimals = Number(tokenData.value.decimals);
    const tokenName = tokenData.value.name;
    const tokenSymbol = tokenData.value.ticker;
    const tokenSupply = BigInt(tokenData.value.supply);
    const lighthouseTestUrl = 'https://fit-aardvark-makzo.lighthouseweb3.xyz/ipfs/';
    let metadataURI = "https://your-json-url.json";
    let imageUrl = '';
    const newFormData = new FormData();

    if (tokenData.value.logo_file) {
      newFormData.append('image', tokenData.value.logo_file);

      const imageResp = await UploadImage(newFormData);

      if (imageResp?.data) {
        // imageUrl = lighthouseTestUrl + imageResp.data.url;
        imageUrl = lighthouseTestUrl + imageResp.data.cid;
      }
    }

    // uri JSON
    const metadata = {
      name: tokenName,
      symbol: tokenSymbol,
      description: tokenData.value.description,
      image: imageUrl,
      external_url: tokenData.value.social_links.website,
      extensions: {
        twitter: tokenData.value.social_links.twitter,
        telegram: tokenData.value.social_links.telegram,
        discord: tokenData.value.social_links.discord
      }
    }

    const metadataResp = await UploadMetadata(metadata);

    if (metadataResp?.data) {
      // metadataURI = metadataResp.data.metadata_url || '';
      metadataURI = lighthouseTestUrl + metadataResp.data.cid || '';
    }

    const RPC_URL = `${import.meta.env.VITE_HELIUS_RPC_URL}API_KEY`;
    // const RPC_URL = 'https://api.devnet.solana.com';
    const connection = new Connection(RPC_URL, 'confirmed');

    const umi = createUmi(RPC_URL);
    const userLibPublicKey = fromWeb3JsPublicKey(publicKey.value);
    umi.use(signerIdentity(createNoopSigner(userLibPublicKey)));

    const mintKeypair = Keypair.generate();
    const lamports = await connection.getMinimumBalanceForRentExemption(MINT_SIZE);

    const transaction = new Transaction();

    transaction.add(
      SystemProgram.createAccount({
        fromPubkey: publicKey.value,
        newAccountPubkey: mintKeypair.publicKey,
        space: MINT_SIZE,
        lamports,
        programId: TOKEN_PROGRAM_ID,
      })
    );

    transaction.add(
      createInitializeMintInstruction(
        mintKeypair.publicKey,
        decimals,
        publicKey.value,
        publicKey.value,
      )
    );

    const createMetadataIx = createV1(umi, {
      mint: fromWeb3JsPublicKey(mintKeypair.publicKey),
      authority: umi.identity,
      name: tokenName,
      symbol: tokenSymbol,
      uri: metadataURI,
      sellerFeeBasisPoints: 0,
      tokenStandard: TokenStandard.Fungible,
      isMutable: !tokenData.value.immutable,
    });

    transaction.add(toWeb3JsInstruction(createMetadataIx.getInstructions()[0]));

    const ata = await getAssociatedTokenAddress(mintKeypair.publicKey, publicKey.value);

    transaction.add(
      createAssociatedTokenAccountInstruction(
        publicKey.value,
        ata,
        publicKey.value,
        mintKeypair.publicKey
      )
    );

    transaction.add(
      createMintToInstruction(
        mintKeypair.publicKey,
        ata,
        publicKey.value,
        tokenSupply * BigInt(10 ** decimals)
      )
    );

    if (tokenData.value.immutable) {
      const updateIx = updateV1(umi, {
        mint: fromWeb3JsPublicKey(mintKeypair.publicKey),
        authority: umi.identity,
        newUpdateAuthority: none(),
        isMutable: some(false),
      });
      transaction.add(toWeb3JsInstruction(updateIx.getInstructions()[0]));
    }

    if (tokenData.value.fixed_supply) {
      transaction.add(
        createSetAuthorityInstruction(
          mintKeypair.publicKey,
          publicKey.value,
          AuthorityType.MintTokens,
          null
        )
      );
    }

    if (tokenData.value.revoke_freeze) {
      transaction.add(
        createSetAuthorityInstruction(
          mintKeypair.publicKey,
          publicKey.value,
          AuthorityType.FreezeAccount,
          null
        )
      );
    }

    const { blockhash } = await connection.getLatestBlockhash('finalized');
    transaction.recentBlockhash = blockhash;
    transaction.feePayer = publicKey.value;

    transaction.partialSign(mintKeypair);

    const signedTx = await walletProvider.value.signTransaction(transaction);
    const sig = await connection.sendRawTransaction(signedTx.serialize());

    trackGoogleTagEvent('Create Token');

    modalsStore.modalData.title = `Token ${tokenName} created successfully`
    modalsStore.modalData.type = 'create-confirmation'
    modalsStore.modalData.action = 'confirmation';
    modalsStore.modalData.item = {mint: mintKeypair.publicKey.toBase58()};
    modalsStore.modalData.is_open = true;
  } catch (error) {
    console.error("ERROR:", error);
  }
};

useHeaderRefresh(() => handlePageRefresh(true));

watch(() => [route.name, route.params.campaign_id], async () => {
  smartCampaignStore.clearStore();

  await handlePageRefresh();
})

watch(() => userStore.isUserAuth, async (newVal) => {
  if (newVal) {
    await handlePageRefresh(false, true);
  }
})

watch(() => address.value, async (newVal) => {
  if (newVal?.length) {
    tokenData.value.ownership = newVal;
  }
});

onMounted(async () => {
  if (route.params?.campaign_id !== 'create' && !userStore.isUserAuth) {
    await router.push({params: {campaign_id: 'create'}});
  }
  await handlePageRefresh();
});
</script>
<style scoped lang="scss">
.token-create {
  &__block {
    margin-top: 32px;
  }

  &__btns {
    display: flex;
    gap: 12px;

  }

  &__back {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #374151;
    cursor: pointer;
  }

  &__start {
    margin: 32px 0;
  }

  &__estimate {
    width: 100%;
    height: 40px;
    border-radius: 8px;
    background: rgba(209, 213, 219, 0.25);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 20px 0;
    color: #64748B;
  }

  &__estimate {
    margin-top: 32px;
  }

  &__alert {
    margin-top: 32px;
  }

  &__desktop {
    width: 100%;
    display: flex;
    gap: 10px;
    justify-content: space-between;

    &_left {
      max-width: 669px;
      width: 100%;
    }

    &_right {
      width: 100%;
      max-width: 371px;
    }

    &_settings {
      margin-top: 20px;
      display: flex;
      flex-direction: column;
      gap: 12px;


      & .setting {
        border-radius: 8px;
        border: 1px solid #D1D5DB;
        padding: 16px;
        display: flex;
        gap: 16px;
        flex-direction: column;
        justify-content: space-between;

        &__top {
          display: flex;
          justify-content: space-between;
        }

        &__left {
          display: flex;
          flex-direction: column;
          gap: 8px;
          max-width: 227px;
        }

        &__social-links {
          width: 100%;
          display: flex;
          flex-direction: column;
          gap: 12px;

          ::v-deep(.base-input__input) {
            gap: 6px;
          }

          & .icon {
            min-width: 16px;
            max-width: 16px;
            min-height: 16px;
            max-height: 16px;

            ::v-deep(path) {
              fill: #9CA3AF;
            }
          }
        }
      }
    }
  }

  &__mobile {
    display: none;
  }

  &.pull-down {
    & .token-create {
      &__block {
        &.transaction-setup, &.target {
          opacity: .5;
          pointer-events: none;
        }
      }
    }
  }

}

.modal-stop-campaign {
  max-width: 480px;
}

.create-confirmation {
  max-width: 500px;
}

@media (max-width: 1200px) {
  .token-create {
    &__desktop {
      display: none;
    }

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
