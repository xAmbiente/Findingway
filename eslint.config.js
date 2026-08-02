import common from 'eslint-config-neon/common';
import prettier from 'eslint-config-neon/prettier';
import typescript from 'eslint-config-neon/typescript';
import eslintPluginPrettierRecommended from 'eslint-plugin-prettier/recommended';
import merge from 'lodash.merge';

/**
 * @type {import('@typescript-eslint/utils').TSESLint.FlatConfig.ConfigArray}
 */
const config = [
	{
		ignores: ['.yarn/**', 'src/lib/generated/**']
	},
	...[...common, ...typescript, ...prettier].map((config) =>
		merge(config, {
			files: ['src/**/*.ts', 'prisma/initial-data-seed.ts'],
			languageOptions: {
				parserOptions: {
					project: 'tsconfig.eslint.json'
				}
			},
			rules: {
				'@typescript-eslint/consistent-type-definitions': 'off',
				'import-x/order': 'off'
			}
		})
	),
	eslintPluginPrettierRecommended
];

export default config;
