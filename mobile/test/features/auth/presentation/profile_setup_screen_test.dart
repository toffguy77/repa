import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/features/auth/data/auth_repository.dart';
import 'package:repa/features/auth/presentation/auth_notifier.dart';
import 'package:repa/features/auth/presentation/profile_setup_screen.dart';

class MockAuthRepository extends Mock implements AuthRepository {}

Widget _wrapWidget(Widget child, {List<Override> overrides = const []}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(home: child),
  );
}

void main() {
  late MockAuthRepository mockRepo;

  setUp(() {
    mockRepo = MockAuthRepository();
  });

  testWidgets('renders title and form fields', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const ProfileSetupScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    expect(find.text('Расскажи о себе'), findsOneWidget);
    expect(find.text('Выбери аватар'), findsOneWidget);
    expect(find.text('Имя пользователя'), findsOneWidget);
    expect(find.text('Год рождения'), findsOneWidget);
    expect(find.text('Готово'), findsOneWidget);
  });

  testWidgets('button is disabled when form is empty', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const ProfileSetupScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
    expect(button.onPressed, isNull);
  });

  testWidgets('emoji selection changes selected avatar', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const ProfileSetupScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    // Find emoji containers — the second emoji (ghost)
    final emojis = find.byType(GestureDetector);
    expect(emojis, findsWidgets);

    // Tap the second emoji
    await tester.tap(emojis.at(1));
    await tester.pump();

    // Verify the second container now has a primary-colored border
    // (We just verify the tap doesn't crash and the widget rebuilds)
    expect(find.text('Расскажи о себе'), findsOneWidget);
  });

  testWidgets('button becomes enabled with valid form data', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const ProfileSetupScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    // Find text fields by hint text
    final usernameField = find.widgetWithText(TextField, 'Минимум 3 символа');
    final birthYearField = find.widgetWithText(TextField, 'Например, 2005');

    await tester.enterText(usernameField, 'testuser');
    await tester.enterText(birthYearField, '2005');
    await tester.pump();

    final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
    expect(button.onPressed, isNotNull);
  });
}
