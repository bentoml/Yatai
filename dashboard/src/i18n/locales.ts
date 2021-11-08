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
    'build_at': {
        en: 'Build At',
        cn: '编译时间',
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
    'deployment snapshot': {
        en: 'Deployment Snapshot',
        cn: '部署快照',
    },
    'snapshot': {
        en: 'Snapshot',
        cn: '快照',
    },
    'bento': {
        en: 'Bento',
        cn: 'Bento',
    },
    'bento version': {
        cn: 'Bento 版本',
        en: 'Bento Version',
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
    'status': {
        cn: '状态',
        en: 'Status',
    },
    'status name': {
        cn: '状态名称',
        en: 'Status Name',
    },
    'image build status': {
        cn: '镜像编译状态',
        en: 'Image Build Status',
    },
    'user profile': {
        cn: '用户资料',
        en: 'User Profile',
    },
    'api token': {
        cn: 'API 令牌',
        en: 'API Token',
    },
    'production': {
        cn: '正式',
        en: 'Production',
    },
    'stable': {
        cn: '稳定',
        en: 'Stable',
    },
    'canary': {
        cn: '金丝雀',
        en: 'Canary',
    },
    'canary rules': {
        cn: '金丝雀规则',
        en: 'Canary Rules',
    },
    'weight': {
        cn: '权重',
        en: 'Weight',
    },
    'header': {
        cn: 'Header',
        en: 'Header',
    },
    'cookie': {
        cn: 'Cookie',
        en: 'Cookie',
    },
    'header value': {
        cn: 'Header 值',
        en: 'Header value',
    },
    'add app deployment canary rule': {
        cn: '增加灰度发布规则',
        en: 'Add app deployment gray rule',
    },
    'create canary deploy': {
        cn: '创建灰度发布',
        en: 'Create gray deploy',
    },
    'create deploy': {
        cn: '创建部署',
        en: 'Create deploy',
    },
    'force': {
        cn: '强制',
        en: 'Force',
    },
    'deployment detail': {
        cn: '部署详情',
        en: 'Deployment Detail',
    },
    'visibility': {
        cn: '可见性',
        en: 'Visibility',
    },
    'non-deployed': {
        cn: '未部署',
        en: 'Non Deployed',
    },
    'unhealthy': {
        cn: '不健康',
        en: 'Unhealthy',
    },
    'failed': {
        cn: '失败',
        en: 'Failed',
    },
    'deploying': {
        cn: '部署中',
        en: 'Deploying',
    },
    'running': {
        cn: '运行中',
        en: 'Running',
    },
    'unknown': {
        cn: '未知',
        en: 'Unknown',
    },
    'replicas': {
        cn: '副本',
        en: 'Replicas',
    },
    'type': {
        cn: '类型',
        en: 'Type',
    },
    'view log': {
        cn: '查看日志',
        en: 'View Log',
    },
    'view terminal history': {
        cn: '查看终端操作记录',
        en: 'View Terminal History',
    },
    'download log': {
        cn: '下载日志',
        en: 'Download Log',
    },
    'log forward': {
        cn: '日志正序',
        en: 'Log Forward',
    },
    'stop scroll': {
        cn: '停止滚动',
        en: 'Stop Scroll',
    },
    'scroll': {
        cn: '滚动',
        en: 'Scroll',
    },
    'all pods': {
        cn: '所有的 Pod',
        en: 'All Pods',
    },
    'lines': {
        cn: '行数',
        en: 'Lines',
    },
    'realtime': {
        cn: '实时',
        en: 'Realtime',
    },
    'max lines': {
        cn: '最大行数',
        en: 'Max Lines',
    },
    'advanced': {
        cn: '高级',
        en: 'Advanced',
    },
    'cpu limits': {
        cn: 'CPU 资源限制',
        en: 'CPU Resources Limits',
    },
    'gpu limits': {
        cn: 'GPU 资源限制',
        en: 'GPU Resources Limits',
    },
    'gpu requests': {
        cn: 'GPU 资源请求',
        en: 'GPU Resources Requests',
    },
    'cpu': {
        cn: 'CPU',
        en: 'CPU',
    },
    'gpu': {
        cn: '显卡',
        en: 'GPU',
    },
    'memory': {
        cn: '内存',
        en: 'Memory',
    },
    'memory limits': {
        cn: '内存资源限制',
        en: 'Memory Resources Limits',
    },
    'cpu requests': {
        cn: 'CPU 资源请求',
        en: 'CPU Resources Requests',
    },
    'memory requests': {
        cn: '内存资源请求',
        en: 'Memory Resources Requests',
    },
    'add': {
        cn: '添加',
        en: 'Add',
    },
    'action': {
        cn: '行为',
        en: 'Action',
    },
    'Terminating': {
        cn: '结束中',
        en: 'Terminating',
    },
    'ContainerCreating': {
        cn: '创建中',
        en: 'Creating',
    },
    'pending': {
        cn: '等待中',
        en: 'Pending',
    },
    'building': {
        cn: '编译中',
        en: 'Building',
    },
    'Pending': {
        cn: '等待中',
        en: 'Pending',
    },
    'Running': {
        cn: '运行中',
        en: 'Running',
    },
    'success': {
        cn: '成功',
        en: 'Succeeded',
    },
    'Succeeded': {
        cn: '成功',
        en: 'Succeeded',
    },
    'Failed': {
        cn: '失败',
        en: 'Failed',
    },
    'Unknown': {
        cn: '未知',
        en: 'Unknown',
    },
    'start time': {
        cn: '启动时间',
        en: 'Start Time',
    },
    'end time': {
        cn: '结束时间',
        en: 'End Time',
    },
    'terminal': {
        cn: '终端',
        en: 'Terminal',
    },
    'operation': {
        en: 'Operation',
        cn: '操作',
    },
    'pod': {
        cn: 'Pod ',
        en: 'Pod',
    },
    'container': {
        cn: '容器',
        en: 'Container',
    },
    'playback operation': {
        cn: '回放操作',
        en: 'Playback Operation',
    },
    'no event': {
        cn: '暂无事件',
        en: 'No Event',
    },
    'event': {
        cn: '事件',
        en: 'Event',
    },
    'events': {
        cn: '事件列表',
        en: 'Events',
    },
    'node': {
        cn: '节点',
        en: 'Node',
    },
    'settings': {
        cn: '设置',
        en: 'Settings',
    },
    'regexp': {
        cn: '正则表达式',
        en: 'Regexp',
    },
    'contains': {
        cn: '包含',
        en: 'Contains',
    },
    'not contains': {
        cn: '不包含',
        en: 'Not Contains',
    },
    'clear': {
        cn: '清空',
        en: 'Clear',
    },
    'and': {
        cn: '且',
        en: 'And',
    },
    'add filter condition': {
        cn: '添加过滤条件',
        en: 'Add Filter Condition',
    },
    'match': {
        cn: '匹配',
        en: 'Match',
    },
    'not match': {
        cn: '不匹配',
        en: 'Not Match',
    },
    'advanced search': {
        cn: '高级搜索',
        en: 'Advanced Search',
    },
    'search': {
        cn: '搜索',
        en: 'Search',
    },
    'please enter keywords': {
        cn: '请输入关键词',
        en: 'Please enter keywords',
    },
    'yatai component': {
        cn: 'Yatai 组件',
        en: 'Yatai Component',
    },
    'deployed': {
        cn: '部署完毕',
        en: 'Deployed',
    },
    'uninstalled': {
        cn: '卸载完毕',
        en: 'Uninstalled',
    },
    'superseded': {
        cn: '已废弃',
        en: 'Superseded',
    },
    'uninstalling': {
        cn: '卸载中',
        en: 'Uninstalling',
    },
    'pending-install': {
        cn: '等待安装',
        en: 'Pending Install',
    },
    'pending-upgrade': {
        cn: '等待升级',
        en: 'Pending Upgrade',
    },
    'pending-rollback': {
        cn: '等待回滚',
        en: 'Pending Rollback',
    },
    'logging': {
        cn: '日志',
        en: 'Logging',
    },
    'monitoring': {
        cn: '监控',
        en: 'Monitoring',
    },
    'please install yatai component first': {
        cn: '请先安装 Yatai {{0}} 组件',
        en: 'Please install yatai component {{0}} first',
    },
    'monitor': {
        cn: '监控器',
        en: 'Monitor',
    },
    'upgrade': {
        cn: '升级',
        en: 'Upgrade',
    },
    'upgrade to sth': {
        cn: '升级到 {{0}}',
        en: 'Upgrade to {{0}}',
    },
    'do you want to upgrade yatai component sth to some version': {
        cn: '你想把 Yatai 组件 {{0}} 升级到 {{1}} 版本吗？',
        en: 'Do you want to upgrade yatai component {{0}} to {{1}} version?',
    },
    'do you want to delete yatai component sth': {
        cn: '你想删除 Yatai 组件 {{0}} 吗？',
        en: 'Do you want to delete yatai component {{0}}?',
    },
    'cancel': {
        cn: '取消',
        en: 'Cancel',
    },
    'ok': {
        cn: '确定',
        en: 'Okay',
    },
    'delete': {
        cn: '删除',
        en: 'Delete',
    },
    'kube resources': {
        cn: 'K8S 资源列表',
        en: 'K8S Resources',
    },
    'helm release name': {
        cn: 'HELM 发布名称',
        en: 'HELM Release Name',
    },
    'helm chart name': {
        cn: 'HELM Chart 名称',
        en: 'HELM Chart Name',
    },
    'helm chart description': {
        cn: 'HELM Chart 详情',
        en: 'HELM Chart Description',
    },
    'model': {
        cn: '模型',
        en: 'Model',
    },
    'model versions': {
        cn: '模型版本',
        en: 'Model Versions',
    },
    'name or email': {
        cn: '用户名或者邮箱地址',
        en: 'Name or Email',
    },
    'password': {
        cn: '密码',
        en: 'Password',
    },
}

export const locales: { [key in keyof typeof locales0]: ILocaleItem } = locales0
