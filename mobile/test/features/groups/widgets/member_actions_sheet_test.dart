import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/theme/app_theme.dart';
import 'package:repa/features/groups/domain/group.dart';
import 'package:repa/features/groups/presentation/widgets/member_actions_sheet.dart';

const _member = Member(
  id: 'u2',
  username: 'Маша',
  avatarEmoji: '🍆',
  isAdmin: false,
);

Future<void> pumpSheet(
  WidgetTester tester, {
  bool isAdminViewer = false,
  VoidCallback? onBlock,
  VoidCallback? onReport,
  VoidCallback? onRemove,
}) async {
  tester.view.physicalSize = const Size(1200, 2400);
  tester.view.devicePixelRatio = 1.0;
  addTearDown(tester.view.reset);

  await tester.pumpWidget(MaterialApp(
    theme: AppTheme.dark,
    home: Scaffold(
      body: SingleChildScrollView(
        child: MemberActionsSheet(
          member: _member,
          isAdminViewer: isAdminViewer,
          onBlock: onBlock ?? () {},
          onReport: onReport ?? () {},
          onRemove: onRemove,
        ),
      ),
    ),
  ));
  await tester.pump(const Duration(milliseconds: 500));
}

void main() {
  group('MemberActionsSheet', () {
    testWidgets('offers block and report to any member', (tester) async {
      await pumpSheet(tester);

      expect(find.text('Маша'), findsOneWidget);
      expect(find.text('Заблокировать'), findsOneWidget);
      expect(find.text('Пожаловаться'), findsOneWidget);
    });

    testWidgets('hides removal from a non-admin', (tester) async {
      await pumpSheet(tester, onRemove: () {});

      expect(find.text('Удалить из группы'), findsNothing);
    });

    testWidgets('offers removal to the admin', (tester) async {
      await pumpSheet(tester, isAdminViewer: true, onRemove: () {});

      expect(find.text('Удалить из группы'), findsOneWidget);
      expect(find.textContaining('вернуться не получится'), findsOneWidget);
    });

    testWidgets('block states its consequence and is reversible', (tester) async {
      var blocked = 0;
      await pumpSheet(tester, onBlock: () => blocked++);

      await tester.tap(find.text('Заблокировать'));
      await tester.pumpAndSettle();

      expect(find.textContaining('не сможете голосовать друг за друга'), findsOneWidget);
      expect(find.textContaining('Это можно отменить'), findsOneWidget);
      expect(blocked, 0, reason: 'nothing happens until the confirmation is accepted');

      await tester.tap(find.widgetWithText(ElevatedButton, 'Заблокировать').last);
      await tester.pumpAndSettle();

      expect(blocked, 1);
    });

    testWidgets('report says the member will not be told', (tester) async {
      var reported = 0;
      await pumpSheet(tester, onReport: () => reported++);

      await tester.tap(find.text('Пожаловаться'));
      await tester.pumpAndSettle();

      expect(find.textContaining('не получит уведомления'), findsOneWidget);

      await tester.tap(find.widgetWithText(ElevatedButton, 'Пожаловаться').last);
      await tester.pumpAndSettle();

      expect(reported, 1);
    });

    testWidgets('removal says it cannot be undone, even with a new link', (tester) async {
      var removed = 0;
      await pumpSheet(tester, isAdminViewer: true, onRemove: () => removed++);

      await tester.tap(find.text('Удалить из группы'));
      await tester.pumpAndSettle();

      expect(find.textContaining('даже если создать новую'), findsOneWidget);
      expect(find.textContaining('Отменить это нельзя'), findsOneWidget);

      await tester.tap(find.widgetWithText(ElevatedButton, 'Удалить').last);
      await tester.pumpAndSettle();

      expect(removed, 1);
    });

    testWidgets('cancelling a confirmation does nothing', (tester) async {
      var blocked = 0;
      await pumpSheet(tester, onBlock: () => blocked++);

      await tester.tap(find.text('Заблокировать'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Отмена'));
      await tester.pumpAndSettle();

      expect(blocked, 0);
    });
  });
}
