import type { FindingwayEvents, PostMessagePayload } from '#lib/util/constants';
import { Listener } from '@sapphire/framework';

export class UserListener extends Listener<typeof FindingwayEvents.PostMerc> {
	public override async run({ entries }: PostMessagePayload) {
		return undefined;
	}
}
