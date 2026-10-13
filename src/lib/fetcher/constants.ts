import type { ListingData, ListingEntry } from '#lib/fetcher/types';
import os from 'node:os';

export const mercKeywords = ['merc', 'mercenary', 'carry', 'gil', 'payment', 'paid', 'boost', 'sell', 'selling', 'pay', 'fee', 'wage', 'commission'];

export const fetchUrl = new URL('https://better-xivpf.fly.dev/api/listings');
fetchUrl.searchParams.set('excludePrivate', 'true');
fetchUrl.searchParams.set('limit', '1000');
fetchUrl.searchParams.set('dataCenter', 'Light');

// eslint-disable-next-line @typescript-eslint/no-inferrable-types
const packageVersion: string = '[VI]{{inject}}[/VI]';

export const fetchHeaders = new Headers();
fetchHeaders.append(
	'User-Agent',
	`Findingway/${packageVersion} (undici) ${os.platform()}/${os.release()} (https://github.com/xAAmbiente/Findingway)`
);

export function mapListingDataToListingEntry(data: ListingData): ListingEntry {
	return {
		creator: data.creator,
		description: data.description,
		duty: data.duty,
		expires: Math.floor(new Date(data.expires).getTime() / 1_000),
		minIlvl: data.minItemLevel === 0 ? null : data.minItemLevel,
		party: data.party,
		tags: data.tags,
		updated: Math.floor(new Date(data.updated).getTime() / 1_000),
		world: data.world
	};
}
