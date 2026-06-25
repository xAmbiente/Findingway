import type { ChannelType } from '#lib/generated/prisma-client/enums';
import type { PostMessagePayload } from '#utils/constants';
import { buildPfPost } from '#utils/functions/buildPfPost';
import { resolveOnErrorCodes } from '#utils/functions/resolveOnErrorCodes';
import { container } from '@sapphire/framework';
import { MessageFlags, RESTJSONErrorCodes, type MessageCreateOptions, type MessageEditOptions } from 'discord.js';

export async function sendPfPost(payload: PostMessagePayload, type: ChannelType): Promise<void> {
	const channelsToPostIn = await container.prisma.channel.findMany({
		where: {
			type
		}
	});

	for (const channel of channelsToPostIn) {
		if (!channel.enabled) continue;

		const guild = await resolveOnErrorCodes(
			container.client.guilds.fetch(channel.guildId),
			RESTJSONErrorCodes.UnknownGuild,
			RESTJSONErrorCodes.InvalidGuild
		);

		if (!guild) continue;

		const guildChannel = await resolveOnErrorCodes(guild.channels.fetch(channel.channelId), RESTJSONErrorCodes.UnknownChannel);

		if (!guildChannel?.isSendable()) continue;

		const messagePayload: MessageCreateOptions | MessageEditOptions = {
			flags: [MessageFlags.IsComponentsV2],
			components: [await buildPfPost(payload, type)]
		};

		if (channel.messageId) {
			const oldPostedMessage = await resolveOnErrorCodes(guildChannel.messages.fetch(channel.messageId), RESTJSONErrorCodes.UnknownMessage);

			await oldPostedMessage?.edit(messagePayload as MessageEditOptions);
		} else {
			const newMessage = await guildChannel.send(messagePayload as MessageCreateOptions);
			await container.prisma.channel.update({
				where: {
					channelId: guildChannel.id,
					type_guildId: {
						type,
						guildId: guild.id
					}
				},
				data: {
					messageId: newMessage.id
				}
			});
		}
	}
}
