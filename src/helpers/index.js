import {useToastStore} from "../store/toastStore.js";
import {useTokensStore} from "../store/tokensStore.js";

export function decodeJwt(token) {
  const base64Url = token.split('.')[1]; // Get payload part
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  return JSON.parse(atob(base64));
}

export const getRandomNumber = () => {
  return Math.floor(10000 + Math.random() * 90000);
}

export const formatDate = (str) => {
  if (!str) return '';

  const localDate = new Date(str);
  const hh = localDate.getHours();
  const mm = String(localDate.getMinutes()).padStart(2, '0');
  const day = String(localDate.getDate()).padStart(2, '0');
  const month = String(localDate.getMonth() + 1).padStart(2, '0');
  const year = String(localDate.getFullYear()).slice(-2);
  const formatedDate = `${day}.${month}.${year}`;

  return {date: formatedDate, time: `${hh}:${mm}`}
}

export function daysSince(dateString) {
  if (!dateString) return '0'
  const pastDate = new Date(dateString);
  const today = new Date();

  const diffMs = today - pastDate;

  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));

  const result = diffDays < 0 ? 0 : diffDays;

  return result === 1 ? `${result} day` : `${result} days`;
}

export function formatWalletAddress(address, symbols=10) {
  if (!address) return '';

  const start = address.slice(0, symbols);
  const end = address.slice(-symbols);

  return `${start}...${end}`;
}

export const formatText = (val) => {
  if (!val) return '';

  const normalizedUnderscore = val.replaceAll('_', ' ');
  const firstLetter = normalizedUnderscore.slice(0, 1).toUpperCase();
  const restText = normalizedUnderscore.slice(1);

  return firstLetter + restText;
}

export function formatAmount(value, length=4) {
  if (value === null || value === undefined) return ''

  let str = String(value).replace('.', ',')

  const parts = str.split(',')
  const start = parts[0]
  let end = parts[1] || ''

  end = end.slice(0, length)

  end = end.replace(/0+$/, '')

  return end ? `${start},${end}` : start
}

export function errorToast(errorMessage) {
  const toastStore = useToastStore();
  const message = errorMessage ? formatText(errorMessage) : 'Something went wrong';
  toastStore.error({text: message});
  console.error(message);
}

export function toDynamicFix(value) {
  if (isNaN(+value)) return 0;

  const num = parseFloat(String(value));
  if (Number.isInteger(num)) return num;

  const stringValue = num.toString();
  const decimalIndex = stringValue.indexOf(".");
  if (decimalIndex === -1) return num;

  const fractionalPart = stringValue.slice(decimalIndex + 1);

  const firstSignificantIndex = fractionalPart.search(/[^0]/);
  if (firstSignificantIndex === -1) return num;

  const fix = firstSignificantIndex + 4;

  const multiplier = 10 ** fix;
  return Math.round(num * multiplier) / multiplier;
}
export const fetchSolTokenMetadata = async (mintSet, storage) => {
  if (!mintSet?.size) return;
  const url = `${import.meta.env.VITE_HELIUS_RPC_URL}${import.meta.env.VITE_HELIUS_RPC_API_KEY}`;

      try {
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        jsonrpc: '2.0',
        id: 'metadata',
        method: 'getAssetBatch',
        params: {
          ids: Array.from(mintSet),
          displayOptions: { showFungible: true }
        }
      })
    });

    const { result } = await response.json();
    const filteredList = result.filter(asset => asset !== null);

    filteredList.forEach(token => {
      if (!storage.value[token.id]) {
        storage.value[token.id] = {
          name: token.content?.metadata?.name || '',
          symbol: token.content?.metadata?.symbol || '',
          image: token.content?.links?.image || '',
        }
      }
    })

  } catch (e) {
    console.error('Helius Error:', e);
  }
};

export const isTokenPictureValid = (token_mint, isToken=true) => {
  const tokensStore = useTokensStore();

  const imageUrl = isToken ? tokensStore.solTokensData[token_mint]?.image : token_mint;
  if (typeof imageUrl !== 'string' || !imageUrl.trim()) return false;

  const url = imageUrl.trim();
  if (url.startsWith('data:image/')) {
    return url.length > 'data:image/'.length;
  }
  if (url.startsWith('http://') || url.startsWith('https://')) {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  return url.startsWith('../') || url.startsWith('/public/');
}

export const calculateBudget = (balance, percent) => {
  if (!balance || !percent) return 0;

  return (balance / 100) * percent;
}

export function sanitizeFilename(name) {
  const base = String(name ?? "export")
    .trim()
    .replace(/[\\/:*?"<>|]+/g, "_")
    .replace(/\s+/g, " ")
    .slice(0, 80);

  return base.length ? base : "export";
}

export function getToday() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}
