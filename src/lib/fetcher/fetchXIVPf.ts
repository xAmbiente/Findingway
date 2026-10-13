import { fetchHeaders, fetchUrl, mapListingDataToListingEntry, mercKeywords } from '#lib/fetcher/constants';
import { type ListingData, type ListingEntry, type ListingResponse } from '#lib/fetcher/types';
import { ChannelType } from '#lib/generated/prisma-client/enums';
import { fetch, FetchMethods, FetchResultTypes } from '@sapphire/fetch';

export async function fetchXIVPf(): Promise<{ [key in ChannelType]: ListingEntry[] }> {
	const listingResponse = await fetch<ListingResponse>(
		fetchUrl,
		{
			method: FetchMethods.Get,
			headers: fetchHeaders
		},
		FetchResultTypes.JSON
	);

	const teaListings = filterListingsForDuty(listingResponse, 'The Epic of Alexander (Ultimate)');
	const ucobListings = filterListingsForDuty(listingResponse, 'The Unending Coil of Bahamut (Ultimate)');
	const uwuListings = filterListingsForDuty(listingResponse, "The Weapon's Refrain (Ultimate)");
	const mercListings = filterListingsForMercKeywords(listingResponse);

	return {
		[ChannelType.TheEpicOfAlexander]: sortAndLimitListingsByUpdated(uniqueifyListings(teaListings)),
		[ChannelType.TheUnendingCoilOfBahamut]: sortAndLimitListingsByUpdated(uniqueifyListings(ucobListings)),
		[ChannelType.TheWeaponsRefrain]: sortAndLimitListingsByUpdated(uniqueifyListings(uwuListings)),
		[ChannelType.Mercantile]: sortAndLimitListingsByUpdated(uniqueifyListings(mercListings))
	};
}

function filterListingsForDuty(listingResponse: ListingResponse, dutyName: string): ListingEntry[] {
	return listingResponse.data.filter((listing) => listing.duty === dutyName).map(mapListingDataToListingEntry);
}

function filterListingsForMercKeywords(listingResponse: ListingResponse): ListingEntry[] {
	const listingsByTerm: ListingData[][] = [];

	for (const keyword of mercKeywords) {
		listingsByTerm.push(listingResponse.data.filter((listing) => listing.description.toLowerCase().includes(keyword.toLowerCase())));
	}

	return listingsByTerm.flatMap((data) => data.map(mapListingDataToListingEntry));
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
