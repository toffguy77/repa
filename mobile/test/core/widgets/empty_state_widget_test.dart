import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/widgets/empty_state_widget.dart';

void main() {
  Widget buildWidget({
    String emoji = '🍑',
    String title = 'No data',
    String subtitle = 'Try again',
    String? buttonText,
    VoidCallback? onButtonPressed,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: EmptyStateWidget(
          emoji: emoji,
          title: title,
          subtitle: subtitle,
          buttonText: buttonText,
          onButtonPressed: onButtonPressed,
        ),
      ),
    );
  }

  testWidgets('renders emoji, title, and subtitle', (tester) async {
    await tester.pumpWidget(buildWidget());
    await tester.pumpAndSettle();

    expect(find.text('🍑'), findsOneWidget);
    expect(find.text('No data'), findsOneWidget);
    expect(find.text('Try again'), findsOneWidget);
  });

  testWidgets('shows CTA button when buttonText is provided', (tester) async {
    var pressed = false;
    await tester.pumpWidget(buildWidget(
      buttonText: 'Create',
      onButtonPressed: () => pressed = true,
    ));
    await tester.pumpAndSettle();

    expect(find.text('Create'), findsOneWidget);
    await tester.tap(find.text('Create'));
    expect(pressed, true);
  });

  testWidgets('hides CTA button when buttonText is null', (tester) async {
    await tester.pumpWidget(buildWidget());
    await tester.pumpAndSettle();

    expect(find.byType(ElevatedButton), findsNothing);
  });
}
