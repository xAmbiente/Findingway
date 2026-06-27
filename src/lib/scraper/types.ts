export interface ListingEntry {
	creator: string;
	description: string;
	duty: string;
	expires: number;
	minIlvl: string;
	party: Party[];
	pfTags?: string;
	updated: number;
	world: string;
}

export interface Party {
	filled: boolean;
	title: string;
	type: 'dps' | 'healer' | 'healerDps' | 'none' | 'tank' | 'tankDps' | 'tankHealer' | 'tankHealerDps';
}
