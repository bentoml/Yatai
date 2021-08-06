export interface ILocaleItem {
    cn: string
    en: string
}

const locales0 = {
    'register': {
        en: 'Register',
        cn: '注册',
    },
    'login': {
        en: 'Login',
        cn: '登录',
    },
    'logout': {
        en: 'Logout',
        cn: '登出',
    },
    'sth required': {
        cn: '需要填写{{0}}',
        en: '{{0}} was required',
    },
}

export const locales: { [key in keyof typeof locales0]: ILocaleItem } = locales0
