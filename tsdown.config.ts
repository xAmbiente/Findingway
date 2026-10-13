import { defineConfig } from 'tsdown';
import Replace from 'unplugin-replace/rolldown';
import packageJson from './package.json' with { type: 'json' };

export default defineConfig({
	clean: true,
	entry: ['src/**', '!src/locales/**'],
	dts: false,
	root: 'src',
	unbundle: true,
	minify: false,
	deps: { skipNodeModulesBundle: true },
	sourcemap: true,
	target: 'es2024',
	tsconfig: 'src/tsconfig.json',
	treeshake: true,
	format: 'esm',
	plugins: [
		Replace({
			values: [
				{
					find: /\[VI\]{{inject}}\[\/VI\]/,
					replacement: packageJson.version
				}
			]
		})
	]
});
