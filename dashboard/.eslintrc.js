module.exports = {
    parser: '@typescript-eslint/parser',
    parserOptions: {
        jsx: true,
        useJSXTextNode: true,
        project: './tsconfig.json',
    },
    env: {
        browser: true,
        jasmine: true,
        es6: true,
        node: true,
        jest: true,
    },
    extends: [
        'airbnb-typescript',
        'airbnb/hooks',
        'plugin:react/recommended',
        'plugin:react-hooks/recommended',
        'plugin:@typescript-eslint/recommended',
        'plugin:prettier/recommended',
        'prettier',
    ],
    plugins: ['@typescript-eslint', 'react', 'react-hooks', 'baseui'],
    rules: {
        'quotes': ['error', 'single', { avoidEscape: true }],
        'require-atomic-updates': 'off',
        'react-hooks/rules-of-hooks': 'error',
        'react-hooks/exhaustive-deps': 'error',
        '@typescript-eslint/no-unused-vars': 'error',
        'no-console': "on",
        '@typescript-eslint/explicit-function-return-type': 'off',
        '@typescript-eslint/explicit-module-boundary-types': 'off',
        '@typescript-eslint/naming-convention': [
            'error',
            {
                selector: 'interface',
                format: ['PascalCase'],
                custom: {
                    regex: '^I[A-Z]',
                    match: true,
                },
            },
            // Allow camelCase variables (23.2), PascalCase variables (23.8), and UPPER_CASE variables (23.10)
            {
                selector: 'variable',
                format: ['camelCase', 'PascalCase', 'UPPER_CASE'],
                leadingUnderscore: 'allow',
                trailingUnderscore: 'allow',
            },
            // Allow camelCase functions (23.2), and PascalCase functions (23.8)
            {
                selector: 'function',
                format: ['camelCase', 'PascalCase'],
                leadingUnderscore: 'allow',
                trailingUnderscore: 'allow',
            },
            // Airbnb recommends PascalCase for classes (23.3), and although Airbnb does not make TypeScript recommendations, we are assuming this rule would similarly apply to anything "type like", including interfaces, type aliases, and enums
            {
                selector: 'typeLike',
                format: ['PascalCase'],
            },
        ],
        'react/react-in-jsx-scope': 'off',
        'react/display-name': 'off',
        'react/prop-types': 'off',
        'react/jsx-indent': ['error', 4],
        'react/jsx-indent-props': ['error', 4],
        'react/jsx-one-expression-per-line': 'off',
        'react/jsx-wrap-multilines': 'off',
        'react/no-array-index-key': 'off',
        'react/require-default-props': ['error', { ignoreFunctionalComponents: true }],
        'react-hooks/exhaustive-deps': [
            'warn',
            {
                additionalHooks:
                    '(useApp|useAppLoading|useUser|useUserLoading|useCurrentUser|useSidebar|useSidebarSize)',
            },
        ],
        'import/prefer-default-export': 'off',
        'jsx-a11y/control-has-associated-label': 'off',
        'jsx-a11y/click-events-have-key-events': 'off',
        'camelcase': 'off',
        'no-plusplus': 'off',
        'no-underscore-dangle': 'off',
        'eqeqeq': ['error', 'always'],
        'prettier/prettier': [
            'error',
            {
                endOfLine: 'auto',
            },
        ],
        'baseui/deprecated-theme-api': 'warn',
        'baseui/deprecated-component-api': 'warn',
        'baseui/no-deep-imports': 'warn',
    },
    settings: {
        react: {
            version: 'detect',
        },
    },
    globals: {
        asciinema: 'readonly',
    },
    overrides: [
        {
            files: ['*.js'],
            rules: {
                '@typescript-eslint/no-var-requires': 'off',
            },
        },
    ],
    ignorePatterns: ['.eslintrc.js', 'craco.config.js'],
}
