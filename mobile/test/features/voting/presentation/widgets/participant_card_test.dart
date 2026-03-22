import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/voting/presentation/widgets/participant_card.dart';

void main() {
  Widget buildWidget({
    String username = 'alice',
    String? avatarEmoji = '😎',
    bool selected = false,
    bool disabled = false,
    VoidCallback? onTap,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: ParticipantCard(
          username: username,
          avatarEmoji: avatarEmoji,
          selected: selected,
          disabled: disabled,
          onTap: onTap ?? () {},
        ),
      ),
    );
  }

  testWidgets('shows username', (tester) async {
    await tester.pumpWidget(buildWidget());
    expect(find.text('alice'), findsOneWidget);
  });

  testWidgets('shows avatar emoji', (tester) async {
    await tester.pumpWidget(buildWidget(avatarEmoji: '🐱'));
    expect(find.text('🐱'), findsOneWidget);
  });

  testWidgets('calls onTap when tapped', (tester) async {
    var tapped = false;
    await tester.pumpWidget(buildWidget(onTap: () => tapped = true));

    await tester.tap(find.byType(ParticipantCard));
    expect(tapped, true);
  });

  testWidgets('does not call onTap when disabled', (tester) async {
    var tapped = false;
    await tester.pumpWidget(
        buildWidget(disabled: true, onTap: () => tapped = true));

    await tester.tap(find.byType(ParticipantCard));
    expect(tapped, false);
  });

  testWidgets('shows checkmark when selected', (tester) async {
    await tester.pumpWidget(buildWidget(selected: true));
    expect(find.byIcon(Icons.check), findsOneWidget);
  });

  testWidgets('no checkmark when not selected', (tester) async {
    await tester.pumpWidget(buildWidget(selected: false));
    expect(find.byIcon(Icons.check), findsNothing);
  });
}
