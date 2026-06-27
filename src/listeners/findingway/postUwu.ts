import { ChannelType } from '#lib/generated/prisma-client/enums';
import { type FindingwayEvents, type PostMessagePayload } from '#lib/util/constants';
import { sendPfPost } from '#utils/functions/sendPfPost';
import { Listener } from '@sapphire/framework';

export class UserListener extends Listener<typeof FindingwayEvents.PostUwu> {
	public override async run(payload: PostMessagePayload) {
		return sendPfPost(payload, ChannelType.TheWeaponsRefrain);
	}
}
