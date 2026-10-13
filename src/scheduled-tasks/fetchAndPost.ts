import { fetchXIVPf } from '#lib/fetcher/fetchXIVPf';
import { FindingwayEvents } from '#utils/constants';
import { ApplyOptions } from '@sapphire/decorators';
import { ScheduledTask } from '@sapphire/plugin-scheduled-tasks';
import { Status } from 'discord.js';

@ApplyOptions<ScheduledTask.Options>({
	pattern: '*/3 * * * *', // Every 3 minutes
	customJobOptions: {
		removeOnComplete: true
	}
})
export class UserScheduledTask extends ScheduledTask {
	public override async run() {
		// If the websocket isn't ready, skip for now
		if (this.container.client.ws.status !== Status.Ready || !this.container.client.user) {
			return;
		}

		// If there are no channels configured for any server, skip (this is probably in development mode)
		const channels = await this.container.prisma.channel.findMany();

		if (channels.length === 0) {
			return;
		}

		const timeOfFetch = Date.now();
		const listings = await fetchXIVPf();

		this.container.client.emit(FindingwayEvents.PostTea, { entries: listings.TheEpicOfAlexander, timeOfFetch });
		this.container.client.emit(FindingwayEvents.PostUcob, { entries: listings.TheUnendingCoilOfBahamut, timeOfFetch });
		this.container.client.emit(FindingwayEvents.PostUwu, { entries: listings.TheWeaponsRefrain, timeOfFetch });
		this.container.client.emit(FindingwayEvents.PostMerc, { entries: listings.Mercantile, timeOfFetch });
	}
}
