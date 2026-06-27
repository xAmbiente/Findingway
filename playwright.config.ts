import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	fullyParallel: true,
	retries: 0,
	workers: 4,
	use: { trace: 'on-first-retry' },
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		}
	]
});
