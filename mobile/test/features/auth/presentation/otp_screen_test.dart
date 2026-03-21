import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/features/auth/data/auth_repository.dart';
import 'package:repa/features/auth/presentation/auth_notifier.dart';
import 'package:repa/features/auth/presentation/otp_screen.dart';

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

  testWidgets('renders title and phone number', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const OtpScreen(phone: '+79001234567'),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    expect(find.text('Введи код'), findsOneWidget);
    expect(find.text('Отправили SMS на +79001234567'), findsOneWidget);
  });

  testWidgets('shows countdown timer', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const OtpScreen(phone: '+79001234567'),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    // Timer starts at 05:00
    expect(find.textContaining('05:00'), findsOneWidget);
  });

  testWidgets('timer decrements after 1 second', (tester) async {
    await tester.pumpWidget(_wrapWidget(
      const OtpScreen(phone: '+79001234567'),
      overrides: [
        authRepositoryProvider.overrideWithValue(mockRepo),
      ],
    ));

    await tester.pump(const Duration(seconds: 1));

    expect(find.textContaining('04:59'), findsOneWidget);
  });
}
