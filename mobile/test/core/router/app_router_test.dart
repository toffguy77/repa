import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_client.dart';
import 'package:repa/core/providers/api_provider.dart';
import 'package:repa/core/providers/auth_provider.dart';
import 'package:repa/core/router/app_router.dart';

class MockFlutterSecureStorage extends Mock implements FlutterSecureStorage {}

class MockDio extends Mock implements Dio {}

void main() {
  testWidgets('unauthenticated user sees phone screen', (tester) async {
    final mockStorage = MockFlutterSecureStorage();
    final mockDio = MockDio();
    when(() => mockStorage.read(key: tokenKey))
        .thenAnswer((_) async => null);

    final container = ProviderContainer(
      overrides: [
        secureStorageProvider.overrideWithValue(mockStorage),
        dioProvider.overrideWithValue(mockDio),
      ],
    );

    await container.read(authProvider.notifier).checkAuth();

    final router = container.read(routerProvider);
    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: MaterialApp.router(routerConfig: router),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Вход в Репу'), findsOneWidget);
    container.dispose();
  });
}
