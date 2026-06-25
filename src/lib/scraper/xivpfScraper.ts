import { ChannelType } from '#lib/generated/prisma-client/enums';
import { SearchTerms } from '#lib/scraper/constants';
import type { ListingEntry, Party } from '#lib/scraper/types';
import { chromium, type Browser, type BrowserContext, type Page } from 'playwright';

export async function scrape(): Promise<{
	[ChannelType.TheEpicOfAlexander]: ListingEntry[];
	[ChannelType.TheUnendingCoilOfBahamut]: ListingEntry[];
	[ChannelType.TheWeaponsRefrain]: ListingEntry[];
}> {
	const browser = await chromium.launch();

	try {
		const [teaListings, ucobListings, uwuListings] = await Promise.all([
			scrapeListingsForDuty(browser, ChannelType.TheEpicOfAlexander),
			scrapeListingsForDuty(browser, ChannelType.TheUnendingCoilOfBahamut),
			scrapeListingsForDuty(browser, ChannelType.TheWeaponsRefrain)
		]);

		return {
			[ChannelType.TheEpicOfAlexander]: sortAndLimitListingsByUpdated(teaListings),
			[ChannelType.TheUnendingCoilOfBahamut]: sortAndLimitListingsByUpdated(ucobListings),
			[ChannelType.TheWeaponsRefrain]: sortAndLimitListingsByUpdated(uwuListings)
		};
	} finally {
		await browser.close();
	}
}

function sortAndLimitListingsByUpdated(listings: ListingEntry[]): ListingEntry[] {
	return [...listings]
		.sort((left, right) => parseRelativeUpdatedAgeSeconds(left.updated) - parseRelativeUpdatedAgeSeconds(right.updated))
		.slice(0, 7);
}

function parseRelativeUpdatedAgeSeconds(updated: string): number {
	const normalized = updated.trim().toLowerCase();
	if (!normalized || normalized === 'just now' || normalized === 'now') {
		return 0;
	}

	const compactMatch = /^(?<amount>\d+)\s*(?<unit>[dhmsw])\s*ago$/.exec(normalized);
	if (compactMatch?.groups) {
		const amount = Number.parseInt(compactMatch.groups.amount, 10);
		const multiplier = getTimeUnitSeconds(compactMatch.groups.unit);
		return amount * multiplier;
	}

	const verboseMatch = /^(?<amount>\d+)\s*(?<unit>second|minute|hour|day|week)s?\s*ago$/.exec(normalized);
	if (verboseMatch?.groups) {
		const amount = Number.parseInt(verboseMatch.groups.amount, 10);
		const unit = verboseMatch.groups.unit;
		const shortUnit = unit === 'second' ? 's' : unit === 'minute' ? 'm' : unit === 'hour' ? 'h' : unit === 'day' ? 'd' : 'w';
		const multiplier = getTimeUnitSeconds(shortUnit);
		return amount * multiplier;
	}

	return Number.POSITIVE_INFINITY;
}

function getTimeUnitSeconds(unit: string): number {
	if (unit === 's') {
		return 1;
	}

	if (unit === 'm') {
		return 60;
	}

	if (unit === 'h') {
		return 60 * 60;
	}

	if (unit === 'd') {
		return 60 * 60 * 24;
	}

	return 60 * 60 * 24 * 7;
}

async function scrapeListingsForDuty(browser: Browser, duty: keyof typeof SearchTerms): Promise<ListingEntry[]> {
	const context = await browser.newContext();

	try {
		const page = await initWebpage(context);
		await filterListings(page, duty);
		return await getListingAsJson(page);
	} finally {
		await context.close();
	}
}

async function initWebpage(browserOrContext: BrowserContext): Promise<Page> {
	const page = await browserOrContext.newPage();

	await page.goto('https://xivpf.com/listings');
	await page.locator('body').click();
	await page.locator('#data-centre-filter').selectOption('Light');
	await page.getByText('advanced').click();
	await page.getByLabel('Categories Duty Roulette').selectOption('HighEndDuty');
	await page.getByText('advanced').click();

	return page;
}

async function filterListings(page: Page, filterTerm: keyof typeof SearchTerms) {
	const searchbox = page.getByRole('searchbox', { name: 'search' });
	await searchbox.click();
	await searchbox.press('ControlOrMeta+A');
	await searchbox.pressSequentially(SearchTerms[filterTerm], { delay: 30 });
}

async function getListingAsJson(page: Page): Promise<ListingEntry[]> {
	await page.waitForSelector('#listings .listing');

	const listings = await page.locator('#listings > .listing').evaluateAll((listingNodeElements) => {
		return listingNodeElements.map((listingNode) => {
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

			const normalizedMinIlvl = minIlvl === '0' ? 'unspecified' : minIlvl;

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
				minIlvl: normalizedMinIlvl,
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
