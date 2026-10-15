import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/chronicle/domain/chronicle.dart';
import 'package:repa/features/chronicle/presentation/widgets/chronicle_season_card.dart';
import 'package:repa/features/chronicle/presentation/widgets/guess_standing_card.dart';
import 'package:repa/features/groups/domain/group.dart';
import 'package:repa/features/groups/presentation/widgets/growth_notice.dart';

Future<void> pump(WidgetTester tester, Widget child) async {
  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.dark,
    home: Scaffold(body: SingleChildScrollView(child: child)),
  ));
  await tester.pump(const Duration(milliseconds: 400));
}

ChronicleEntry entry({
  String username = 'alice',
  double percentage = 66.7,
  bool smallSample = false,
  int totalVoters = 10,
}) =>
    ChronicleEntry(
      questionId: 'q1',
      questionText: 'Кто чаще всех смеётся?',
      category: 'FUNNY',
      userId: 'u1',
      username: username,
      percentage: percentage,
      voteCount: 2,
      totalVoters: totalVoters,
      smallSample: smallSample,
    );

void main() {
  group('ChronicleSeasonCard', () {
    testWidgets('shows the season, the question and who led it', (tester) async {
      await pump(
        tester,
        ChronicleSeasonCard(
          season: ChronicleSeason(
            seasonId: 's1',
            number: 3,
            revealAt: '2026-09-25T17:00:00Z',
            entries: [entry()],
          ),
        ),
      );

      expect(find.text('Репа №3'), findsOneWidget);
      expect(find.text('Кто чаще всех смеётся?'), findsOneWidget);
      expect(find.text('alice'), findsOneWidget);
      expect(find.text('67%'), findsOneWidget);
    });

    testWidgets('marks a result computed over too few voters', (tester) async {
      await pump(
        tester,
        ChronicleSeasonCard(
          season: ChronicleSeason(
            seasonId: 's1',
            number: 1,
            revealAt: '2026-09-18T17:00:00Z',
            entries: [entry(smallSample: true, totalVoters: 3)],
          ),
        ),
      );
      expect(find.textContaining('Голосовало 3'), findsOneWidget);
      expect(find.textContaining('выдают, кто как голосовал'), findsOneWidget);
    });

    testWidgets('does not mark a result with enough voters', (tester) async {
      await pump(
        tester,
        ChronicleSeasonCard(
          season: ChronicleSeason(
            seasonId: 's1',
            number: 1,
            revealAt: '2026-09-18T17:00:00Z',
            entries: [entry()],
          ),
        ),
      );
      expect(find.textContaining('Голосовало'), findsNothing);
    });

    testWidgets('a season whose entries are all blocked still appears',
        (tester) async {
      // Dropping it would tell the reader that a block is in play.
      await pump(
        tester,
        const ChronicleSeasonCard(
          season: ChronicleSeason(
            seasonId: 's1',
            number: 2,
            revealAt: '2026-09-25T17:00:00Z',
          ),
        ),
      );
      expect(find.text('Репа №2'), findsOneWidget);
      expect(find.text('Нечего показать'), findsOneWidget);
    });

    testWidgets('an unparseable timestamp does not lose the season',
        (tester) async {
      await pump(
        tester,
        ChronicleSeasonCard(
          season: ChronicleSeason(
            seasonId: 's1',
            number: 1,
            revealAt: 'not a date',
            entries: [entry()],
          ),
        ),
      );
      expect(find.text('Репа №1'), findsOneWidget);
    });
  });

  group('GuessStandingCard', () {
    testWidgets('ranks members with the seasons each is based on',
        (tester) async {
      await pump(
        tester,
        const GuessStandingCard(
          standing: GuessStanding(
            available: true,
            members: [
              StandingEntry(
                  userId: 'u1',
                  username: 'alice',
                  rank: 1,
                  ranked: true,
                  seasonsPlayed: 4),
              StandingEntry(
                  userId: 'u2',
                  username: 'bob',
                  rank: 2,
                  ranked: true,
                  seasonsPlayed: 4),
            ],
          ),
        ),
      );

      expect(find.text('Знатоки'), findsOneWidget);
      // Places, and the history each rests on — the accuracy figure is withheld by the server, since
      // next to the published winners it would recover a member's individual votes.
      expect(find.text('1'), findsOneWidget);
      expect(find.text('2'), findsOneWidget);
      expect(find.text('за 4 репы'), findsNWidgets(2));
      expect(find.textContaining('%'), findsNothing);
    });

    testWidgets('shows the reason while it is withheld', (tester) async {
      await pump(
        tester,
        const GuessStandingCard(
          standing: GuessStanding(
            available: false,
            reason: 'Нужно ещё хотя бы одну репу, чтобы сравнивать',
          ),
        ),
      );
      expect(find.textContaining('хотя бы одну репу'), findsOneWidget);
      // No ranking shown, so nobody is placed on a single season's luck.
      expect(find.textContaining('за '), findsNothing);
    });

    testWidgets('falls back to its own copy when the server sends no reason',
        (tester) async {
      await pump(tester,
          const GuessStandingCard(standing: GuessStanding(available: false)));
      expect(find.text('Пока рано сравнивать'), findsOneWidget);
    });

    testWidgets('shows an unranked member as not yet played', (tester) async {
      await pump(
        tester,
        const GuessStandingCard(
          standing: GuessStanding(
            available: true,
            members: [
              StandingEntry(
                  userId: 'u1',
                  username: 'alice',
                  rank: 1,
                  ranked: true,
                  seasonsPlayed: 2),
              StandingEntry(userId: 'u2', username: 'новенький'),
            ],
          ),
        ),
      );
      expect(find.text('ещё не играл'), findsOneWidget);
      // A dash rather than a position: placed nowhere, not placed last.
      expect(find.text('—'), findsOneWidget);
      expect(find.textContaining('за 0'), findsNothing);
    });
  });

  group('GrowthNotice', () {
    testWidgets('a group that cannot reveal is told so', (tester) async {
      await pump(
        tester,
        const GrowthNotice(
          threshold: GrowthThreshold(
              size: 3, needed: 1, unlocks: GrowthUnlock.reveal),
        ),
      );
      expect(find.textContaining('репа не откроется'), findsOneWidget);
      expect(find.textContaining('1 человека'), findsOneWidget);
    });

    testWidgets('a group below the detector is told what it unlocks',
        (tester) async {
      await pump(
        tester,
        const GrowthNotice(
          threshold: GrowthThreshold(
              size: 5, needed: 2, unlocks: GrowthUnlock.detector),
        ),
      );
      expect(find.textContaining('откроется детектор'), findsOneWidget);
      expect(find.textContaining('2 человека'), findsOneWidget);
    });

    testWidgets('a group past every threshold is given no target',
        (tester) async {
      // The absence of the line is the message: the app must not invent a goal above the last real
      // threshold.
      await pump(tester, const GrowthNotice(threshold: null));
      expect(find.textContaining('Позови'), findsNothing);
    });

    testWidgets('an unknown unlock states the number without inventing a reason',
        (tester) async {
      await pump(
        tester,
        const GrowthNotice(
          threshold: GrowthThreshold(
              size: 9, needed: 4, unlocks: GrowthUnlock.unknown),
        ),
      );
      expect(find.textContaining('4 человека'), findsOneWidget);
      expect(find.textContaining('откроется детектор'), findsNothing);
    });
  });
}
