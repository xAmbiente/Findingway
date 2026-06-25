export interface ListingEntry {
	creator: string;
	description: string;
	duty: string;
	expires: string;
	minIlvl: string;
	party: Party[];
	pfTags?: string;
	updated: string;
	world: string;
}

export interface Party {
	filled: boolean;
	title: string;
	type: 'dps' | 'healer' | 'none' | 'tank';
}
