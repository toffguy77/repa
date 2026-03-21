import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_client.dart';
import 'package:repa/features/auth/data/auth_repository.dart';
import 'package:repa/features/auth/domain/user.dart';
import 'package:repa/features/auth/presentation/auth_notifier.dart';
import 'package:repa/core/providers/auth_provider.dart';

class MockAuthRepository extends Mock implements AuthRepository {}

class MockAuthNotifier extends Mock implements AuthNotifier {}

const _testUser = User(
  id: 'u1',
  username: 'testuser',
  avatarEmoji: '\u{1F60E}',
  birthYear: 2005,
  createdAt: '2026-01-01T00:00:00Z',
);

void main() {
  late MockAuthRepository mockRepo;
  late MockAuthNotifier mockAuthNotifier;

  setUp(() {
    mockRepo = MockAuthRepository();
    mockAuthNotifier = MockAuthNotifier();
  });

  group('OtpSendNotifier', () {
    late OtpSendNotifier notifier;

    setUp(() {
      notifier = OtpSendNotifier(mockRepo);
    });

    test('initial state is not loading and no error', () {
      expect(notifier.state.loading, false);
      expect(notifier.state.error, isNull);
    });

    test('send succeeds — returns true and clears loading', () async {
      when(() => mockRepo.sendOtp('+79001234567'))
          .thenAnswer((_) async {});

      final result = await notifier.send('+79001234567');

      expect(result, true);
      expect(notifier.state.loading, false);
      expect(notifier.state.error, isNull);
      verify(() => mockRepo.sendOtp('+79001234567')).called(1);
    });

    test('send fails — returns false and sets error', () async {
      when(() => mockRepo.sendOtp('+79001234567'))
          .thenThrow(AppException('Слишком много запросов', code: 'RATE_LIMIT'));

      final result = await notifier.send('+79001234567');

      expect(result, false);
      expect(notifier.state.loading, false);
      expect(notifier.state.error, 'Слишком много запросов');
    });

    test('send sets loading to true during request', () async {
      bool wasLoading = false;
      when(() => mockRepo.sendOtp(any())).thenAnswer((_) async {
        // Can't check mid-flight in unit test, but we verify state after
      });

      notifier.addListener((state) {
        if (state.loading) wasLoading = true;
      });

      await notifier.send('+79001234567');
      expect(wasLoading, true);
    });
  });

  group('OtpVerifyNotifier', () {
    late OtpVerifyNotifier notifier;

    setUp(() {
      notifier = OtpVerifyNotifier(mockRepo, mockAuthNotifier);
    });

    test('initial state is not loading and no error', () {
      expect(notifier.state.loading, false);
      expect(notifier.state.error, isNull);
    });

    test('verify succeeds — returns true and calls login', () async {
      when(() => mockRepo.verifyOtp('+79001234567', '123456'))
          .thenAnswer((_) async => AuthResult(token: 'jwt', user: _testUser));
      when(() => mockAuthNotifier.login('jwt', _testUser))
          .thenAnswer((_) async {});

      final result = await notifier.verify('+79001234567', '123456');

      expect(result, true);
      expect(notifier.state.loading, false);
      expect(notifier.state.error, isNull);
      verify(() => mockAuthNotifier.login('jwt', _testUser)).called(1);
    });

    test('verify fails — returns false and sets error', () async {
      when(() => mockRepo.verifyOtp('+79001234567', '000000'))
          .thenThrow(AppException('Неверный код', code: 'INVALID_OTP'));

      final result = await notifier.verify('+79001234567', '000000');

      expect(result, false);
      expect(notifier.state.error, 'Неверный код');
    });

    test('reset clears error state', () async {
      when(() => mockRepo.verifyOtp(any(), any()))
          .thenThrow(AppException('Ошибка'));

      await notifier.verify('+79001234567', '000000');
      expect(notifier.state.error, isNotNull);

      notifier.reset();
      expect(notifier.state.error, isNull);
      expect(notifier.state.loading, false);
    });
  });

  group('ProfileSetupNotifier', () {
    late ProfileSetupNotifier notifier;

    setUp(() {
      notifier = ProfileSetupNotifier(mockRepo, mockAuthNotifier);
    });

    test('initial state has no loading, no error, no username check', () {
      expect(notifier.state.loading, false);
      expect(notifier.state.error, isNull);
      expect(notifier.state.usernameAvailable, isNull);
      expect(notifier.state.checkingUsername, false);
    });

    test('checkUsername with short name sets available to null', () async {
      await notifier.checkUsername('ab');
      expect(notifier.state.usernameAvailable, isNull);
      expect(notifier.state.checkingUsername, false);
    });

    test('checkUsername succeeds — sets available', () async {
      when(() => mockRepo.checkUsername('coolname'))
          .thenAnswer((_) async => true);

      await notifier.checkUsername('coolname');

      expect(notifier.state.usernameAvailable, true);
      expect(notifier.state.checkingUsername, false);
    });

    test('checkUsername returns unavailable', () async {
      when(() => mockRepo.checkUsername('taken'))
          .thenAnswer((_) async => false);

      await notifier.checkUsername('taken');

      expect(notifier.state.usernameAvailable, false);
    });

    test('checkUsername on error sets available to null', () async {
      when(() => mockRepo.checkUsername('err'))
          .thenThrow(AppException('Ошибка'));

      await notifier.checkUsername('err');

      expect(notifier.state.usernameAvailable, isNull);
    });

    test('submit succeeds — returns true and calls profileCompleted', () async {
      when(() => mockRepo.updateProfile(
            username: 'newname',
            avatarEmoji: '\u{1F60E}',
            birthYear: 2005,
          )).thenAnswer((_) async => _testUser);
      when(() => mockAuthNotifier.profileCompleted(_testUser)).thenReturn(null);

      final result = await notifier.submit(
        username: 'newname',
        avatarEmoji: '\u{1F60E}',
        birthYear: 2005,
      );

      expect(result, true);
      expect(notifier.state.loading, false);
      expect(notifier.state.error, isNull);
      verify(() => mockAuthNotifier.profileCompleted(_testUser)).called(1);
    });

    test('submit fails — returns false and sets error', () async {
      when(() => mockRepo.updateProfile(
            username: any(named: 'username'),
            avatarEmoji: any(named: 'avatarEmoji'),
            birthYear: any(named: 'birthYear'),
          )).thenThrow(AppException('Имя занято', code: 'USERNAME_TAKEN'));

      final result = await notifier.submit(
        username: 'taken',
        avatarEmoji: '\u{1F60E}',
        birthYear: 2005,
      );

      expect(result, false);
      expect(notifier.state.error, 'Имя занято');
    });
  });
}
