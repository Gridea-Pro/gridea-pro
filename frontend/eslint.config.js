import js from '@eslint/js'
import globals from 'globals'
import tseslint from 'typescript-eslint'
import pluginVue from 'eslint-plugin-vue'
import eslintConfigPrettier from 'eslint-config-prettier'

export default tseslint.config(
    // 忽略目录
    {
        ignores: [
            'dist/**',
            'dist-electron/**',
            'build/**',
            'node_modules/**',
            'wailsjs/**',
            '*.min.js',
            'public/**',
        ],
    },

    // 基础配置
    js.configs.recommended,

    // TypeScript 配置
    ...tseslint.configs.recommended,

    // Vue 3 配置
    ...pluginVue.configs['flat/recommended'],

    // Prettier 配置（禁用冲突规则）
    eslintConfigPrettier,

    // 全局配置
    {
        languageOptions: {
            ecmaVersion: 'latest',
            sourceType: 'module',
            globals: {
                ...globals.browser,
                ...globals.node,
                ...globals.es2021,
            },
        },
    },

    // Vue 文件特殊配置
    {
        files: ['**/*.vue'],
        languageOptions: {
            parserOptions: {
                parser: tseslint.parser,
            },
        },
    },

    // 自定义规则
    {
        rules: {
            'vue/multi-word-component-names': 'off',
            'vue/require-default-prop': 'off',
            '@typescript-eslint/no-explicit-any': 'warn',
            '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
            'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
            'no-debugger': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
            'vue/no-v-html': 'off',
            '@typescript-eslint/ban-ts-comment': 'off',
        },
    },
)
