import Cookies from 'js-cookie';

class CookieManager {
    static setItem(
        name,
        value,
        expires = Date.now() + 31449600, // 1 year by default
    ) {
        const hostname = window.location.hostname;

        const isLocalhost = hostname === 'localhost' || /^192\.\d+\.\d+\.\d+$/.test(hostname);
        Cookies.set(name, value, {
            domain: hostname.endsWith('.tonversity.com') ? '.tonversity.com' : undefined,
            expires: new Date(expires * 1000),
            secure: !isLocalhost,
            sameSite: !isLocalhost ? 'None' : 'Lax',
        })
    }

    static getItem(name) {
        return Cookies.get(name);
    }

    static removeItem(name) {
        const hostname = window.location.hostname;
        const domain = hostname.endsWith('.tonversity.com') ? '.tonversity.com' : undefined;

        Cookies.remove(name, { domain });
    }

    static clear() {
        const cookies = Object.keys(Cookies.get());
        for (const cookie of cookies) {
            Cookies.remove(cookie, { domain: window.location.hostname === 'localhost' ? 'localhost' : '.tonversity.com' });
        }
    }

    static getAll() {
        return Cookies.get();
    }
}

export default CookieManager;
