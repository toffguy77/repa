import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../api/api_client.dart';
import '../api/api_service.dart';
import 'auth_provider.dart';

final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final dioProvider = Provider<Dio>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return createDio(storage, () {
    ref.read(authProvider.notifier).logout();
  });
});

final apiServiceProvider = Provider<ApiService>((ref) {
  final dio = ref.watch(dioProvider);
  return ApiService(dio);
});
