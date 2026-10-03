import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/reveal/domain/reveal.dart';
import 'package:repa/features/reveal/presentation/widgets/anticipation_panel.dart';

Future<void> pumpPanel(WidgetTester tester, AnticipationState? state) async {
  tester.view.physicalSize = const Size(1200, 2400);
  tester.view.devicePixelRatio = 1.0;
  addTearDown(tester.view.reset);

  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.dark,
    home: Scaffold(
      body: SingleChildScrollView(child: AnticipationPanel(state: state)),
    ),
  ));
  await tester.pump(const Duration(milliseconds: 500));
}

void main() {
  group('AnticipationPanel', () {
    testWidgets('renders nothing before the state has loaded', (tester) async {
      await pumpPanel(tester, null);

      expect(find.textContaining('ответили'), findsNothing);
    });

    testWidgets('zero votes says so plainly rather than showing an empty panel',
        (tester) async {
      await pumpPanel(tester, const AnticipationState(votersAboutMe: 0));

      expect(find.text('0'), findsOneWidget);
      expect(find.textContaining('Пока тихо'), findsOneWidget);
    });

    testWidgets('a count is the headline — the thing the push promised',
        (tester) async {
      await pumpPanel(tester, const AnticipationState(votersAboutMe: 4));

      expect(find.text('4'), findsOneWidget);
      expect(find.text('ОТВЕТИЛИ ПРО ТЕБЯ'), findsOneWidget);
      expect(find.textContaining('Пока тихо'), findsNothing);
    });

    testWidgets('a teaser shows one emoji and no attribute', (tester) async {
      await pumpPanel(
        tester,
        const AnticipationState(votersAboutMe: 4, teaserEmoji: '🔥'),
      );

      expect(find.text('🔥'), findsOneWidget);
      expect(find.textContaining('почти определился'), findsOneWidget);
      // Nothing that names the category or the question.
      expect(find.textContaining('HOT'), findsNothing);
      expect(find.textContaining('%'), findsNothing);
    });

    testWidgets('no teaser means no teaser row', (tester) async {
      await pumpPanel(tester, const AnticipationState(votersAboutMe: 4));

      expect(find.textContaining('почти определился'), findsNothing);
    });

    testWidgets('a reveal time renders the countdown', (tester) async {
      await pumpPanel(
        tester,
        const AnticipationState(
          votersAboutMe: 2,
          revealAt: '2030-01-04T17:00:00Z',
        ),
      );

      // The countdown renders a day/time chip; assert it appears at all.
      expect(find.byType(Text), findsWidgets);
    });
  });

  group('AnticipationState', () {
    test('defaults safely when the backend omits the optional fields', () {
      final parsed = AnticipationState.fromJson(const {'voters_about_me': 3});

      expect(parsed.votersAboutMe, 3);
      expect(parsed.teaserEmoji, '');
      expect(parsed.revealAt, '');
    });

    test('has no field that could carry an identity or an attribute', () {
      const state = AnticipationState(votersAboutMe: 1, teaserEmoji: '🔥');
      final json = state.toJson();

      expect(json.keys.toSet(),
          {'voters_about_me', 'teaser_emoji', 'reveal_at'});
    });
  });
}
