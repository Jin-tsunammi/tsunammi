// export const BASE_URL = import.meta.env.VITE_BASE_URL;
import Raydium from "../../public/images/raydium-icon.webp";
import PumpFun from "../../public/images/pumpfun-icon.webp";

export const BASE_URL = import.meta.env.VITE_BASE_URL;
export const HELIUS_RPC_URL = 'https://mainnet.helius-rpc.com/?api-key=' + import.meta.env.VITE_HELIUS_RPC_API_KEY;
export const SOL_SCAN_BASE_URL = 'https://solscan.io/account/';
export const SOLANA_MINT = 'So11111111111111111111111111111111111111112';
export const NANO_IN_SECOND = 1000000000;
export const VALIDATORS_LIST = [
    {name: 'Jito', val: 'jito', image: 'https://www.jito.wtf/jitoBig.png'},
    {name: 'Public Solana Node', val: 'custom', image: 'https://solana.com/src/img/branding/solanaLogoMark.svg'},
]
export const DEX_LIST = [
    {name: 'Raydium', id: 1, val: 'raydium', image: Raydium},
    {name: 'PumpFun', id: 2, val: 'pumpfun', image: PumpFun},
]