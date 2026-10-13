import { FindingwayCommand } from '#lib/extensions/FindingwayComand';
import { fetchXIVPf } from '#lib/fetcher/fetchXIVPf';
import { FindingwayEvents } from '#utils/constants';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder, resolveKey } from '@sapphire/plugin-i18next';
import { ApplicationIntegrationType, MessageFlags } from 'discord.js';

@ApplyOptions<ChatInputCommand.Options>({
	description: 'Triggers an immediate fetch'
})
@RegisterChatInputCommand((builder) =>
	applyLocalizedBuilder(builder, 'commands/fetch:root').setIntegrationTypes(ApplicationIntegrationType.GuildInstall)
)
export class SlashCommands extends FindingwayCommand {
	public override async chatInputRun(interaction: ChatInputCommand.Interaction<'cached'>) {
		await interaction.deferReply({ flags: MessageFlags.Ephemeral });

		const started = await resolveKey(interaction, 'commands/fetch:started');
		await interaction.editReply(started);

		const timeOfFetch = Date.now();
		const fetchJson = await fetchXIVPf();

		this.container.client.emit(FindingwayEvents.PostTea, { entries: fetchJson.TheEpicOfAlexander, timeOfFetch });
		this.container.client.emit(FindingwayEvents.PostUcob, { entries: fetchJson.TheUnendingCoilOfBahamut, timeOfFetch });
		this.container.client.emit(FindingwayEvents.PostUwu, { entries: fetchJson.TheWeaponsRefrain, timeOfFetch });
		this.container.client.emit(FindingwayEvents.PostMerc, { entries: fetchJson.Mercantile, timeOfFetch });

		const response = await resolveKey(interaction, 'commands/fetch:success');
		return interaction.editReply(response);
	}
}
