export interface ILocaleItem {
    cn: string
    en: string
}

const locales0 = {
    'homepage': {
        en: 'Homepage',
        cn: '主页',
    },
    'overview': {
        en: 'Overview',
        cn: '概览',
    },
    'user': {
        en: 'User',
        cn: '用户',
    },
    'creator': {
        en: 'Creator',
        cn: '创建者',
    },
    'created_at': {
        en: 'Created At',
        cn: '创建时间',
    },
    'user group': {
        en: 'User Group',
        cn: '用户组',
    },
    'kube_config': {
        en: 'Kubernetes Configuration',
        cn: 'Kubernetes Configuration',
    },
    'config': {
        en: 'Configuration',
        cn: '配置',
    },
    'role': {
        en: 'Role',
        cn: '角色',
    },
    'developer': {
        en: 'Developer',
        cn: '开发者',
    },
    'guest': {
        en: 'Guest',
        cn: '访客',
    },
    'admin': {
        en: 'Admin',
        cn: '管理员',
    },
    'create': {
        en: 'Create',
        cn: '创建',
    },
    'submit': {
        en: 'Submit',
        cn: '提交',
    },
    'name': {
        en: 'Name',
        cn: '名称',
    },
    'description': {
        en: 'Description',
        cn: '描述',
    },
    'member': {
        en: 'Member',
        cn: '成员',
    },
    'create sth': {
        en: 'Create {{0}}',
        cn: '创建{{0}}',
    },
    'sth list': {
        en: '{{0}} List',
        cn: '{{0}}列表',
    },
    'select sth': {
        en: 'Select {{0}}',
        cn: '选择{{0}}',
    },
    'organization': {
        en: 'Organization',
        cn: '组织',
    },
    'cluster': {
        en: 'Cluster',
        cn: '集群',
    },
    'deployment': {
        en: 'Deployment',
        cn: '部署',
    },
    'bundle': {
        en: 'Bundle',
        cn: '包',
    },
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
    'no data': {
        cn: '暂无数据',
        en: 'No Data',
    },
    'latest version': {
        cn: '最新版',
        en: 'Latest Version',
    },
    'version': {
        cn: '版本',
        en: 'Version',
    },
}

export const locales: { [key in keyof typeof locales0]: ILocaleItem } = locales0
