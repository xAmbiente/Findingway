import { scrape } from '#lib/scraper/xivpfScraper';
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

		const timeOfScrape = Date.now();
		const scrapeJson = await scrape();

		this.container.client.emit(FindingwayEvents.PostTea, { entries: scrapeJson.TheEpicOfAlexander, timeOfScrape });
		this.container.client.emit(FindingwayEvents.PostUcob, { entries: scrapeJson.TheUnendingCoilOfBahamut, timeOfScrape });
		this.container.client.emit(FindingwayEvents.PostUwu, { entries: scrapeJson.TheWeaponsRefrain, timeOfScrape });
	}
}
