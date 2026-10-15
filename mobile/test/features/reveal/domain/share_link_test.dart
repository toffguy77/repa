import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/reveal/domain/share_link.dart';

void main() {
  group('buildShareLink', () {
    test('carries the group code and the channel marker', () {
      final link = buildShareLink(inviteCode: 'AB2CD3', channel: ShareChannel.card);

      expect(link, 'https://repa.app/join/AB2CD3?s=card');
      expect(link, contains('AB2CD3'));
    });

    test('different channels produce distinguishable links', () {
      final card = buildShareLink(inviteCode: 'AB2CD3', channel: ShareChannel.card);
      final tg = buildShareLink(inviteCode: 'AB2CD3', channel: ShareChannel.telegram);

      expect(card, isNot(tg));
    });

    test('the code stays the path segment so the router still matches', () {
      final link = buildShareLink(inviteCode: 'AB2CD3', channel: ShareChannel.link);

      expect(Uri.parse(link).pathSegments, ['join', 'AB2CD3']);
    });
  });

  group('buildShareText', () {
    test('includes the code-bearing link', () {
      final text = buildShareText(inviteCode: 'AB2CD3', channel: ShareChannel.card);

      expect(text, contains('https://repa.app/join/AB2CD3?s=card'));
      expect(text, isNot('Моя репа repa.app'),
          reason: 'the pre-change text carried no way back into the product');
    });
  });

  group('shareChannelFromLink', () {
    test('reads the marker a shared link carries', () {
      expect(shareChannelFromLink('https://repa.app/join/AB2CD3?s=card'),
          ShareChannel.card);
      expect(shareChannelFromLink('https://repa.app/join/AB2CD3?s=telegram'),
          ShareChannel.telegram);
      expect(shareChannelFromLink('https://repa.app/join/AB2CD3?s=link'),
          ShareChannel.link);
    });

    test('reads a marker that is not the first parameter', () {
      expect(shareChannelFromLink('https://repa.app/join/AB2CD3?x=1&s=card'),
          ShareChannel.card);
    });

    test('returns null for a plain link', () {
      expect(shareChannelFromLink('https://repa.app/join/AB2CD3'), isNull);
    });

    test('returns null for an unknown marker rather than guessing', () {
      expect(shareChannelFromLink('https://repa.app/join/AB2CD3?s=instagram'), isNull);
    });
  });

  group('JoinSource.fromShareChannel', () {
    test('a card share attributes the join to the card', () {
      expect(JoinSource.fromShareChannel(ShareChannel.card), JoinSource.card);
    });

    test('a telegram share attributes to telegram', () {
      expect(JoinSource.fromShareChannel(ShareChannel.telegram), JoinSource.telegram);
    });

    test('a plain link, or no marker at all, attributes to the link', () {
      expect(JoinSource.fromShareChannel(ShareChannel.link), JoinSource.link);
      expect(JoinSource.fromShareChannel(null), JoinSource.link);
    });
  });

  referrerTests();

  group('wire values match the backend enums', () {
    test('join sources are upper-case', () {
      expect(JoinSource.link.wireValue, 'LINK');
      expect(JoinSource.code.wireValue, 'CODE');
      expect(JoinSource.card.wireValue, 'CARD');
      expect(JoinSource.telegram.wireValue, 'TELEGRAM');
    });

    test('share channels are lower-case', () {
      expect(ShareChannel.card.wireValue, 'card');
      expect(ShareChannel.telegram.wireValue, 'telegram');
      expect(ShareChannel.link.wireValue, 'link');
    });
  });
}

void referrerTests() {
  group('referrer in the share link', () {
    test('a card link carries the sharer so the reward can reach them', () {
      final link = buildShareLink(
        inviteCode: 'AB2CD3',
        channel: ShareChannel.card,
        referrerId: 'user-7',
      );

      expect(link, 'https://repa.app/join/AB2CD3?s=card&ref=user-7');
      expect(referrerFromLink(link), 'user-7');
    });

    test('the share text carries it too', () {
      final text = buildShareText(
        inviteCode: 'AB2CD3',
        channel: ShareChannel.card,
        referrerId: 'user-7',
      );

      expect(text, contains('ref=user-7'));
    });

    test('a link without a referrer still parses and still joins', () {
      final link = buildShareLink(inviteCode: 'AB2CD3', channel: ShareChannel.link);

      expect(link, isNot(contains('ref=')));
      expect(referrerFromLink(link), isNull);
      expect(Uri.parse(link).pathSegments, ['join', 'AB2CD3']);
    });

    test('an empty referrer is omitted rather than sent blank', () {
      final link = buildShareLink(
        inviteCode: 'AB2CD3',
        channel: ShareChannel.card,
        referrerId: '',
      );

      expect(link, isNot(contains('ref=')));
    });

    test('the referrer is read regardless of parameter order', () {
      expect(referrerFromLink('https://repa.app/join/AB2CD3?ref=user-7&s=card'), 'user-7');
      expect(referrerFromLink('https://repa.app/join/AB2CD3?s=card&ref=user-7'), 'user-7');
    });

    test('a fragment does not leak into the referrer', () {
      expect(referrerFromLink('https://repa.app/join/AB2CD3?ref=user-7#x'), 'user-7');
    });

    test('channel and referrer are read independently', () {
      const link = 'https://repa.app/join/AB2CD3?s=telegram&ref=user-9';

      expect(shareChannelFromLink(link), ShareChannel.telegram);
      expect(referrerFromLink(link), 'user-9');
    });
  });
}
