import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/features/home/home_screen.dart';

void main() {
  testWidgets('renders welcome message and logout button', (tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(home: HomeScreen()),
      ),
    );

    expect(find.text('Добро пожаловать в Репу!'), findsOneWidget);
    expect(find.text('Выйти'), findsOneWidget);
  });

  testWidgets('shows stub message about T07', (tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(home: HomeScreen()),
      ),
    );

    expect(find.text('Экран групп появится в T07'), findsOneWidget);
  });
}
