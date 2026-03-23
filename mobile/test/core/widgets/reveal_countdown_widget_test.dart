import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/widgets/reveal_countdown_widget.dart';

void main() {
  Widget buildWidget({required String revealAt}) {
    return MaterialApp(
      home: Scaffold(
        body: RevealCountdownWidget(revealAt: revealAt),
      ),
    );
  }

  testWidgets('shows day and time from revealAt', (tester) async {
    // Use a future Friday at 17:00 UTC (20:00 MSK)
    final friday = _nextFriday();
    await tester.pumpWidget(buildWidget(revealAt: friday.toUtc().toIso8601String()));
    await tester.pump();

    // Should contain the middle dot separator
    expect(find.textContaining('\u00b7'), findsOneWidget);
  });

  testWidgets('shows remaining time format with days', (tester) async {
    final future = DateTime.now().add(const Duration(days: 3, hours: 5));
    await tester.pumpWidget(buildWidget(revealAt: future.toUtc().toIso8601String()));
    await tester.pump();

    // Should show something like "3д 5ч"
    expect(find.textContaining('д'), findsOneWidget);
  });

  testWidgets('shows remaining time format with hours', (tester) async {
    final future = DateTime.now().add(const Duration(hours: 5, minutes: 30));
    await tester.pumpWidget(buildWidget(revealAt: future.toUtc().toIso8601String()));
    await tester.pump();

    expect(find.textContaining('ч'), findsOneWidget);
  });

  testWidgets('shows remaining time format with minutes only', (tester) async {
    final future = DateTime.now().add(const Duration(minutes: 30));
    await tester.pumpWidget(buildWidget(revealAt: future.toUtc().toIso8601String()));
    await tester.pump();

    expect(find.textContaining('мин'), findsOneWidget);
  });

  testWidgets('shows "Скоро!" when time has passed', (tester) async {
    final past = DateTime.now().subtract(const Duration(hours: 1));
    await tester.pumpWidget(buildWidget(revealAt: past.toUtc().toIso8601String()));
    await tester.pump();

    expect(find.textContaining('Скоро!'), findsOneWidget);
  });

  testWidgets('updates when revealAt changes via didUpdateWidget', (tester) async {
    final future1 = DateTime.now().add(const Duration(days: 2));
    final future2 = DateTime.now().add(const Duration(minutes: 15));

    await tester.pumpWidget(buildWidget(revealAt: future1.toUtc().toIso8601String()));
    await tester.pump();
    expect(find.textContaining('д'), findsOneWidget);

    await tester.pumpWidget(buildWidget(revealAt: future2.toUtc().toIso8601String()));
    await tester.pump();
    expect(find.textContaining('мин'), findsOneWidget);
  });

  testWidgets('disposes timer without errors', (tester) async {
    final future = DateTime.now().add(const Duration(days: 1));
    await tester.pumpWidget(buildWidget(revealAt: future.toUtc().toIso8601String()));
    await tester.pump();

    // Remove widget — should dispose timer cleanly
    await tester.pumpWidget(const MaterialApp(home: Scaffold()));
    await tester.pump();
    // No exception = pass
  });
}

DateTime _nextFriday() {
  var d = DateTime.now();
  while (d.weekday != DateTime.friday) {
    d = d.add(const Duration(days: 1));
  }
  return DateTime(d.year, d.month, d.day, 20, 0);
}
