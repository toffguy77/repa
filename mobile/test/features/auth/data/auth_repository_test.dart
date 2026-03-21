import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_client.dart';
import 'package:repa/core/api/api_service.dart';
import 'package:repa/features/auth/data/auth_repository.dart';

class MockApiService extends Mock implements ApiService {}

void main() {
  late MockApiService mockApi;
  late AuthRepository repo;

  setUp(() {
    mockApi = MockApiService();
    repo = AuthRepository(mockApi);
  });

  group('sendOtp', () {
    test('calls api with phone number', () async {
      when(() => mockApi.otpSend({'phone': '+79001234567'}))
          .thenAnswer((_) async => {'data': {'sent': true}});

      await repo.sendOtp('+79001234567');

      verify(() => mockApi.otpSend({'phone': '+79001234567'})).called(1);
    });

    test('throws AppException on DioException', () async {
      when(() => mockApi.otpSend(any())).thenThrow(DioException(
        requestOptions: RequestOptions(path: '/auth/otp/send'),
        response: Response(
          statusCode: 429,
          data: {
            'error': {'code': 'RATE_LIMIT', 'message': 'Too many requests'}
          },
          requestOptions: RequestOptions(path: '/auth/otp/send'),
        ),
      ));

      expect(
        () => repo.sendOtp('+79001234567'),
        throwsA(isA<AppException>()),
      );
    });
  });

  group('verifyOtp', () {
    test('returns AuthResult with token and user', () async {
      when(() => mockApi.otpVerify({'phone': '+79001234567', 'code': '123456'}))
          .thenAnswer((_) async => {
                'data': {
                  'token': 'jwt-token',
                  'user': {
                    'id': 'u1',
                    'username': 'testuser',
                    'avatar_emoji': '\u{1F60E}',
                    'birth_year': 2005,
                    'created_at': '2026-01-01T00:00:00Z',
                  }
                }
              });

      final result = await repo.verifyOtp('+79001234567', '123456');

      expect(result.token, 'jwt-token');
      expect(result.user.id, 'u1');
      expect(result.user.username, 'testuser');
    });

    test('throws AppException on invalid code', () async {
      when(() => mockApi.otpVerify(any())).thenThrow(DioException(
        requestOptions: RequestOptions(path: '/auth/otp/verify'),
        response: Response(
          statusCode: 401,
          data: {
            'error': {'code': 'INVALID_OTP', 'message': 'Invalid code'}
          },
          requestOptions: RequestOptions(path: '/auth/otp/verify'),
        ),
      ));

      expect(
        () => repo.verifyOtp('+79001234567', '000000'),
        throwsA(isA<AppException>()),
      );
    });
  });

  group('checkUsername', () {
    test('returns true for available username', () async {
      when(() => mockApi.checkUsername('coolname'))
          .thenAnswer((_) async => {
                'data': {'available': true}
              });

      final result = await repo.checkUsername('coolname');
      expect(result, true);
    });

    test('returns false for taken username', () async {
      when(() => mockApi.checkUsername('taken'))
          .thenAnswer((_) async => {
                'data': {'available': false}
              });

      final result = await repo.checkUsername('taken');
      expect(result, false);
    });
  });

  group('updateProfile', () {
    test('sends profile data and returns updated user', () async {
      when(() => mockApi.updateProfile(any())).thenAnswer((_) async => {
            'data': {
              'id': 'u1',
              'username': 'newname',
              'avatar_emoji': '\u{1F60E}',
              'birth_year': 2005,
              'created_at': '2026-01-01T00:00:00Z',
            }
          });

      final user = await repo.updateProfile(
        username: 'newname',
        avatarEmoji: '\u{1F60E}',
        birthYear: 2005,
      );

      expect(user.username, 'newname');
      verify(() => mockApi.updateProfile({
            'username': 'newname',
            'avatar_emoji': '\u{1F60E}',
            'birth_year': 2005,
          })).called(1);
    });

    test('only sends non-null fields', () async {
      when(() => mockApi.updateProfile(any())).thenAnswer((_) async => {
            'data': {
              'id': 'u1',
              'username': 'existing',
              'avatar_emoji': '\u{1F60E}',
              'created_at': '2026-01-01T00:00:00Z',
            }
          });

      await repo.updateProfile(avatarEmoji: '\u{1F60E}');

      verify(() => mockApi.updateProfile({'avatar_emoji': '\u{1F60E}'})).called(1);
    });
  });
}
