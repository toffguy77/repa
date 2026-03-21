import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/features/auth/data/auth_repository.dart';
import 'package:repa/features/auth/presentation/auth_notifier.dart';
import 'package:repa/features/auth/presentation/phone_screen.dart';

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

  testWidgets('renders title and phone input', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const PhoneScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    expect(find.text('Вход в Репу'), findsOneWidget);
    expect(find.byType(TextField), findsOneWidget);
    expect(find.text('Получить код'), findsOneWidget);
  });

  testWidgets('button is disabled when phone is empty', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const PhoneScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
    expect(button.onPressed, isNull);
  });

  testWidgets('button becomes enabled after entering valid phone', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const PhoneScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    await tester.enterText(find.byType(TextField), '9001234567');
    await tester.pump();

    final button = tester.widget<ElevatedButton>(find.byType(ElevatedButton));
    expect(button.onPressed, isNotNull);
  });

  testWidgets('shows Apple and Google sign-in buttons', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const PhoneScreen(),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    expect(find.text('Войти через Apple'), findsOneWidget);
    expect(find.text('Войти через Google'), findsOneWidget);
  });
}
