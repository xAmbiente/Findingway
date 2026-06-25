import { ChannelType } from '#lib/generated/prisma-client/enums';
import { SearchTerms } from '#lib/scraper/constants';
import type { ListingEntry, Party } from '#lib/scraper/types';
import { chromium, type Page } from 'playwright';

export async function scrape(): Promise<{
	[ChannelType.TheEpicOfAlexander]: ListingEntry[];
	[ChannelType.TheUnendingCoilOfBahamut]: ListingEntry[];
	[ChannelType.TheWeaponsRefrain]: ListingEntry[];
}> {
	const page = await initWebpage();

	const [teaListings, ucobListings, uwuListings] = await Promise.all([
		filterListings(page, ChannelType.TheEpicOfAlexander),
		filterListings(page, ChannelType.TheUnendingCoilOfBahamut),
		filterListings(page, ChannelType.TheWeaponsRefrain)
	]);

	const [teaListingsJson, ucobListingsJson, uwuListingsJson] = await Promise.all([
		getListingAsJson(teaListings),
		getListingAsJson(ucobListings),
		getListingAsJson(uwuListings)
	]);

	return {
		[ChannelType.TheEpicOfAlexander]: teaListingsJson,
		[ChannelType.TheUnendingCoilOfBahamut]: ucobListingsJson,
		[ChannelType.TheWeaponsRefrain]: uwuListingsJson
	};
}

async function initWebpage(): Promise<Page> {
	const browser = await chromium.launch();
	const page = await browser.newPage();

	await page.goto('https://xivpf.com/listings');
	await page.locator('body').click();
	await page.locator('#data-centre-filter').selectOption('Light');
	await page.getByText('advanced').click();
	await page.getByLabel('Categories Duty Roulette').selectOption('HighEndDuty');
	await page.getByText('advanced').click();

	await page.screenshot({
		path: `./screenshots/init-web-page-${Date.now()}.png`
	});

	return page;
}

async function filterListings(page: Page, filterTerm: keyof typeof SearchTerms) {
	const searchbox = page.getByRole('searchbox', { name: 'search' });
	await searchbox.click();
	await searchbox.press('ControlOrMeta+A');
	await searchbox.pressSequentially(SearchTerms[filterTerm], { delay: 30 });

	await page.screenshot({
		path: `./screenshots/filter-listings-${Date.now()}.png`
	});

	return page;
}

async function getListingAsJson(page: Page): Promise<ListingEntry[]> {
	await page.waitForSelector('#listings .listing');

	const listings = await page.locator('#listings > .listing').evaluateAll((listingNodeements) => {
		return listingNodeements.map((listingNode) => {
			const duty = listingNode.querySelector('.duty.cross')?.textContent?.trim() ?? '';
			const creator = listingNode.querySelector('.item.creator .text')?.textContent?.trim() ?? '';
			const world = listingNode.querySelector('.item.world .text')?.textContent?.trim() ?? '';
			const expires = listingNode.querySelector('.item.expires .text')?.textContent?.trim() ?? '';
			const updated = listingNode.querySelector('.item.updated .text')?.textContent?.trim() ?? '';

			const minIlvl =
				Array.from(listingNode.querySelectorAll('.stat'))
					.map((statNode) => statNode)
					.find((statEl) => {
						const statName = statEl.querySelector('.name')?.textContent?.trim() ?? '';
						return statName === 'Min IL';
					})
					?.querySelector('.value')
					?.textContent?.trim() ?? '';

			const descriptionEl = listingNode.querySelector('.description');
			const pfTagEl = descriptionEl?.querySelector('span');
			const pfTags = pfTagEl?.textContent?.trim() ?? '';

			let description = descriptionEl?.textContent?.trim() ?? '';
			if (pfTags) {
				description = description.replace(pfTags, '').trim();
			}

			const partySlots = Array.from(listingNode.querySelectorAll('.party .slot')).map((slotNode) => {
				const slotEl = slotNode;
				const hasTank = slotEl.classList.contains('tank');
				const hasHealer = slotEl.classList.contains('healer');
				const hasDps = slotEl.classList.contains('dps');
				let type: Party['type'] = 'none';
				if (hasTank && hasHealer && hasDps) {
					type = 'tankHealerDps';
				} else if (hasTank && hasHealer) {
					type = 'tankHealer';
				} else if (hasTank && hasDps) {
					type = 'tankDps';
				} else if (hasHealer && hasDps) {
					type = 'healerDps';
				} else if (hasTank) {
					type = 'tank';
				} else if (hasHealer) {
					type = 'healer';
				} else if (hasDps) {
					type = 'dps';
				}

				return {
					type,
					filled: slotEl.classList.contains('filled'),
					title: slotEl.getAttribute('title')?.trim() ?? ''
				};
			});

			const totalText = listingNode.querySelector('.party .total')?.textContent?.trim() ?? '';
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

	const seenCreators = new Set<string>();
	return listings.filter((listing) => {
		const normalizedCreator = listing.creator.trim().toLowerCase();
		if (seenCreators.has(normalizedCreator)) {
			return false;
		}

		seenCreators.add(normalizedCreator);
		return true;
	});
}
