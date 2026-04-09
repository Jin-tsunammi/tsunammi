import { initializeApp } from "firebase/app";
import { getAuth } from "firebase/auth";


const firebaseConfig = {
    apiKey: "AIzaSyBhfnBvxaJZWYRwUcR3gPuVQ_7U7BuFW5U",
    authDomain: "mm-test-40ccf.firebaseapp.com",
    projectId: "mm-test-40ccf",
    storageBucket: "mm-test-40ccf.firebasestorage.app",
    messagingSenderId: "620570627907",
    appId: "1:620570627907:web:ee283e5a71acfb880069f6"
};

const app = initializeApp(firebaseConfig);

const auth = getAuth(app);
export default auth;