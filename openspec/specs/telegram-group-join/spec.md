# telegram-group-join Specification

## Purpose
How someone already sitting in a Telegram chat gets into that chat's Repa group, and the limits
on what the bot will say while helping them.

## Requirements

### Requirement: Anyone in a connected chat can get into its group

The bot SHALL respond to a `/join` command in a connected chat with that group's invite link, so
a member can join without the founder distributing anything by hand.

#### Scenario: The bot replies with the group's invite

- **GIVEN** a Telegram chat connected to a Repa group
- **WHEN** someone in the chat sends `/join`
- **THEN** the bot replies with a link that opens that group's join flow
- **AND** the reply names the group

#### Scenario: The command works for any member of the chat, not only the founder

- **GIVEN** a connected chat
- **WHEN** a person who did not create the group sends `/join`
- **THEN** they receive the same invite link

#### Scenario: An unconnected chat is told what to do instead

- **GIVEN** a chat not connected to any group
- **WHEN** someone sends `/join`
- **THEN** the bot explains that the chat is not connected and how to connect it
- **AND** does not reveal any group

#### Scenario: The command works with a bot mention

- **GIVEN** a connected chat
- **WHEN** someone sends `/join@` followed by the bot's username
- **THEN** the bot replies as it would to a bare `/join`

### Requirement: A join from a chat is attributed to Telegram

An invite the bot hands out SHALL mark Telegram as the channel, so joins arriving from chats are
distinguishable from joins arriving from shared cards.

#### Scenario: The bot's link carries the Telegram channel

- **WHEN** the bot replies to `/join`
- **THEN** the link it sends marks Telegram as the channel it came from

### Requirement: The bot tells people the command exists

The bot SHALL list what it can do in response to `/help`, and SHALL respond to an unrecognised
command with the same list rather than staying silent.

#### Scenario: Help lists the commands

- **WHEN** someone sends `/help` in a chat with the bot
- **THEN** the reply names the join, status, and connect commands

#### Scenario: An unknown command is answered, not ignored

- **WHEN** someone sends a command the bot does not recognise
- **THEN** the bot replies with what it can do

#### Scenario: Ordinary chat messages are not answered

- **WHEN** someone sends a message that is not a command
- **THEN** the bot says nothing

### Requirement: Connecting a chat announces the join command once

When a chat is connected to a group, the bot SHALL say in that chat that members can now use the
join command — the one moment the whole chat is reliably looking.

#### Scenario: The connect confirmation mentions joining

- **WHEN** a chat is successfully connected to a group
- **THEN** the bot's confirmation tells members they can use the join command

### Requirement: Connecting an already-connected chat reports the conflict

If a chat is already connected to a group, a connect attempt for a *different* group SHALL be
refused with an explanation rather than silently re-pointing the chat.

#### Scenario: A second group cannot take over a connected chat

- **GIVEN** a chat connected to group A
- **WHEN** someone runs connect with group B's code in that chat
- **THEN** the bot refuses and says the chat is already connected
- **AND** the chat remains connected to group A

#### Scenario: Reconnecting the same group is not an error

- **GIVEN** a chat connected to group A
- **WHEN** someone runs connect with group A's own code in that chat
- **THEN** the bot confirms the chat is connected to that group

### Requirement: The bot never exposes results in a chat

Nothing the bot says in response to a join, help, or connect command SHALL contain a member's
attributes, percentages, or any individual's results.

#### Scenario: The join reply contains no results

- **WHEN** the bot replies to `/join` for a group whose season has revealed
- **THEN** the reply contains no attribute, percentage, or member name from the results
