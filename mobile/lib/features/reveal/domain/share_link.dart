/// The channel a share went out through. Mirrors the backend's allowed set
/// (`revealsvc.NormalizeShareChannel`); anything else is recorded as `other` server-side.
enum ShareChannel {
  /// The card image, shared through the OS share sheet.
  card,

  /// Posted into the group's Telegram chat.
  telegram,

  /// A bare link, shared without a card.
  link;

  String get wireValue => name;
}

/// Builds the link that goes out with a shared card.
///
/// Carries the group's invite code so the card works as a standalone invitation, and a channel
/// marker so shares and the joins attributed to them can be compared. The code stays the path
/// segment the router already handles, and the marker is a query parameter the server strips
/// during normalisation — so the link resolves whether or not a client understands it.
String buildShareLink({
  required String inviteCode,
  required ShareChannel channel,
  String? referrerId,
}) {
  final base = 'https://repa.app/join/$inviteCode?s=${channel.wireValue}';
  // A group's code identifies the group, not the inviter, so the referral reward needs the
  // sharer carried separately. Omitted when unknown — the link still joins.
  return referrerId == null || referrerId.isEmpty ? base : '$base&ref=$referrerId';
}

/// The text that accompanies a shared card.
String buildShareText({
  required String inviteCode,
  required ShareChannel channel,
  String? referrerId,
}) =>
    'Моя репа 🍆 Залетай: '
    '${buildShareLink(inviteCode: inviteCode, channel: channel, referrerId: referrerId)}';

/// Reads the channel marker out of an incoming link, so an arrival can be attributed to the
/// share it came from. Returns null when the link carries no marker.
ShareChannel? shareChannelFromLink(String url) {
  final match = RegExp(r'[?&]s=([a-z]+)').firstMatch(url);
  if (match == null) return null;
  final value = match.group(1);
  for (final c in ShareChannel.values) {
    if (c.wireValue == value) return c;
  }
  return null;
}

/// How a member arrived, as the join endpoint's `source` parameter. Mirrors the backend's
/// `join_source` enum.
enum JoinSource {
  link,
  code,
  card,
  telegram;

  String get wireValue => name.toUpperCase();

  /// The source implied by a link's channel marker: a link shared from a card attributes the
  /// join to the card, which is what makes the card's contribution visible.
  static JoinSource fromShareChannel(ShareChannel? channel) {
    switch (channel) {
      case ShareChannel.card:
        return JoinSource.card;
      case ShareChannel.telegram:
        return JoinSource.telegram;
      case ShareChannel.link:
      case null:
        return JoinSource.link;
    }
  }
}

/// Reads the referrer out of an incoming link, so the member who shared it can be paid.
String? referrerFromLink(String url) {
  final match = RegExp(r'[?&]ref=([^&#]+)').firstMatch(url);
  return match?.group(1);
}
