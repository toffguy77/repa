import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:repa/core/api/api_client.dart';

void main() {
  group('parseError', () {
    test('extracts message and code from API error response', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/test'),
        response: Response(
          statusCode: 400,
          data: {
            'error': {'code': 'VALIDATION', 'message': 'Invalid input'}
          },
          requestOptions: RequestOptions(path: '/test'),
        ),
      );

      final result = parseError(error);
      expect(result.message, 'Invalid input');
      expect(result.code, 'VALIDATION');
    });

    test('returns generic message for non-API error', () {
      final error = DioException(
        requestOptions: RequestOptions(path: '/test'),
        response: Response(
          statusCode: 500,
          data: 'Internal Server Error',
          requestOptions: RequestOptions(path: '/test'),
        ),
      );

      final result = parseError(error);
      expect(result.message, 'Что-то пошло не так');
    });

    test('returns connection error for timeout', () {
      final error = DioException(
        type: DioExceptionType.connectionTimeout,
        requestOptions: RequestOptions(path: '/test'),
      );

      final result = parseError(error);
      expect(result.message, 'Нет соединения с сервером');
    });

    test('returns connection error for receive timeout', () {
      final error = DioException(
        type: DioExceptionType.receiveTimeout,
        requestOptions: RequestOptions(path: '/test'),
      );

      final result = parseError(error);
      expect(result.message, 'Нет соединения с сервером');
    });
  });

  group('AppException', () {
    test('toString returns message', () {
      final e = AppException('Test error', code: 'TEST');
      expect(e.toString(), 'Test error');
    });
  });
}
