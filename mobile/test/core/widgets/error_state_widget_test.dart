import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/widgets/error_state_widget.dart';

void main() {
  Widget buildWidget({String? message, VoidCallback? onRetry}) {
    return MaterialApp(
      home: Scaffold(
        body: ErrorStateWidget(message: message, onRetry: onRetry),
      ),
    );
  }

  testWidgets('shows error icon', (tester) async {
    await tester.pumpWidget(buildWidget());

    expect(find.byIcon(Icons.error_outline_rounded), findsOneWidget);
  });

  testWidgets('shows friendly message for connection errors', (tester) async {
    await tester.pumpWidget(buildWidget(message: 'connection timeout'));

    expect(find.text('Нет соединения, проверь интернет'), findsOneWidget);
  });

  testWidgets('shows friendly message for server errors', (tester) async {
    await tester.pumpWidget(buildWidget(message: '500 Internal Server Error'));

    expect(find.text('Что-то пошло не так, попробуй позже'), findsOneWidget);
  });

  testWidgets('shows raw message when not a known pattern', (tester) async {
    await tester.pumpWidget(buildWidget(message: 'Group not found'));

    expect(find.text('Group not found'), findsOneWidget);
  });

  testWidgets('shows default message when null', (tester) async {
    await tester.pumpWidget(buildWidget());

    expect(find.text('Что-то пошло не так'), findsOneWidget);
  });

  testWidgets('shows retry button and calls callback', (tester) async {
    var retried = false;
    await tester.pumpWidget(buildWidget(onRetry: () => retried = true));

    expect(find.text('Повторить'), findsOneWidget);
    await tester.tap(find.text('Повторить'));
    expect(retried, true);
  });

  testWidgets('hides retry button when onRetry is null', (tester) async {
    await tester.pumpWidget(buildWidget());

    expect(find.byType(ElevatedButton), findsNothing);
  });

  group('friendlyMessage', () {
    test('returns friendly text for connection errors', () {
      expect(
        ErrorStateWidget.friendlyMessage('connection refused'),
        'Нет соединения, проверь интернет',
      );
    });

    test('returns friendly text for timeout', () {
      expect(
        ErrorStateWidget.friendlyMessage('timeout error'),
        'Нет соединения, проверь интернет',
      );
    });

    test('returns friendly text for socket errors', () {
      expect(
        ErrorStateWidget.friendlyMessage('SocketException'),
        'Нет соединения, проверь интернет',
      );
    });

    test('returns friendly text for server errors', () {
      expect(
        ErrorStateWidget.friendlyMessage('Internal server error 500'),
        'Что-то пошло не так, попробуй позже',
      );
    });

    test('returns raw message for unknown errors', () {
      expect(
        ErrorStateWidget.friendlyMessage('User already exists'),
        'User already exists',
      );
    });

    test('returns default for null', () {
      expect(
        ErrorStateWidget.friendlyMessage(null),
        'Что-то пошло не так',
      );
    });

    test('returns default for empty string', () {
      expect(
        ErrorStateWidget.friendlyMessage(''),
        'Что-то пошло не так',
      );
    });
  });
}
