# Tsunammi Front-End

Web interface for managing market-making operations: campaign setup, wallet management, CEX API connection, token
creation, and action history tracking.

## What the project includes

- Market-making scenario management (for example, `Target Pull Up`, `Target Drop`, `Smart Buyback`).
- Wallet sections: creation, import, top-up, history, and project management.
- Token workflows: token creation, liquidity pools, liquidity burn, and history.
- CEX connection and tools for exchange-side operations.
- Vue 3 UI with Pinia for state management and Vite for build tooling.

## Technologies

- `Vue 3`
- `Vue Router`
- `Pinia`
- `Vite`
- `Axios`
- `Sass`

## Getting started

To launch `Target Pull Up` and `Target Drop`, complete the setup flow below first:

1. Go to **Wallets -> Manage Wallets** and create a **Wallet Pool**.
2. Add wallets to this pool:
    - generate new wallets, or
    - import existing wallets (manually or from `.txt`, `.csv`, `.xlsx`).
3. Go to **Wallets -> Connect CEX API** and connect your exchange API account.
4. Open **Wallets -> Top up from CEX** and create a funding request:
    - select connected CEX account;
    - select Wallet Pool;
    - choose wallet quantity (all/half/custom);
    - set min/max deposit per wallet and confirm top up.
5. After wallets are funded, open **Market making -> Target Pull Up** or **Target Drop** and configure campaign
   parameters.

In campaign setup, the selected Wallet Pool is used as the budget source, so pools without required balance cannot be
used to run campaigns.

## Building the Project

To build the project for deployment (for example, to GitHub Pages or any static hosting), follow these steps:

Create .env file
```bash
cp .env.example env
   ```

1. **Install dependencies**

   Make sure you have [Node.js](https://nodejs.org/) (version 20+ or higher recommended) and [npm](https://www.npmjs.com/) installed.

   ```bash
   npm install
   ```

2. **Build the project**

   This will generate a `dist` folder with the production-ready static files:

   ```bash
   npm run build
   ```

3. **Preview the production build locally (optional)**

   You can preview the built project locally to ensure everything works as expected:

   ```bash
   npm run preview
   ```

4. **Deploy**

   Upload the contents of the `dist` folder to your preferred static hosting service (such as GitHub Pages, Vercel,
   Netlify, or your own server).

5. **DEV start**

    ```bash
   npm run dev
   ```

---
