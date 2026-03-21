import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:repa/core/api/api_service.dart';

class MockDio extends Mock implements Dio {}

class FakeRequestOptions extends Fake implements RequestOptions {}

void main() {
  late MockDio mockDio;
  late ApiService service;

  setUpAll(() {
    registerFallbackValue(FakeRequestOptions());
  });

  setUp(() {
    mockDio = MockDio();
    service = ApiService(mockDio);
  });

  test('otpSend posts to correct path', () async {
    when(() => mockDio.post('/auth/otp/send', data: any(named: 'data')))
        .thenAnswer((_) async => Response(
              data: {'data': {'sent': true}},
              statusCode: 200,
              requestOptions: RequestOptions(path: '/auth/otp/send'),
            ));

    final result = await service.otpSend({'phone': '+79001234567'});
    expect(result['data']['sent'], true);
  });

  test('otpVerify posts to correct path', () async {
    when(() => mockDio.post('/auth/otp/verify', data: any(named: 'data')))
        .thenAnswer((_) async => Response(
              data: {
                'data': {
                  'token': 'jwt',
                  'user': {'id': 'u1', 'username': 'test', 'created_at': '2026-01-01'}
                }
              },
              statusCode: 200,
              requestOptions: RequestOptions(path: '/auth/otp/verify'),
            ));

    final result = await service.otpVerify({'phone': '+79001234567', 'code': '123456'});
    expect(result['data']['token'], 'jwt');
  });

  test('getMe gets from correct path', () async {
    when(() => mockDio.get('/auth/me')).thenAnswer((_) async => Response(
          data: {
            'data': {'id': 'u1', 'username': 'test', 'created_at': '2026-01-01'}
          },
          statusCode: 200,
          requestOptions: RequestOptions(path: '/auth/me'),
        ));

    final result = await service.getMe();
    expect(result['data']['id'], 'u1');
  });

  test('checkUsername gets with query param', () async {
    when(() => mockDio.get('/auth/username/check',
            queryParameters: any(named: 'queryParameters')))
        .thenAnswer((_) async => Response(
              data: {'data': {'available': true}},
              statusCode: 200,
              requestOptions: RequestOptions(path: '/auth/username/check'),
            ));

    final result = await service.checkUsername('testname');
    expect(result['data']['available'], true);
  });

  test('updateProfile puts to correct path', () async {
    when(() => mockDio.put('/auth/profile', data: any(named: 'data')))
        .thenAnswer((_) async => Response(
              data: {
                'data': {'id': 'u1', 'username': 'newname', 'created_at': '2026-01-01'}
              },
              statusCode: 200,
              requestOptions: RequestOptions(path: '/auth/profile'),
            ));

    final result = await service.updateProfile({'username': 'newname'});
    expect(result['data']['username'], 'newname');
  });

  test('appVersion gets from correct path', () async {
    when(() => mockDio.get('/auth/version')).thenAnswer((_) async => Response(
          data: {
            'data': {
              'min_version': '1.0.0',
              'latest_version': '1.0.0',
              'force_update': false,
            }
          },
          statusCode: 200,
          requestOptions: RequestOptions(path: '/auth/version'),
        ));

    final result = await service.appVersion();
    expect(result['data']['force_update'], false);
  });

  test('appleAuth posts to correct path', () async {
    when(() => mockDio.post('/auth/apple', data: any(named: 'data')))
        .thenAnswer((_) async => Response(
              data: {
                'data': {
                  'token': 'jwt',
                  'user': {'id': 'u1', 'username': 'test', 'created_at': '2026-01-01'}
                }
              },
              statusCode: 200,
              requestOptions: RequestOptions(path: '/auth/apple'),
            ));

    final result = await service.appleAuth({'id_token': 'tok'});
    expect(result['data']['token'], 'jwt');
  });

  test('googleAuth posts to correct path', () async {
    when(() => mockDio.post('/auth/google', data: any(named: 'data')))
        .thenAnswer((_) async => Response(
              data: {
                'data': {
                  'token': 'jwt',
                  'user': {'id': 'u1', 'username': 'test', 'created_at': '2026-01-01'}
                }
              },
              statusCode: 200,
              requestOptions: RequestOptions(path: '/auth/google'),
            ));

    final result = await service.googleAuth({'id_token': 'tok'});
    expect(result['data']['token'], 'jwt');
  });
}
