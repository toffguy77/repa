import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/groups/domain/group.dart';
import 'package:repa/features/groups/presentation/widgets/pending_reveal_notice.dart';
import 'package:repa/features/groups/presentation/widgets/small_group_notice.dart';
import 'package:repa/core/theme/app_theme.dart';

ActiveSeason seasonWith(String state, {int members = 0, int voters = 0}) =>
    ActiveSeason.fromJson({
      'id': 's1',
      'status': state == 'REVEALED' ? 'REVEALED' : 'VOTING',
      'reveal_at': '2026-10-09T17:00:00Z',
      'voted_count': 2,
      'total_count': 4,
      'user_voted': true,
      'reveal_state': state,
      'members_needed': members,
      'voters_needed': voters,
    });

Future<void> pumpNotice(WidgetTester tester, Widget child) =>
    tester.pumpWidget(MaterialApp(theme: AppTheme.dark, home: Scaffold(body: child)));

void main() {
  group('PendingRevealNotice', () {
    testWidgets('shows nothing while the Reveal is on schedule', (tester) async {
      await pumpNotice(
          tester, PendingRevealNotice(season: seasonWith('SCHEDULED')));

      expect(find.byType(Icon), findsNothing);
      expect(find.textContaining('Нужно'), findsNothing);
    });

    testWidgets('shows nothing once the season is revealed', (tester) async {
      await pumpNotice(
          tester, PendingRevealNotice(season: seasonWith('REVEALED')));

      expect(find.byType(Icon), findsNothing);
    });

    testWidgets('waiting for members names how many to invite', (tester) async {
      await pumpNotice(tester,
          PendingRevealNotice(season: seasonWith('WAITING_FOR_MEMBERS', members: 1)));

      expect(find.text('Нужно больше людей'), findsOneWidget);
      expect(find.textContaining('1 человека'), findsOneWidget);
      expect(find.byIcon(Icons.group_add_outlined), findsOneWidget);
    });

    testWidgets('waiting for voters names how many must vote', (tester) async {
      await pumpNotice(tester,
          PendingRevealNotice(season: seasonWith('WAITING_FOR_VOTERS', voters: 2)));

      expect(find.text('Нужно больше голосов'), findsOneWidget);
      expect(find.textContaining('2 человека'), findsOneWidget);
    });

    testWidgets('postponed says votes are kept', (tester) async {
      await pumpNotice(
          tester, PendingRevealNotice(season: seasonWith('POSTPONED')));

      expect(find.text('Репа перенесена'), findsOneWidget);
      expect(find.textContaining('сохранены'), findsOneWidget);
      expect(find.byIcon(Icons.schedule), findsOneWidget);
    });
  });

  group('SmallGroupNotice', () {
    testWidgets('appears at 4 members', (tester) async {
      await pumpNotice(tester, const SmallGroupNotice(memberCount: 4));

      expect(find.text('Вас пока мало'), findsOneWidget);
      expect(find.textContaining('меньше 5 человек'), findsOneWidget);
    });

    testWidgets('appears at 2 members', (tester) async {
      await pumpNotice(tester, const SmallGroupNotice(memberCount: 2));

      expect(find.text('Вас пока мало'), findsOneWidget);
    });

    testWidgets('is absent at 5 members', (tester) async {
      await pumpNotice(tester, const SmallGroupNotice(memberCount: 5));

      expect(find.text('Вас пока мало'), findsNothing);
    });

    testWidgets('is absent in a large group', (tester) async {
      await pumpNotice(tester, const SmallGroupNotice(memberCount: 20));

      expect(find.text('Вас пока мало'), findsNothing);
    });
  });
}
