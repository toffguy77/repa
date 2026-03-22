import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/voting/presentation/widgets/question_card.dart';

void main() {
  Widget buildWidget({
    String text = 'Кто первым убежит?',
    String category = 'FUNNY',
  }) {
    return MaterialApp(
      home: Scaffold(
        body: QuestionCard(text: text, category: category),
      ),
    );
  }

  testWidgets('shows question text', (tester) async {
    await tester.pumpWidget(buildWidget());
    await tester.pumpAndSettle();

    expect(find.text('Кто первым убежит?'), findsOneWidget);
  });

  testWidgets('shows category emoji for FUNNY', (tester) async {
    await tester.pumpWidget(buildWidget(category: 'FUNNY'));
    await tester.pumpAndSettle();

    expect(find.text('\u{1F602}'), findsOneWidget);
  });

  testWidgets('shows category emoji for HOT', (tester) async {
    await tester.pumpWidget(buildWidget(category: 'HOT'));
    await tester.pumpAndSettle();

    expect(find.text('\u{1F525}'), findsOneWidget);
  });

  testWidgets('shows fallback emoji for unknown category', (tester) async {
    await tester.pumpWidget(buildWidget(category: 'UNKNOWN'));
    await tester.pumpAndSettle();

    expect(find.text('\u{2753}'), findsOneWidget);
  });
}
