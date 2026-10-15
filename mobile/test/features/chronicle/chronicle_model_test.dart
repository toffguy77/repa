import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/chronicle/domain/chronicle.dart';

Map<String, dynamic> entryJson({bool smallSample = false, int voters = 10}) => {
      'question_id': 'q1',
      'question_text': 'Кто чаще всех смеётся?',
      'category': 'FUNNY',
      'user_id': 'u1',
      'username': 'alice',
      'avatar_emoji': '\u{1F60E}',
      'percentage': 66.7,
      'vote_count': 2,
      'total_voters': voters,
      'small_sample': smallSample,
    };

void main() {
  group('ChronicleEntry', () {
    test('round-trips', () {
      final entry = ChronicleEntry.fromJson(entryJson());
      expect(entry.username, 'alice');
      expect(entry.percentage, 66.7);
      expect(entry.smallSample, isFalse);
      expect(ChronicleEntry.fromJson(entry.toJson()), entry);
    });

    test('reads the small-sample mark', () {
      final entry = ChronicleEntry.fromJson(entryJson(smallSample: true, voters: 3));
      expect(entry.smallSample, isTrue);
      expect(entry.totalVoters, 3);
    });

    test('a response without the mark reads as not marked', () {
      // The server decides; an older server that does not send the field must not crash the screen.
      final json = entryJson()..remove('small_sample');
      expect(ChronicleEntry.fromJson(json).smallSample, isFalse);
    });
  });

  group('Chronicle', () {
    test('an empty chronicle is empty, not null', () {
      final chronicle = Chronicle.fromJson({
        'seasons': <dynamic>[],
        'standing': {'available': false, 'reason': 'рано', 'members': <dynamic>[]},
        'seasons_total': 0,
        'seasons_shown': 0,
      });
      expect(chronicle.seasons, isEmpty);
      expect(chronicle.isEmpty, isTrue);
      expect(chronicle.standing.available, isFalse);
      expect(chronicle.standing.members, isEmpty);
    });

    test('a missing standing does not crash', () {
      final chronicle = Chronicle.fromJson({'seasons': <dynamic>[]});
      expect(chronicle.standing.available, isFalse);
      expect(chronicle.seasonsTotal, 0);
    });

    test('isTruncated reports the cap honestly', () {
      final shown = Chronicle.fromJson({
        'seasons': <dynamic>[],
        'seasons_total': 50,
        'seasons_shown': 20,
      });
      expect(shown.isTruncated, isTrue);

      final all = Chronicle.fromJson({
        'seasons': <dynamic>[],
        'seasons_total': 3,
        'seasons_shown': 3,
      });
      expect(all.isTruncated, isFalse);
    });

    test('seasons round-trip with their entries', () {
      final chronicle = Chronicle.fromJson({
        'seasons': [
          {
            'season_id': 's1',
            'number': 2,
            'reveal_at': '2026-09-25T17:00:00Z',
            'entries': [entryJson()],
          }
        ],
        'seasons_total': 1,
        'seasons_shown': 1,
      });
      expect(chronicle.seasons.single.number, 2);
      expect(chronicle.seasons.single.entries.single.username, 'alice');
    });
  });

  group('StandingEntry', () {
    test('a ranked member reads their place and its basis', () {
      final entry = StandingEntry.fromJson({
        'user_id': 'u1',
        'username': 'alice',
        'rank': 1,
        'ranked': true,
        'seasons_played': 4,
      });
      expect(entry.isRanked, isTrue);
      expect(entry.rank, 1);
      expect(entry.seasonsPlayed, 4);
    });

    test('an unranked member has no place rather than the last one', () {
      // "Not yet ranked" and "always wrong" are different facts, and the second one would be a lie
      // about someone who has not played yet.
      final entry = StandingEntry.fromJson({
        'user_id': 'u2',
        'username': 'новенький',
        'rank': 0,
        'ranked': false,
      });
      expect(entry.isRanked, isFalse);
      expect(entry.rank, 0);
    });

    test('a server that still sends an accuracy figure is ignored', () {
      // The figure is withheld deliberately: next to the chronicle's published winners it would
      // recover a member's individual votes. An extra field must be dropped, not surfaced.
      final entry = StandingEntry.fromJson({
        'user_id': 'u1',
        'username': 'alice',
        'rank': 1,
        'ranked': true,
        'seasons_played': 4,
        'accuracy': 0.95,
      });
      expect(entry.toJson().containsKey('accuracy'), isFalse);
    });
  });
}
