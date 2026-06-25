import type { ListingEntry, Party } from '#lib/scraper/types';
import { expect, test } from '@playwright/test';

test('debug playwright', async ({ page }) => {
	await page.goto('https://xivpf.com/listings');

	await page.locator('body').click();
	await page.locator('#data-centre-filter').selectOption('Light');
	await page.getByText('advanced').click();
	await page.getByLabel('Categories Duty Roulette').selectOption('HighEndDuty');
	await page.getByText('advanced').click();

	const searchbox = page.getByRole('searchbox', { name: 'search' });
	await searchbox.click();
	await searchbox.press('ControlOrMeta+A');
	await searchbox.pressSequentially("The Weapon's Refrain", { delay: 30 });

	await expect(page.locator('#listings .listing .duty.cross').first()).toContainText("The Weapon's Refrain");

	await page.waitForSelector('#listings .listing');

	const record = await page.locator('#listings > .listing').evaluateAll((listingElements) => {
		return listingElements.map((listingNode) => {
			const listingEl = listingNode;
			const duty = listingEl.querySelector('.duty.cross')?.textContent?.trim() ?? '';
			const creator = listingEl.querySelector('.item.creator .text')?.textContent?.trim() ?? '';
			const world = listingEl.querySelector('.item.world .text')?.textContent?.trim() ?? '';
			const expires = listingEl.querySelector('.item.expires .text')?.textContent?.trim() ?? '';
			const updated = listingEl.querySelector('.item.updated .text')?.textContent?.trim() ?? '';

			const minIlvl =
				Array.from(listingEl.querySelectorAll('.stat'))
					.map((statNode) => statNode)
					.find((statEl) => {
						const statName = statEl.querySelector('.name')?.textContent?.trim() ?? '';
						return statName === 'Min IL';
					})
					?.querySelector('.value')
					?.textContent?.trim() ?? '';

			const descriptionEl = listingEl.querySelector('.description');
			const pfTagEl = descriptionEl?.querySelector('span');
			const pfTags = pfTagEl?.textContent?.trim() ?? '';

			let description = descriptionEl?.textContent?.trim() ?? '';
			if (pfTags) {
				description = description.replace(pfTags, '').trim();
			}

			const partySlots = Array.from(listingEl.querySelectorAll('.party .slot')).map((slotNode) => {
				const slotEl = slotNode;
				let type: Party['type'] = 'none';
				if (slotEl.classList.contains('tank')) {
					type = 'tank';
				} else if (slotEl.classList.contains('healer')) {
					type = 'healer';
				} else if (slotEl.classList.contains('dps')) {
					type = 'dps';
				}

				return {
					type,
					filled: slotEl.classList.contains('filled'),
					title: slotEl.getAttribute('title')?.trim() ?? ''
				};
			});

			const totalText = listingEl.querySelector('.party .total')?.textContent?.trim() ?? '';
			const totalRegex = /^\d+\s*\/\s*(?<totalAvailable>\d+)$/;
			const totalMatch = totalRegex.exec(totalText);
			const totalAvailable = totalMatch?.groups?.totalAvailable ? Number.parseInt(totalMatch.groups.totalAvailable, 10) : 8;
			if (totalAvailable < 8) {
				const unavailableCount = 8 - totalAvailable;
				for (let index = 0; index < unavailableCount; index += 1) {
					partySlots.push({
						type: 'none',
						filled: true,
						title: 'UNAVAILABLE'
					});
				}
			}

			const result: ListingEntry = {
				creator,
				duty,
				description,
				expires,
				minIlvl,
				pfTags,
				party: partySlots,
				updated,
				world
			};
			return result;
		});
	});

	console.log(JSON.stringify(record));
});
