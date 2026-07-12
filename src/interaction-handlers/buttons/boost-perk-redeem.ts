import { BOOST_PERK_REDEEM_BUTTON_CUSTOM_ID } from '#utils/constants/boost-perk';
import { runBoostPerkRedeemFlow } from '#utils/functions/boostPerk';
import { ApplyOptions } from '@sapphire/decorators';
import { InteractionHandler, InteractionHandlerTypes } from '@sapphire/framework';
import type { ButtonInteraction } from 'discord.js';

@ApplyOptions<InteractionHandler.Options>({
	interactionHandlerType: InteractionHandlerTypes.Button
})
export class UserInteractionHandler extends InteractionHandler {
	public override parse(interaction: ButtonInteraction) {
		if (interaction.customId === BOOST_PERK_REDEEM_BUTTON_CUSTOM_ID && interaction.inCachedGuild()) {
			return this.some();
		}

		return this.none();
	}

	public override async run(interaction: ButtonInteraction) {
		if (!interaction.inCachedGuild()) return;

		await interaction.deferUpdate();
		await runBoostPerkRedeemFlow(interaction);
	}
}
