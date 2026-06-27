import { ChannelType } from '#lib/generated/prisma-client/browser';
import type { FindingwayEvents, PostMessagePayload } from '#lib/util/constants';
import { sendPfPost } from '#utils/functions/sendPfPost';
import { Listener } from '@sapphire/framework';

export class UserListener extends Listener<typeof FindingwayEvents.PostTea> {
	public override async run(payload: PostMessagePayload) {
		return sendPfPost(payload, ChannelType.TheEpicOfAlexander);
	}
}
