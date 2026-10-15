import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_service.dart';
import 'package:repa/core/providers/api_provider.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/profile/domain/profile.dart';
import 'package:repa/features/profile/presentation/member_profile_screen.dart';
import 'package:repa/features/profile/presentation/widgets/stat_card.dart';

class MockDio extends Mock implements Dio {}

Map<String, dynamic> statsJson({bool withAccuracy = false}) => {
      'seasons_played': 4,
      'voting_streak': 3,
      'max_voting_streak': 4,
      'total_votes_cast': 40,
      'total_votes_received': 22,
      if (withAccuracy) 'guess_accuracy': 87.5,
    };

void main() {
  group('UserStats.guessAccuracy', () {
    test('is null when the server withholds it', () {
      // Another member's profile: the figure is absent, and absent must not become 0.
      final stats = UserStats.fromJson(statsJson());
      expect(stats.guessAccuracy, isNull);
      // Everything that counts participation rather than matches is still there.
      expect(stats.seasonsPlayed, 4);
      expect(stats.votingStreak, 3);
      expect(stats.maxVotingStreak, 4);
      expect(stats.totalVotesCast, 40);
      expect(stats.totalVotesReceived, 22);
    });

    test('is read when the server sends it', () {
      final stats = UserStats.fromJson(statsJson(withAccuracy: true));
      expect(stats.guessAccuracy, 87.5);
    });

    test('round-trips both shapes', () {
      final withheld = UserStats.fromJson(statsJson());
      expect(UserStats.fromJson(withheld.toJson()).guessAccuracy, isNull);

      final own = UserStats.fromJson(statsJson(withAccuracy: true));
      expect(UserStats.fromJson(own.toJson()).guessAccuracy, 87.5);
    });
  });

  group('StatCard', _statCardRegression);

  group('MemberProfileScreen', () {
    // The real screen with a mocked API, so the assertion is about the widget tree the condition in
    // _buildStats actually produces — not about a Column rebuilt in the test.
    Future<void> pumpProfile(WidgetTester tester, {required bool withAccuracy}) async {
      final mockDio = MockDio();
      when(() => mockDio.get('/groups/g1/members/u1/profile'))
          .thenAnswer((_) async => Response(
                data: {
                  'data': {
                    'user': {'username': 'alice'},
                    'stats': statsJson(withAccuracy: withAccuracy),
                    'achievements': <dynamic>[],
                    'legend': 'alice — загадочная личность',
                    'season_history': <dynamic>[],
                  }
                },
                statusCode: 200,
                requestOptions:
                    RequestOptions(path: '/groups/g1/members/u1/profile'),
              ));

      await tester.pumpWidget(ProviderScope(
        overrides: [apiServiceProvider.overrideWithValue(ApiService(mockDio))],
        child: MaterialApp(
          theme: AppTheme.dark,
          home: const MemberProfileScreen(groupId: 'g1', userId: 'u1'),
        ),
      ));
      await tester.pumpAndSettle();
    }

    testWidgets('hides the accuracy tile when the figure is withheld',
        (tester) async {
      await pumpProfile(tester, withAccuracy: false);

      expect(find.text('Точность угадывания'), findsNothing);
      // Specifically not a fallback: "0%" would read as "never guessed right", which is a different
      // and false claim about another member.
      expect(find.text('0%'), findsNothing);
      expect(find.text('0.0%'), findsNothing);
      // The rest of the statistics are still rendered.
      expect(find.text('Сезонов сыграно'), findsOneWidget);
      expect(find.text('Стрик голосований'), findsOneWidget);
    });

    testWidgets("shows the accuracy tile on one's own profile", (tester) async {
      await pumpProfile(tester, withAccuracy: true);

      expect(find.text('Точность угадывания'), findsOneWidget);
      expect(find.text('87.5%'), findsOneWidget);
    });
  });
}

// StatCard previously read MediaQuery from initState (via context.motion), which Flutter forbids: it
// threw in any test that mounted the card, and in release it meant the controller's duration never
// followed a change in the reduce-motion preference — the one thing the motion token exists for.
// Guarded here because the profile screen is where the card is used.
void _statCardRegression() {
  testWidgets('StatCard mounts without reading MediaQuery too early',
      (tester) async {
    await tester.pumpWidget(MaterialApp(
      theme: AppTheme.dark,
      home: const Scaffold(
        body: StatCard(label: 'Сезонов сыграно', value: '4', icon: Icons.calendar_today),
      ),
    ));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('Сезонов сыграно'), findsOneWidget);
  });

  testWidgets('StatCard honours the reduce-motion preference', (tester) async {
    await tester.pumpWidget(MaterialApp(
      theme: AppTheme.dark,
      home: const MediaQuery(
        data: MediaQueryData(disableAnimations: true),
        child: Scaffold(
          body: StatCard(
              label: 'Сезонов сыграно',
              value: '4',
              icon: Icons.calendar_today,
              animateNumber: true),
        ),
      ),
    ));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('4'), findsOneWidget);
  });
}
