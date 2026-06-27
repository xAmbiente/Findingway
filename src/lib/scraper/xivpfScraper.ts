import { ChannelType } from '#lib/generated/prisma-client/enums';
import type { ListingEntry, Party } from '#lib/scraper/types';
import { container } from '@sapphire/framework';
import { chromium, type Browser, type BrowserContext, type Page } from 'playwright';

const mercKeywords = ['merc', 'mercenary', 'carry', 'gil', 'payment', 'paid', 'boost', 'sell', 'selling', 'pay', 'fee', 'wage', 'commission'];
export async function scrape(): Promise<{ [key in ChannelType]: ListingEntry[] }> {
	const browser = await chromium.launch();

	try {
		const teaListings = await scrapeListingsForFilterTerm(browser, 'The Epic of Alexander');
		const ucobListings = await scrapeListingsForFilterTerm(browser, 'The Unending Coil of Bahamut');
		const uwuListings = await scrapeListingsForFilterTerm(browser, "The Weapon's Refrain");
		const mercListings = await scrapeListingsForFilterTerms(browser, mercKeywords);

		return {
			[ChannelType.TheEpicOfAlexander]: sortAndLimitListingsByUpdated(uniqueifyListings(teaListings)),
			[ChannelType.TheUnendingCoilOfBahamut]: sortAndLimitListingsByUpdated(uniqueifyListings(ucobListings)),
			[ChannelType.TheWeaponsRefrain]: sortAndLimitListingsByUpdated(uniqueifyListings(uwuListings)),
			[ChannelType.Mercantile]: sortAndLimitListingsByUpdated(
				uniqueifyListings(
					mercListings.filter((listing) => mercKeywords.some((keyword) => listing.description.toLowerCase().includes(keyword)))
				)
			)
		};
	} finally {
		await browser.close();
	}
}

async function scrapeListingsForFilterTerm(browser: Browser, filterTerm: string): Promise<ListingEntry[]> {
	const context = await browser.newContext();

	try {
		const page = await initWebpage(context);
		await filterListings(page, filterTerm);
		return await getListingAsJson(page);
	} finally {
		await context.close();
	}
}

async function scrapeListingsForFilterTerms(browser: Browser, filterTerms: string[]): Promise<ListingEntry[]> {
	const listingsByTerm: ListingEntry[][] = [];
	for (const filterTerm of filterTerms) {
		listingsByTerm.push(await scrapeListingsForFilterTerm(browser, filterTerm));
	}

	return listingsByTerm.flat();
}

async function initWebpage(browserOrContext: BrowserContext): Promise<Page> {
	const page = await browserOrContext.newPage();

	await page.goto('https://xivpf.com/listings');
	await page.locator('body').click();
	await page.locator('#data-centre-filter').selectOption('Light');
	const advancedFiltersToggle = page.locator('.filter-controls').getByText('advanced');
	await advancedFiltersToggle.click();
	await page.getByLabel('Categories Duty Roulette').selectOption('HighEndDuty');
	await advancedFiltersToggle.click();

	return page;
}

async function filterListings(page: Page, filterTerm: string) {
	const searchbox = page.getByRole('searchbox', { name: 'search' });
	await searchbox.click();
	await searchbox.press('ControlOrMeta+A');
	await searchbox.pressSequentially(filterTerm, { delay: 30 });
}

async function getListingAsJson(page: Page): Promise<ListingEntry[]> {
	await page.waitForSelector('#listings');

	const listingElements = page.locator('#listings > .listing');
	if ((await listingElements.count()) === 0) {
		return [];
	}

	const nowUnixTimestamp = Math.floor(Date.now() / 1_000);

	const listings = await listingElements.evaluateAll((listingNodeElements) => {
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

			return {
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
		});
	});

	return listings.map((listing) => {
		container.logger.debug(
			`Building post for ${listing.duty} from ${listing.creator} with updated ${listing.updated} and parsed ${parseRelativeUpdatedToUnixTimestamp(listing.updated, nowUnixTimestamp)}`
		);
		return {
			...listing,
			expires: parseRelativeExpiresToUnixTimestamp(listing.expires, nowUnixTimestamp),
			updated: parseRelativeUpdatedToUnixTimestamp(listing.updated, nowUnixTimestamp)
		};
	});
}

function uniqueifyListings(listings: ListingEntry[]): ListingEntry[] {
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

function sortAndLimitListingsByUpdated(listings: ListingEntry[]): ListingEntry[] {
	return [...listings].sort((left, right) => right.updated - left.updated).slice(0, 7);
}

function parseRelativeUpdatedToUnixTimestamp(updated: string, nowUnixTimestamp: number): number {
	return nowUnixTimestamp - parseRelativeUpdatedAgeSeconds(updated);
}

function parseRelativeUpdatedAgeSeconds(updated: string): number {
	const normalized = updated.trim().toLowerCase();
	if (!normalized || normalized === 'just now' || normalized === 'now') {
		return 0;
	}

	if (normalized === 'a second ago' || normalized === 'an second ago') {
		return getTimeUnitSeconds('s');
	}

	if (normalized === 'a minute ago' || normalized === 'an minute ago') {
		return getTimeUnitSeconds('m');
	}

	if (normalized === 'an hour ago' || normalized === 'a hour ago') {
		return getTimeUnitSeconds('h');
	}

	const compactMatch = /^(?<amount>\d+)\s*(?<unit>[hms])\s*ago$/.exec(normalized);
	if (compactMatch?.groups) {
		const amount = Number.parseInt(compactMatch.groups.amount, 10);
		const multiplier = getTimeUnitSeconds(compactMatch.groups.unit);
		return amount * multiplier;
	}

	const verboseMatch = /^(?<amount>\d+)\s*(?<unit>second|minute|hour)s?\s*ago$/.exec(normalized);
	if (verboseMatch?.groups) {
		const amount = Number.parseInt(verboseMatch.groups.amount, 10);
		const unit = verboseMatch.groups.unit;
		let shortUnit: string;
		if (unit === 'second') {
			shortUnit = 's';
		} else if (unit === 'minute') {
			shortUnit = 'm';
		} else {
			shortUnit = 'h';
		}

		const multiplier = getTimeUnitSeconds(shortUnit);
		return amount * multiplier;
	}

	return 0;
}

function parseRelativeExpiresToUnixTimestamp(expires: string, nowUnixTimestamp: number): number {
	return nowUnixTimestamp + parseRelativeExpiresInSeconds(expires);
}

function parseRelativeExpiresInSeconds(expires: string): number {
	const normalized = expires.trim().toLowerCase();
	if (!normalized) {
		return 0;
	}

	if (normalized === 'in an hour' || normalized === 'in a hour') {
		return getTimeUnitSeconds('h');
	}

	if (normalized === 'in a minute' || normalized === 'in one minute') {
		return getTimeUnitSeconds('m');
	}

	const minutesMatch = /^in\s+(?<amount>\d+)\s+minutes?$/.exec(normalized);
	if (minutesMatch?.groups) {
		const amount = Number.parseInt(minutesMatch.groups.amount, 10);
		return amount * getTimeUnitSeconds('m');
	}

	return 0;
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

	return 0;
}
