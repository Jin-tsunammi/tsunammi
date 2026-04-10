import { createApp } from 'vue';
import {createPinia} from 'pinia';
import App from './App.vue';
import {router} from "./routes/index.js";
import { VueDatePicker } from '@vuepic/vue-datepicker';
import '@vuepic/vue-datepicker/dist/main.css'

const pinia = createPinia();


createApp(App)
    .use(pinia)
    .use(router)
    .component('VueDatePicker', VueDatePicker)
    .mount('#app')
