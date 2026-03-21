import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_client.dart';
import 'package:repa/core/providers/api_provider.dart';
import 'package:repa/core/providers/auth_provider.dart';
import 'package:repa/features/auth/domain/user.dart';

class MockFlutterSecureStorage extends Mock implements FlutterSecureStorage {}

class MockDio extends Mock implements Dio {}

void main() {
  late MockFlutterSecureStorage mockStorage;
  late MockDio mockDio;
  late ProviderContainer container;

  setUp(() {
    mockStorage = MockFlutterSecureStorage();
    mockDio = MockDio();
    container = ProviderContainer(
      overrides: [
        secureStorageProvider.overrideWithValue(mockStorage),
        dioProvider.overrideWithValue(mockDio),
      ],
    );
  });

  tearDown(() {
    container.dispose();
  });

  test('initial state is unknown', () {
    final state = container.read(authProvider);
    expect(state.status, AuthStatus.unknown);
    expect(state.user, isNull);
  });

  test('checkAuth — no token sets unauthenticated', () async {
    when(() => mockStorage.read(key: tokenKey))
        .thenAnswer((_) async => null);

    await container.read(authProvider.notifier).checkAuth();

    final state = container.read(authProvider);
    expect(state.status, AuthStatus.unauthenticated);
  });

  test('checkAuth — valid token sets authenticated with user', () async {
    when(() => mockStorage.read(key: tokenKey))
        .thenAnswer((_) async => 'valid-jwt');
    when(() => mockDio.get('/auth/me')).thenAnswer((_) async => Response(
          data: {
            'data': {
              'id': 'u1',
              'username': 'testuser',
              'avatar_emoji': '\u{1F60E}',
              'birth_year': 2005,
              'created_at': '2026-01-01T00:00:00Z',
            }
          },
          statusCode: 200,
          requestOptions: RequestOptions(path: '/auth/me'),
        ));

    await container.read(authProvider.notifier).checkAuth();

    final state = container.read(authProvider);
    expect(state.status, AuthStatus.authenticated);
    expect(state.user, isNotNull);
    expect(state.user!.username, 'testuser');
    expect(state.needsProfileSetup, false);
  });

  test('checkAuth — valid token but incomplete profile sets needsProfileSetup',
      () async {
    when(() => mockStorage.read(key: tokenKey))
        .thenAnswer((_) async => 'valid-jwt');
    when(() => mockDio.get('/auth/me')).thenAnswer((_) async => Response(
          data: {
            'data': {
              'id': 'u1',
              'username': 'gen_abc123',
              'created_at': '2026-01-01T00:00:00Z',
            }
          },
          statusCode: 200,
          requestOptions: RequestOptions(path: '/auth/me'),
        ));

    await container.read(authProvider.notifier).checkAuth();

    final state = container.read(authProvider);
    expect(state.status, AuthStatus.authenticated);
    expect(state.needsProfileSetup, true);
  });

  test('checkAuth — expired token sets unauthenticated and deletes token',
      () async {
    when(() => mockStorage.read(key: tokenKey))
        .thenAnswer((_) async => 'expired-jwt');
    when(() => mockDio.get('/auth/me')).thenThrow(DioException(
      requestOptions: RequestOptions(path: '/auth/me'),
      response: Response(
        statusCode: 401,
        requestOptions: RequestOptions(path: '/auth/me'),
      ),
    ));
    when(() => mockStorage.delete(key: tokenKey)).thenAnswer((_) async {});

    await container.read(authProvider.notifier).checkAuth();

    final state = container.read(authProvider);
    expect(state.status, AuthStatus.unauthenticated);
    verify(() => mockStorage.delete(key: tokenKey)).called(1);
  });

  test('login stores token and sets authenticated', () async {
    when(() => mockStorage.write(key: tokenKey, value: 'new-jwt'))
        .thenAnswer((_) async {});

    await container.read(authProvider.notifier).login(
          'new-jwt',
          _completeUser,
        );

    final state = container.read(authProvider);
    expect(state.status, AuthStatus.authenticated);
    expect(state.user!.id, 'u1');
    expect(state.needsProfileSetup, false);
  });

  test('login with incomplete profile sets needsProfileSetup', () async {
    when(() => mockStorage.write(key: tokenKey, value: 'new-jwt'))
        .thenAnswer((_) async {});

    await container.read(authProvider.notifier).login(
          'new-jwt',
          _incompleteUser,
        );

    final state = container.read(authProvider);
    expect(state.needsProfileSetup, true);
  });

  test('logout deletes token and sets unauthenticated', () async {
    when(() => mockStorage.write(key: tokenKey, value: any(named: 'value')))
        .thenAnswer((_) async {});
    when(() => mockStorage.delete(key: tokenKey)).thenAnswer((_) async {});

    // First login
    await container.read(authProvider.notifier).login(
          'jwt',
          _completeUser,
        );
    expect(container.read(authProvider).status, AuthStatus.authenticated);

    // Then logout
    await container.read(authProvider.notifier).logout();

    final state = container.read(authProvider);
    expect(state.status, AuthStatus.unauthenticated);
    expect(state.user, isNull);
    verify(() => mockStorage.delete(key: tokenKey)).called(1);
  });

  test('profileCompleted clears needsProfileSetup', () async {
    when(() => mockStorage.write(key: tokenKey, value: any(named: 'value')))
        .thenAnswer((_) async {});

    await container.read(authProvider.notifier).login(
          'jwt',
          _incompleteUser,
        );
    expect(container.read(authProvider).needsProfileSetup, true);

    container
        .read(authProvider.notifier)
        .profileCompleted(_completeUser);

    final state = container.read(authProvider);
    expect(state.needsProfileSetup, false);
    expect(state.user!.avatarEmoji, '\u{1F60E}');
  });
}

const _completeUser = User(
  id: 'u1',
  username: 'testuser',
  avatarEmoji: '\u{1F60E}',
  birthYear: 2005,
  createdAt: '2026-01-01T00:00:00Z',
);

const _incompleteUser = User(
  id: 'u1',
  username: 'gen_abc123',
  createdAt: '2026-01-01T00:00:00Z',
);
