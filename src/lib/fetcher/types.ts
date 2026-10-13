export interface ListingEntry {
	creator: string;
	description: string;
	duty: string;
	expires: number;
	minIlvl: number | null;
	party: Party[];
	tags: string[];
	updated: number;
	world: string;
}

export interface ListingResponse {
	data: ListingData[];
	meta: ListingMeta;
	success: boolean;
}

export type PartyRoles = 'dps' | 'healer' | 'none' | 'tank';

export interface Party {
	filled: boolean;
	job: string | null;
	roles: PartyRoles[];
}

export interface ListingData {
	/**
	 * The conditions required for the listing
	 */
	conditions: string;
	/**
	 * The creator of the listing
	 */
	creator: string;
	/**
	 * The data center of the listing
	 */
	dataCenter: string;
	/**
	 * The description of the listing
	 */
	description: string;
	/**
	 * The duty associated with the listing
	 */
	duty: string;
	/**
	 * A string timestamp representing the expiration time of the listing
	 */
	expires: string;
	/**
	 * The unique identifier of the listing
	 */
	id: string;
	/**
	 * Indicates whether the listing is private
	 */
	isPrivate: boolean;
	/**
	 * The minimum item level required for the listing
	 */
	minItemLevel: number;
	/**
	 * The objective of the listing
	 */
	objective: string;
	/**
	 * The party composition for the listing
	 */
	party: Party[];
	/**
	 * The category of the listing on the platform
	 */
	pfCategory: string;
	/**
	 * A string timestamp representing when the listing was scraped by better-xivpf from xivpf
	 */
	scrapedAt: string;
	/**
	 * The tags associated with the listing
	 */
	tags: string[];
	/**
	 * A string timestamp representing the last update time
	 */
	updated: string;
	/**
	 * The world associated with the listing
	 */
	world: string;
}

interface ListingMeta {
	count: number;
	limit: number;
}
