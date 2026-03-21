import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

const String apiBaseUrl = 'http://localhost:3000/api/v1';
const String tokenKey = 'auth_token';

class AppException implements Exception {
  final String message;
  final String? code;

  AppException(this.message, {this.code});

  @override
  String toString() => message;
}

Dio createDio(FlutterSecureStorage storage, void Function() onUnauthorized) {
  final dio = Dio(BaseOptions(
    baseUrl: apiBaseUrl,
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 10),
    headers: {'Content-Type': 'application/json'},
  ));

  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) async {
      final token = await storage.read(key: tokenKey);
      if (token != null) {
        options.headers['Authorization'] = 'Bearer $token';
      }
      handler.next(options);
    },
    onError: (error, handler) async {
      if (error.response?.statusCode == 401) {
        await storage.delete(key: tokenKey);
        onUnauthorized();
      }
      handler.next(error);
    },
  ));

  return dio;
}

AppException parseError(DioException e) {
  final data = e.response?.data;
  if (data is Map<String, dynamic> && data.containsKey('error')) {
    final err = data['error'] as Map<String, dynamic>;
    return AppException(
      err['message'] as String? ?? 'Unknown error',
      code: err['code'] as String?,
    );
  }
  if (e.type == DioExceptionType.connectionTimeout ||
      e.type == DioExceptionType.receiveTimeout) {
    return AppException('Нет соединения с сервером');
  }
  return AppException('Что-то пошло не так');
}
